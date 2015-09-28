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
)

// Channel represents a CoreRoller application's channel.
type Channel struct {
	ID            string         `db:"id" json:"id"`
	Name          string         `db:"name" json:"name"`
	Color         string         `db:"color" json:"color"`
	CreatedTs     time.Time      `db:"created_ts" json:"-"`
	ApplicationID string         `db:"application_id" json:"application_id"`
	PackageID     dat.NullString `db:"package_id" json:"package_id"`
	Package       *Package       `db:"package" json:"package"`
}

// AddChannel registers the provided channel.
func (api *API) AddChannel(channel *Channel) (*Channel, error) {
	if channel.PackageID.String != "" {
		if _, err := api.validatePackage(channel.PackageID.String, channel.ApplicationID); err != nil {
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
		if pkg, err = api.validatePackage(channel.PackageID.String, channelBeforeUpdate.ApplicationID); err != nil {
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

// GetChannelJSON returns the channel identified by the id provided in JSON
// format.
func (api *API) GetChannelJSON(channelID string) ([]byte, error) {
	return api.channelsQuery().
		Where("id = $1", channelID).
		QueryJSON()
}

// GetChannels returns all channels associated to the application provided.
func (api *API) GetChannels(appID string) ([]*Channel, error) {
	var channels []*Channel

	err := api.channelsQuery().
		Where("application_id = $1", appID).
		QueryStructs(&channels)

	return channels, err
}

// GetChannelsJSON returns all channels associated to the application provided
// in JSON format.
func (api *API) GetChannelsJSON(appID string, page, perPage uint64) ([]byte, error) {
	page, perPage = validatePaginationParams(page, perPage)

	return api.channelsQuery().
		Where("application_id = $1", appID).
		Paginate(page, perPage).
		QueryJSON()
}

// validatePackage checks if a package belongs to the application provided,
// returning it if found.
func (api *API) validatePackage(packageID, appID string) (*Package, error) {
	pkg, err := api.GetPackage(packageID)
	if err == nil {
		if pkg.ApplicationID != appID {
			return nil, ErrInvalidPackage
		}
	}

	return pkg, nil
}

// channelsQuery returns a SelectDocBuilder prepared to return all channels.
// This query is meant to be extended later in the methods using it to filter
// by a specific channel id, all channels that belong to a given application,
// specify how to query the rows or their destination.
func (api *API) channelsQuery() *dat.SelectDocBuilder {
	return api.dbR.
		SelectDoc("*").
		One("package", api.channelPackageQuery()).
		From("channel").
		OrderBy("created_ts DESC")
}

// channelPackageQuery returns a SQL query prepared to return the package of a
// given channel.
func (api *API) channelPackageQuery() string {
	return "SELECT * FROM package WHERE package.id = channel.package_id"
}
