package api

import (
	"errors"
	"time"

	"gopkg.in/mgutz/dat.v1"
)

var (
	// ErrInvalidPackage error indicates that a package doesn't belong to the
	// application it was supposed to belong to.
	ErrInvalidPackage = errors.New("coreroller: invalid package")

	// ErrBlacklistedChannel error indicates an attempt of creating/updating a
	// channel using a package that has blacklisted the channel.
	ErrBlacklistedChannel = errors.New("coreroller: blacklisted channel")
)

// Channel represents a CoreRoller application's channel.
type Channel struct {
	ID            string         `db:"id" json:"id"`
	Name          string         `db:"name" json:"name"`
	Color         string         `db:"color" json:"color"`
	CreatedTs     time.Time      `db:"created_ts" json:"created_ts"`
	ApplicationID string         `db:"application_id" json:"application_id"`
	PackageID     dat.NullString `db:"package_id" json:"package_id"`
	Package       *Package       `db:"package" json:"package"`
}

// AddChannel registers the provided channel.
func (api *API) AddChannel(channel *Channel) (*Channel, error) {
	if channel.PackageID.String != "" {
		if _, err := api.validatePackage(channel.PackageID.String, channel.ID, channel.ApplicationID); err != nil {
			return nil, err
		}
	}

	err := api.dbR.
		InsertInto("channel").
		Whitelist("name", "color", "application_id", "package_id").
		Record(channel).
		Returning("*").
		QueryStruct(channel)

	return channel, err
}

// UpdateChannel updates an existing channel using the content of the channel
// provided.
func (api *API) UpdateChannel(channel *Channel) error {
	channelBeforeUpdate, err := api.GetChannel(channel.ID)
	if err != nil {
		return err
	}

	var pkg *Package
	if channel.PackageID.String != "" {
		if pkg, err = api.validatePackage(channel.PackageID.String, channel.ID, channelBeforeUpdate.ApplicationID); err != nil {
			return err
		}
	}

	result, err := api.dbR.
		Update("channel").
		SetWhitelist(channel, "name", "color", "package_id").
		Where("id = $1", channel.ID).
		Exec()

	if err != nil {
		return err
	}

	if result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	if channelBeforeUpdate.PackageID.String != channel.PackageID.String && pkg != nil {
		_ = api.newChannelActivityEntry(activityChannelPackageUpdated, activityInfo, pkg.Version, pkg.ApplicationID, channel.ID)
	}

	return nil
}

// DeleteChannel removes the channel identified by the id provided.
func (api *API) DeleteChannel(channelID string) error {
	result, err := api.dbR.
		DeleteFrom("channel").
		Where("id = $1", channelID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// GetChannel returns the channel identified by the id provided.
func (api *API) GetChannel(channelID string) (*Channel, error) {
	var channel Channel

	err := api.channelsQuery().
		Where("id = $1", channelID).
		QueryStruct(&channel)

	if err != nil {
		return nil, err
	}

	return &channel, nil
}

// GetChannels returns all channels associated to the application provided.
func (api *API) GetChannels(appID string, page, perPage uint64) ([]*Channel, error) {
	page, perPage = validatePaginationParams(page, perPage)

	var channels []*Channel

	err := api.channelsQuery().
		Where("application_id = $1", appID).
		Paginate(page, perPage).
		QueryStructs(&channels)

	return channels, err
}

// validatePackage checks if a package belongs to the application provided and
// that the channel is not in the package's channels blacklist. It returns the
// package if everything is ok.
func (api *API) validatePackage(packageID, channelID, appID string) (*Package, error) {
	pkg, err := api.GetPackage(packageID)
	if err == nil {
		if pkg.ApplicationID != appID {
			return nil, ErrInvalidPackage
		}

		for _, blacklistedChannelID := range pkg.ChannelsBlacklist {
			if channelID == blacklistedChannelID {
				return nil, ErrBlacklistedChannel
			}
		}
	}

	return pkg, err
}

// channelsQuery returns a SelectDocBuilder prepared to return all channels.
// This query is meant to be extended later in the methods using it to filter
// by a specific channel id, all channels that belong to a given application,
// specify how to query the rows or their destination.
func (api *API) channelsQuery() *dat.SelectDocBuilder {
	return api.dbR.
		SelectDoc("*").
		One("package", api.packagesQuery().Where("package.id = channel.package_id")).
		From("channel").
		OrderBy("name ASC")
}
