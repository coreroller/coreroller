package api

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/mgutz/dat.v1"
)

var (
	// ErrInvalidChannel error indicates that a channel doesn't belong to the
	// application it was supposed to belong to.
	ErrInvalidChannel = errors.New("coreroller: invalid channel")

	// ErrExpectingValidTimezone error indicates that a valid timezone wasn't
	// provided when enabling the flag PolicyOfficeHours.
	ErrExpectingValidTimezone = errors.New("coreroller: expecting valid timezone")
)

// Group represents a CoreRoller application's group.
type Group struct {
	ID                        string                   `db:"id" json:"id"`
	Name                      string                   `db:"name" json:"name"`
	Description               string                   `db:"description" json:"description"`
	CreatedTs                 time.Time                `db:"created_ts" json:"created_ts"`
	RolloutInProgress         bool                     `db:"rollout_in_progress" json:"rollout_in_progress"`
	ApplicationID             string                   `db:"application_id" json:"application_id"`
	ChannelID                 dat.NullString           `db:"channel_id" json:"channel_id"`
	PolicyUpdatesEnabled      bool                     `db:"policy_updates_enabled" json:"policy_updates_enabled"`
	PolicySafeMode            bool                     `db:"policy_safe_mode" json:"policy_safe_mode"`
	PolicyOfficeHours         bool                     `db:"policy_office_hours" json:"policy_office_hours"`
	PolicyTimezone            dat.NullString           `db:"policy_timezone" json:"policy_timezone"`
	PolicyPeriodInterval      string                   `db:"policy_period_interval" json:"policy_period_interval"`
	PolicyMaxUpdatesPerPeriod int                      `db:"policy_max_updates_per_period" json:"policy_max_updates_per_period"`
	PolicyUpdateTimeout       string                   `db:"policy_update_timeout" json:"policy_update_timeout"`
	VersionBreakdown          []*VersionBreakdownEntry `db:"version_breakdown" json:"version_breakdown,omitempty"`
	Channel                   *Channel                 `db:"channel" json:"channel,omitempty"`
	InstancesStats            InstancesStatusStats     `db:"instances_stats" json:"instances_stats,omitempty"`
}

// VersionBreakdownEntry represents the distribution of the versions currently
// installed in the instances belonging to a given group.
type VersionBreakdownEntry struct {
	Version    string  `db:"version" json:"version"`
	Instances  int     `db:"instances" json:"instances"`
	Percentage float64 `db:"percentage" json:"percentage"`
}

// InstancesStatusStats represents a set of statistics about the status of the
// instances that belong to a given group.
type InstancesStatusStats struct {
	Total         int `db:"total" json:"total"`
	Undefined     int `db:"undefined" json:"undefined"`
	UpdateGranted int `db:"update_granted" json:"update_granted"`
	Error         int `db:"error" json:"error"`
	Complete      int `db:"complete" json:"complete"`
	Installed     int `db:"installed" json:"installed"`
	Downloaded    int `db:"downloaded" json:"downloaded"`
	Downloading   int `db:"downloading" json:"downloading"`
	OnHold        int `db:"onhold" json:"onhold"`
}

// UpdatesStats represents a set of statistics about the status of the updates
// that may be taking place in the instaces belonging to a given group.
type UpdatesStats struct {
	TotalInstances                   int `db:"total_instances"`
	UpdatesToCurrentVersionGranted   int `db:"updates_to_current_version_granted"`
	UpdatesToCurrentVersionAttempted int `db:"updates_to_current_version_attempted"`
	UpdatesToCurrentVersionSucceeded int `db:"updates_to_current_version_succeeded"`
	UpdatesToCurrentVersionFailed    int `db:"updates_to_current_version_failed"`
	UpdatesGrantedInLastPeriod       int `db:"updates_granted_in_last_period"`
	UpdatesInProgress                int `db:"updates_in_progress"`
	UpdatesTimedOut                  int `db:"updates_timed_out"`
}

