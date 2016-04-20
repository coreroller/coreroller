package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

// activityContext represents the context of a given activity entry.
type activityContext struct {
	appID      string
	groupID    string
	channelID  string
	instanceID string
}

// Activity represents a CoreRoller activity entry.
type Activity struct {
	CreatedTs       time.Time      `db:"created_ts" json:"created_ts"`
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
	PerPage    uint64    `json:"perpage"`
}

// GetActivity returns a list of activity entries that match the specified
// criteria in the query parameters.
func (api *API) GetActivity(teamID string, p ActivityQueryParams) ([]*Activity, error) {
	var activityEntries []*Activity

	err := api.activityQuery(teamID, p).QueryStructs(&activityEntries)

	return activityEntries, err
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
			LEFT JOIN groups g ON (a.group_id = g.id)
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

	if err != nil {
		return err
	}

	ctx := &activityContext{
		appID:   appID,
		groupID: groupID,
	}
	go api.postHipchat(class, severity, version, ctx)

	return nil
}

// newChannelActivityEntry creates a new activity entry related to a specific
// channel.
func (api *API) newChannelActivityEntry(class int, severity int, version, appID, channelID string) error {
	_, err := api.dbR.InsertInto("activity").
		Columns("class", "severity", "version", "application_id", "channel_id").
		Values(class, severity, version, appID, channelID).
		Exec()

	if err != nil {
		return err
	}

	ctx := &activityContext{
		appID:     appID,
		channelID: channelID,
	}
	go api.postHipchat(class, severity, version, ctx)

	return nil
}

// newInstanceActivityEntry creates a new activity entry related to a specific
// instance.
func (api *API) newInstanceActivityEntry(class int, severity int, version, appID, groupID, instanceID string) error {
	_, err := api.dbR.InsertInto("activity").
		Columns("class", "severity", "version", "application_id", "group_id", "instance_id").
		Values(class, severity, version, appID, groupID, instanceID).
		Exec()

	if err != nil {
		return err
	}

	ctx := &activityContext{
		appID:      appID,
		groupID:    groupID,
		instanceID: instanceID,
	}
	go api.postHipchat(class, severity, version, ctx)

	return nil
}

// EXPERIMENTAL! This is an experiment playing a bit with Hipchat
//
// postHipchat builds and posts a message representing the activity entry
// provided to Hipchat. This is just an experiment.
func (api *API) postHipchat(class, severity int, version string, ctx *activityContext) {
	room := os.Getenv("CR_HIPCHAT_ROOM")
	token := os.Getenv("CR_HIPCHAT_TOKEN")
	if room == "" || token == "" {
		return
	}

	var msg bytes.Buffer
	var color string

	app, _ := api.GetApp(ctx.appID)
	fmt.Fprintf(&msg, "<b>%s</b> ", app.Name)

	if ctx.groupID != "" {
		group, _ := api.GetGroup(ctx.groupID)
		fmt.Fprintf(&msg, "> <b>%s</b>", group.Name)
	}
	fmt.Fprint(&msg, "<br/>")

	switch class {
	case activityPackageNotFound:
		fmt.Fprint(&msg, "An update request could not be processed because the group's channel is not linked to any package")
		color = "red"
	case activityRolloutStarted:
		fmt.Fprintf(&msg, "Version <i>%s</i> roll out started", version)
		color = "purple"
	case activityRolloutFinished:
		fmt.Fprintf(&msg, "Version <i>%s</i> successfully rolled out", version)
		color = "green"
	case activityRolloutFailed:
		fmt.Fprintf(&msg, "There was an error rolling out version <i>%s</i> as the first update attempt failed. Group's updates have been disabled", version)
		color = "red"
	case activityInstanceUpdateFailed:
		instance, _ := api.GetInstance(ctx.instanceID, ctx.appID)
		fmt.Fprintf(&msg, "Instance <i>%s</i> reported an error while processing update to version <i>%s</i>", instance.IP, version)
		color = "yellow"
	case activityChannelPackageUpdated:
		channel, _ := api.GetChannel(ctx.channelID)
		fmt.Fprintf(&msg, "Channel <i>%s</i> is now pointing to version <i>%s</i>", channel.Name, version)
		color = "purple"
	}

	body := map[string]interface{}{
		"message_format": "html",
		"message":        msg.String(),
		"color":          color,
		"notify":         true,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return
	}

	url := "https://api.hipchat.com/v2/room/%s/notification?auth_token=%s"
	resp, err := http.Post(fmt.Sprintf(url, room, token), "application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		return
	}
	resp.Body.Close()
}
