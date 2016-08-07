package api

import (
	"errors"
	"time"

	"gopkg.in/fatih/set.v0"
	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/mgutz/dat.v1/sqlx-runner"
)

const (
	// PkgTypeCoreos indicates that the package is a CoreOS update package
	PkgTypeCoreos int = 1 + iota

	// PkgTypeDocker indicates that the package is a Docker container
	PkgTypeDocker

	// PkgTypeRocket indicates that the package is a Rocket container
	PkgTypeRocket

	// PkgTypeOther is the generic package type.
	PkgTypeOther
)

var (
	// ErrBlacklistingChannel error indicates that the channel the package is
	// trying to blacklist is already pointing to the package.
	ErrBlacklistingChannel = errors.New("coreroller: channel trying to blacklist is already pointing to the package")
)

// Package represents a CoreRoller application's package.
type Package struct {
	ID                string         `db:"id" json:"id"`
	Type              int            `db:"type" json:"type"`
	Version           string         `db:"version" json:"version"`
	URL               string         `db:"url" json:"url"`
	Filename          dat.NullString `db:"filename" json:"filename"`
	Description       dat.NullString `db:"description" json:"description"`
	Size              dat.NullString `db:"size" json:"size"`
	Hash              dat.NullString `db:"hash" json:"hash"`
	CreatedTs         time.Time      `db:"created_ts" json:"created_ts"`
	ChannelsBlacklist []string       `db:"channels_blacklist" json:"channels_blacklist"`
	ApplicationID     string         `db:"application_id" json:"application_id"`
	CoreosAction      *CoreosAction  `db:"coreos_action" json:"coreos_action"`
}

// AddPackage registers the provided package.
func (api *API) AddPackage(pkg *Package) (*Package, error) {
	if !isValidSemver(pkg.Version) {
		return nil, ErrInvalidSemver
	}

	tx, err := api.dbR.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.AutoRollback()
	}()

	err = tx.InsertInto("package").
		Whitelist("type", "filename", "description", "size", "hash", "url", "version", "application_id").
		Record(pkg).
		Returning("*").
		QueryStruct(pkg)

	if err != nil {
		return nil, err
	}

	if len(pkg.ChannelsBlacklist) > 0 {
		for _, channelID := range pkg.ChannelsBlacklist {
			_, err := tx.InsertInto("package_channel_blacklist").
				Pair("package_id", pkg.ID).
				Pair("channel_id", channelID).
				Exec()

			if err != nil {
				return nil, err
			}
		}
	}

	if pkg.Type == PkgTypeCoreos && pkg.CoreosAction != nil {
		err = tx.InsertInto("coreos_action").
			Columns("package_id", "sha256").
			Values(pkg.ID, pkg.CoreosAction.Sha256).
			Returning("*").
			QueryStruct(pkg.CoreosAction)

		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return pkg, nil
}

// UpdatePackage updates an existing package using the content of the package
// provided.
func (api *API) UpdatePackage(pkg *Package) error {
	if !isValidSemver(pkg.Version) {
		return ErrInvalidSemver
	}

	tx, err := api.dbR.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.AutoRollback()
	}()

	result, err := tx.
		Update("package").
		SetWhitelist(pkg, "type", "filename", "description", "size", "hash", "url", "version").
		Where("id = $1", pkg.ID).
		Exec()

	if err != nil {
		return err
	} else if result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	if err := api.updatePackageBlacklistedChannels(tx, pkg); err != nil {
		return err
	}

	if pkg.Type == PkgTypeCoreos && pkg.CoreosAction != nil {
		err = tx.Upsert("coreos_action").
			Columns("package_id", "sha256").
			Values(pkg.ID, pkg.CoreosAction.Sha256).
			Where("package_id = $1", pkg.ID).
			Returning("*").
			QueryStruct(pkg.CoreosAction)

		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// DeletePackage removes the package identified by the id provided.
func (api *API) DeletePackage(pkgID string) error {
	result, err := api.dbR.
		DeleteFrom("package").
		Where("id = $1", pkgID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// GetPackage returns the package identified by the id provided.
func (api *API) GetPackage(pkgID string) (*Package, error) {
	var pkg Package

	err := api.packagesQuery().
		Where("id = $1", pkgID).
		QueryStruct(&pkg)

	if err != nil {
		return nil, err
	}

	return &pkg, nil
}

// GetPackageByVersion returns the package identified by the application id and
// version provided.
func (api *API) GetPackageByVersion(appID, version string) (*Package, error) {
	var pkg Package

	err := api.packagesQuery().
		Where("application_id = $1", appID).
		Where("version = $1", version).
		QueryStruct(&pkg)

	if err != nil {
		return nil, err
	}

	return &pkg, nil
}

// GetPackages returns all packages associated to the application provided.
func (api *API) GetPackages(appID string, page, perPage uint64) ([]*Package, error) {
	page, perPage = validatePaginationParams(page, perPage)

	var pkgs []*Package
	err := api.packagesQuery().
		Where("application_id = $1", appID).
		Paginate(page, perPage).
		QueryStructs(&pkgs)

	return pkgs, err
}

// packagesQuery returns a SelectDocBuilder prepared to return all packages.
// This query is meant to be extended later in the methods using it to filter
// by a specific package id, all packages that belong to a given application,
// specify how to query the rows or their destination.
func (api *API) packagesQuery() *dat.SelectDocBuilder {
	return api.dbR.
		SelectDoc(`
			package.*, 
			array_agg(pcb.channel_id) FILTER (WHERE pcb.channel_id IS NOT NULL) as channels_blacklist
		`).
		One("coreos_action", "SELECT * FROM coreos_action WHERE package_id = package.id").
		From("package LEFT JOIN package_channel_blacklist pcb ON package.id = pcb.package_id").
		GroupBy("package.id").
		OrderBy("regexp_matches(version, '(\\d+)\\.(\\d+)\\.(\\d+)')::int[] DESC")
}

// updatePackageBlacklistedChannels adds or removes as needed channels to the
// package's channels blacklist based on the new entries provided in the updated
// package entry.
//
// This method is part of the transaction that updates a package and when it's
// called, the package has already been updated except for the channels
// blacklist, that may happen here if needed.
func (api *API) updatePackageBlacklistedChannels(tx *runner.Tx, pkg *Package) error {
	pkgUpdated, err := api.GetPackage(pkg.ID)
	if err != nil {
		return err
	}

	newChannelsBlacklist := set.NewNonTS()
	for _, channelID := range pkg.ChannelsBlacklist {
		newChannelsBlacklist.Add(channelID)
	}

	oldChannelsBlacklist := set.NewNonTS()
	for _, channelID := range pkgUpdated.ChannelsBlacklist {
		oldChannelsBlacklist.Add(channelID)
	}

	for _, channelID := range set.Difference(newChannelsBlacklist, oldChannelsBlacklist).List() {
		channel, err := api.GetChannel(channelID.(string))
		if err != nil {
			return err
		}
		if channel.PackageID.String == pkg.ID {
			return ErrBlacklistingChannel
		}
		_, err = tx.InsertInto("package_channel_blacklist").
			Pair("package_id", pkg.ID).
			Pair("channel_id", channelID.(string)).
			Exec()

		if err != nil {
			return err
		}
	}

	for _, channelID := range set.Difference(oldChannelsBlacklist, newChannelsBlacklist).List() {
		_, err := tx.DeleteFrom("package_channel_blacklist").
			Where("package_id = $1", pkg.ID).
			Where("channel_id = $1", channelID.(string)).
			Exec()

		if err != nil {
			return err
		}
	}

	return nil
}
