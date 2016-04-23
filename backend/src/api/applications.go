package api

import (
	"fmt"
	"time"

	"gopkg.in/mgutz/dat.v1"
)

const (
	coreosAppID = "e96281a6-d1af-4bde-9a0a-97b76e56dc57"
)

// Application represents a CoreRoller application instance.
type Application struct {
	ID          string     `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	CreatedTs   time.Time  `db:"created_ts" json:"created_ts"`
	TeamID      string     `db:"team_id" json:"-"`
	Groups      []*Group   `db:"groups" json:"groups"`
	Channels    []*Channel `db:"channels" json:"channels"`
	Packages    []*Package `db:"packages" json:"packages"`

	Instances struct {
		Count int `db:"count" json:"count"`
	} `db:"instances" json:"instances,omitempty"`
}

// AddApp registers the provided application.
func (api *API) AddApp(app *Application) (*Application, error) {
	err := api.dbR.
		InsertInto("application").
		Whitelist("name", "description", "team_id").
		Record(app).
		Returning("*").
		QueryStruct(app)

	return app, err
}

// AddAppCloning registers the provided application, cloning the groups and
// channels from an existing application. Channels' packages will be set to null
// as packages won't be cloned.
func (api *API) AddAppCloning(app *Application, sourceAppID string) (*Application, error) {
	app, err := api.AddApp(app)
	if err != nil {
		return nil, err
	}

	// NOTE: cloning operation is not transactional and something could go wrong

	if sourceAppID != "" {

		sourceApp, err := api.GetApp(sourceAppID)
		if err != nil {
			return app, nil
		}

		channelsIDsMappings := make(map[string]dat.NullString)

		for _, channel := range sourceApp.Channels {
			originalChannelID := channel.ID
			channel.ApplicationID = app.ID
			channel.PackageID = dat.NullString{}
			channelCopy, err := api.AddChannel(channel)
			if err != nil {
				return app, nil // FIXME - think about what we should return to the caller
			}
			channelsIDsMappings[originalChannelID] = dat.NullStringFrom(channelCopy.ID)
		}

		for _, group := range sourceApp.Groups {
			group.ApplicationID = app.ID
			if group.ChannelID.String != "" {
				group.ChannelID = channelsIDsMappings[group.ChannelID.String]
			}
			group.PolicyUpdatesEnabled = true
			if _, err := api.AddGroup(group); err != nil {
				return app, nil // FIXME - think about what we should return to the caller
			}
		}
	}

	return app, nil
}

// UpdateApp updates an existing application using the content of the
// application provided.
func (api *API) UpdateApp(app *Application) error {
	result, err := api.dbR.
		Update("application").
		SetWhitelist(app, "name", "description").
		Where("id = $1", app.ID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// DeleteApp removes the application identified by the id provided.
func (api *API) DeleteApp(appID string) error {
	result, err := api.dbR.
		DeleteFrom("application").
		Where("id = $1", appID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// GetApp returns the application identified by the id provided.
func (api *API) GetApp(appID string) (*Application, error) {
	var app Application

	err := api.appsQuery().
		Where("id = $1", appID).
		QueryStruct(&app)

	if err != nil {
		return nil, err
	}

	return &app, nil
}

// GetApps returns all applications that belong to the team id provided.
func (api *API) GetApps(teamID string, page, perPage uint64) ([]*Application, error) {
	page, perPage = validatePaginationParams(page, perPage)

	var apps []*Application

	err := api.appsQuery().
		Where("team_id = $1", teamID).
		Paginate(page, perPage).
		QueryStructs(&apps)

	return apps, err
}

// appsQuery returns a SelectDocBuilder prepared to return all applications.
// This query is meant to be extended later in the methods using it to filter
// by a specific application id, all applications that belong to a given team,
// specify how to query the rows or their destination.
func (api *API) appsQuery() *dat.SelectDocBuilder {
	return api.dbR.
		SelectDoc("id, name, description, created_ts").
		One("instances", api.appInstancesCountQuery()).
		Many("groups", api.groupsQuery().Where("application_id = application.id")).
		Many("channels", api.channelsQuery().Where("application_id = application.id")).
		Many("packages", api.packagesQuery().Where("application_id = application.id")).
		From("application").
		OrderBy("created_ts DESC")
}

// appInstancesCountQuery returns a SQL query prepared to return the number of
// instances running a given application.
func (api *API) appInstancesCountQuery() string {
	return fmt.Sprintf(`
	SELECT count(*)
	FROM instance_application 
	WHERE application_id = application.id AND 
	      last_check_for_updates > now() at time zone 'utc' - interval '%s'
	`, validityInterval)
}
