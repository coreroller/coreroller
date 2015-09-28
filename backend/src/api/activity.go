package api

import (
	"fmt"
	"time"

	"gopkg.in/mgutz/dat.v1"
)

const (
	activityPackageNotFound int = 1 + iota
	activityRolloutStarted
	activityRolloutFinished
	activityRolloutFailed
	activityInstanceUpdateFailed
	activityChannelPackageUpdated
)

const (
	activitySuccess int = 1 + iota
	activityInfo
	activityWarning
	activityError
)

const (
	pgDateFormat = "2006-01-02 150405.000"
)

// Activity represents a CoreRoller activity entry.
type Activity struct {
	ID              int            `db:"id" json:"id"`
	CreatedTs       time.Time      `db:"created_ts" json:"-"`
	Class           int            `db:"class" json:"class"`
	Severity        int            `db:"severity" json:"severity"`
	Version         string         `db:"version" json:"version"`
	ApplicationName string         `db:"application_name" json:"application_name"`
	GroupName       dat.NullString `db:"group_name" json:"group_name"`
	ChannelName     dat.NullString `db:"channel_name" json:"channel_name"`
	InstanceID      dat.NullString `db:"instance_id" json:"instance_id"`
}

// ActivityQueryParams represents a helper structure used to pass a set of
// parameters when querying activity entries.
type ActivityQueryParams struct {
	AppID      string    `db:"application_id"`
	GroupID    string    `db:"group_id"`
	ChannelID  string    `db:"channel_id"`
	InstanceID string    `db:"instance_id"`
	Version    string    `db:"version"`
	Severity   int       `db:"severity"`
	Start      time.Time `db:"start"`
	End        time.Time `db:"end"`
	Page       uint64    `json:"page"`
	PerPage    uint64    `json:"per_page"`
}

// GetActivity returns a list of activity entries that match the specified
// criteria in the query parameters.
func (api *API) GetActivity(teamID string, p ActivityQueryParams) ([]*Activity, error) {
	var activityEntries []*Activity

	err := api.activityQuery(teamID, p).QueryStructs(&activityEntries)

	return activityEntries, err
}

// GetActivityJSON returns a list of activity entries that match the specified
// criteria in the query parameters in JSON format.
func (api *API) GetActivityJSON(teamID string, p ActivityQueryParams) ([]byte, error) {
	return api.activityQuery(teamID, p).QueryJSON()
}

// activityQuery returns a SelectDocBuilder prepared to return all activity
// entries that match the criteria provided in ActivityQueryParams.
func (api *API) activityQuery(teamID string, p ActivityQueryParams) *dat.SelectDocBuilder {
	p.Page, p.PerPage = validatePaginationParams(p.Page, p.PerPage)

	var start, end time.Time
	if !p.Start.IsZero() {
		start = p.Start.UTC()
	} else {
		start = time.Now().UTC().AddDate(0, 0, -3)
	}
	if !p.End.IsZero() {
		end = p.End.UTC()
	} else {
		end = time.Now().UTC()
	}

	query := api.dbR.
		SelectDoc("a.created_ts", "a.class", "a.severity", "a.version", "a.instance_id", "app.name as application_name", "g.name as group_name", "c.name as channel_name").
		From(`
			activity a 
			INNER JOIN application app ON (a.application_id = app.id)
			INNER JOIN groups g ON (a.group_id = g.id)
			LEFT JOIN channel c ON (a.channel_id = c.id)
		`).
		Where("app.team_id = $1", teamID).
		Where(fmt.Sprintf("a.created_ts BETWEEN '%s' AND '%s'", start.Format(pgDateFormat), end.Format(pgDateFormat))).
		Paginate(p.Page, p.PerPage).
		OrderBy("a.created_ts DESC")

	if p.AppID != "" {
		query.Where("app.id = $1", p.AppID)
	}

	if p.GroupID != "" {
		query.Where("g.id = $1", p.GroupID)
	}

	if p.ChannelID != "" {
		query.Where("c.id = $1", p.ChannelID)
	}

	if p.InstanceID != "" {
		query.Where("a.instance_id = $1", p.InstanceID)
	}

	if p.Version != "" {
		query.Where("a.version = $1", p.Version)
	}

	if p.Severity != 0 {
		query.Where("a.severity = $1", p.Severity)
	}

	return query
}

// newGroupActivityEntry creates a new activity entry related to a specific
// group.
func (api *API) newGroupActivityEntry(class int, severity int, version, appID, groupID string) error {
	_, err := api.dbR.InsertInto("activity").
		Columns("class", "severity", "version", "application_id", "group_id").
		Values(class, severity, version, appID, groupID).
		Exec()

	return err
}

// newChannelActivityEntry creates a new activity entry related to a specific
// channel.
func (api *API) newChannelActivityEntry(class int, severity int, version, appID, channelID string) error {
	_, err := api.dbR.InsertInto("activity").
		Columns("class", "severity", "version", "application_id", "channel_id").
		Values(class, severity, version, appID, channelID).
		Exec()

	return err
}

// newInstanceActivityEntry creates a new activity entry related to a specific
// instance.
func (api *API) newInstanceActivityEntry(class int, severity int, version, appID, groupID, instanceID string) error {
	_, err := api.dbR.InsertInto("activity").
		Columns("class", "severity", "version", "application_id", "group_id", "instance_id").
		Values(class, severity, version, appID, groupID, instanceID).
		Exec()

	return err
}