// AddGroup registers the provided group.
func (api *API) AddGroup(group *Group) (*Group, error) {
	if group.PolicyOfficeHours && !isTimezoneValid(group.PolicyTimezone.String) {
		return nil, ErrExpectingValidTimezone
	}

	if group.ChannelID.String != "" {
		if err := api.validateChannel(group.ChannelID.String, group.ApplicationID); err != nil {
			return nil, err
		}
	}

	err := api.dbR.
		InsertInto("groups").
		Whitelist("name", "description", "application_id", "channel_id", "policy_updates_enabled", "policy_safe_mode", "policy_office_hours",
			"policy_timezone", "policy_period_interval", "policy_max_updates_per_period", "policy_update_timeout").
		Record(group).
		Returning("*").
		QueryStruct(group)

	return group, err
}

// UpdateGroup updates an existing group using the context of the group
// provided.
func (api *API) UpdateGroup(group *Group) error {
	if group.PolicyOfficeHours && !isTimezoneValid(group.PolicyTimezone.String) {
		return ErrExpectingValidTimezone
	}

	groupBeforeUpdate, err := api.GetGroup(group.ID)
	if err != nil {
		return err
	}

	if group.ChannelID.String != "" {
		if err := api.validateChannel(group.ChannelID.String, groupBeforeUpdate.ApplicationID); err != nil {
			return err
		}
	}

	result, err := api.dbR.
		Update("groups").
		SetWhitelist(group, "name", "description", "channel_id", "policy_updates_enabled", "policy_safe_mode", "policy_office_hours",
			"policy_timezone", "policy_period_interval", "policy_max_updates_per_period", "policy_update_timeout").
		Where("id = $1", group.ID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// DeleteGroup removes the group identified by the id provided.
func (api *API) DeleteGroup(groupID string) error {
	result, err := api.dbR.
		DeleteFrom("groups").
		Where("id = $1", groupID).
		Exec()

	if err == nil && result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}

	return err
}

// GetGroup returns the group identified by the id provided.
func (api *API) GetGroup(groupID string) (*Group, error) {
	var group Group

	err := api.groupsQuery().
		Where("id = $1", groupID).
		QueryStruct(&group)

	if err != nil {
		return nil, err
	}

	return &group, nil
}

// GetGroups returns all groups that belong to the application provided.
func (api *API) GetGroups(appID string, page, perPage uint64) ([]*Group, error) {
	page, perPage = validatePaginationParams(page, perPage)

	var groups []*Group

	err := api.groupsQuery().
		Where("application_id = $1", appID).
		Paginate(page, perPage).
		QueryStructs(&groups)

	return groups, err
}

// validateChannel checks if a channel belongs to the application provided.
func (api *API) validateChannel(channelID, appID string) error {
	channel, err := api.GetChannel(channelID)
	if err == nil {
		if channel.ApplicationID != appID {
			return ErrInvalidChannel
		}
	}

	return nil
}

// getGroupUpdatesStats returns a set of statistics about the distribution of
// updates and their status in the group provided.
func (api *API) getGroupUpdatesStats(group *Group) (*UpdatesStats, error) {
	var updatesStats UpdatesStats

	packageVersion := ""
	if group.Channel.Package != nil {
		packageVersion = group.Channel.Package.Version
	}

	query := fmt.Sprintf(`
	SELECT
		count(*) total_instances,
		sum(case when last_update_version = $1 then 1 else 0 end) updates_to_current_version_granted, 
		sum(case when update_in_progress = 'false' and last_update_version = $1 then 1 else 0 end) updates_to_current_version_attempted, 
		sum(case when update_in_progress = 'false' and last_update_version = $1 and last_update_version = version then 1 else 0 end) updates_to_current_version_succeeded, 
		sum(case when update_in_progress = 'false' and last_update_version = $1 and last_update_version != version then 1 else 0 end) updates_to_current_version_failed, 
		sum(case when last_update_granted_ts > now() at time zone 'utc' - interval $2 then 1 else 0 end) updates_granted_in_last_period,
		sum(case when update_in_progress = 'true' and now() at time zone 'utc' - last_update_granted_ts <= interval $3 then 1 else 0 end) updates_in_progress,
		sum(case when update_in_progress = 'true' and now() at time zone 'utc' - last_update_granted_ts > interval $4 then 1 else 0 end) updates_timed_out
	FROM instance_application
	WHERE group_id=$5 AND last_check_for_updates > now() at time zone 'utc' - interval '%s'
	`, validityInterval)

	err := api.dbR.SQL(query, packageVersion, group.PolicyPeriodInterval, group.PolicyUpdateTimeout, group.PolicyUpdateTimeout, group.ID).
		QueryStruct(&updatesStats)
	if err != nil {
		return nil, err
	}

	return &updatesStats, nil
}

// disableUpdates updates the group provided setting the policy_updates_enabled
// field to false. This usually happens when the first instance in a group
// processing an update to a specific version fails if safe mode is enabled.
func (api *API) disableUpdates(groupID string) error {
	_, err := api.dbR.
		Update("groups").
		Set("policy_updates_enabled", false).
		Where("id = $1", groupID).
		Exec()

	return err
}

// setGroupRolloutInProgress updates the value of the rollout_in_progress flag
// for a given group, indicating if a rollout is taking place now or not.
func (api *API) setGroupRolloutInProgress(groupID string, inProgress bool) error {
	_, err := api.dbR.
		Update("groups").
		Set("rollout_in_progress", inProgress).
		Where("id = $1", groupID).
		Exec()

	return err
}

// groupsQuery returns a SelectDocBuilder prepared to return all groups. This
// query is meant to be extended later in the methods using it to filter by a
// specific group id, all groups of a given app, specify how to query the rows
// or their destination.
func (api *API) groupsQuery() *dat.SelectDocBuilder {
	return api.dbR.
		SelectDoc("*").
		One("instances_stats", api.groupInstancesStatusQuery()).
		One("channel", api.channelsQuery().Where("id = groups.channel_id")).
		Many("version_breakdown", api.groupVersionBreakdownQuery()).
		From("groups").
		OrderBy("created_ts DESC")
}

// groupVersionBreakdownQuery returns a SQL query prepared to return the version
// breakdown of all instances running on a given group.
func (api *API) groupVersionBreakdownQuery() string {
	return fmt.Sprintf(`
	SELECT version, count(*) as instances, (count(*) * 100.0 / total) as percentage
	FROM instance_application, (
		SELECT count(*) as total 
		FROM instance_application 
		WHERE group_id=groups.id AND last_check_for_updates > now() at time zone 'utc' - interval '%s'
		) totals
	WHERE group_id=groups.id AND last_check_for_updates > now() at time zone 'utc' - interval '%s'
	GROUP BY version, total
	ORDER BY regexp_matches(version, '(\d+)\.(\d+)\.(\d+)')::int[] DESC
	`, validityInterval, validityInterval)
}

// groupInstancesStatusQuery returns a SQL query prepared to return a summary
// of the status of the instances that belong to a given group.
func (api *API) groupInstancesStatusQuery() string {
	return fmt.Sprintf(`
	SELECT
		count(*) total,
		sum(case when status IS NULL then 1 else 0 end) undefined,
		sum(case when status = %d then 1 else 0 end) error,
		sum(case when status = %d then 1 else 0 end) update_granted,
		sum(case when status = %d then 1 else 0 end) complete,
		sum(case when status = %d then 1 else 0 end) installed,
		sum(case when status = %d then 1 else 0 end) downloaded,
		sum(case when status = %d then 1 else 0 end) downloading,
		sum(case when status = %d then 1 else 0 end) onhold
	FROM instance_application
	WHERE group_id=groups.id AND last_check_for_updates > now() at time zone 'utc' - interval '%s'`,
		InstanceStatusError, InstanceStatusUpdateGranted, InstanceStatusComplete, InstanceStatusInstalled,
		InstanceStatusDownloaded, InstanceStatusDownloading, InstanceStatusOnHold, validityInterval)
}
