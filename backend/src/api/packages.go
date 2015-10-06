package api

import (
	"time"

	"gopkg.in/mgutz/dat.v1"
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

// Package represents a CoreRoller application's package.
type Package struct {
	ID            string         `db:"id" json:"id"`
	Type          int            `db:"type" json:"type"`
	Version       string         `db:"version" json:"version"`
	URL           string         `db:"url" json:"url"`
	Filename      dat.NullString `db:"filename" json:"filename"`
	Description   dat.NullString `db:"description" json:"description"`
	Size          dat.NullString `db:"size" json:"size"`
	Hash          dat.NullString `db:"hash" json:"hash"`
	CreatedTs     time.Time      `db:"created_ts" json:"-"`
	ApplicationID string         `db:"application_id" json:"application_id"`
}

// AddPackage registers the provided package.
func (api *API) AddPackage(pkg *Package) (*Package, error) {
	err := api.dbR.
		InsertInto("package").
		Whitelist("type", "filename", "description", "size", "hash", "url", "version", "application_id").
		Record(pkg).
		Returning("*").
		QueryStruct(pkg)

	return pkg, err
}

// UpdatePackage updates an existing package using the content of the package
// provided.
func (api *API) UpdatePackage(pkg *Package) error {
	result, err := api.dbR.
		Update("package").
		SetWhitelist(pkg, "type", "filename", "description", "size", "hash", "url", "version").
		Where("id = $1", pkg.ID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
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

// GetPackageJSON returns the package identified by the id provided in JSON
// format.
func (api *API) GetPackageJSON(pkgID string) ([]byte, error) {
	return api.packagesQuery().
		Where("id = $1", pkgID).
		QueryJSON()
}

// GetPackages returns all packages associated to the application provided.
func (api *API) GetPackages(appID string) ([]*Package, error) {
	var pkgs []*Package

	err := api.packagesQuery().
		Where("application_id = $1", appID).
		QueryStructs(&pkgs)

	return pkgs, err
}

// GetPackagesJSON returns all packages associated to the application provided
// in JSON format.
func (api *API) GetPackagesJSON(appID string, page, perPage uint64) ([]byte, error) {
	page, perPage = validatePaginationParams(page, perPage)

	return api.packagesQuery().
		Where("application_id = $1", appID).
		Paginate(page, perPage).
		QueryJSON()
}

// packagesQuery returns a SelectDocBuilder prepared to return all packages.
// This query is meant to be extended later in the methods using it to filter
// by a specific package id, all packages that belong to a given application,
// specify how to query the rows or their destination.
func (api *API) packagesQuery() *dat.SelectDocBuilder {
	return api.dbR.
		SelectDoc("*").
		From("package").
		OrderBy("created_ts DESC")
}
