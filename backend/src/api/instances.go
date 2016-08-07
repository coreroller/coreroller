package api

import (
	"fmt"
	"time"

	"github.com/satori/go.uuid"
	"gopkg.in/mgutz/dat.v1"
)

const (
	// InstanceStatusUndefined indicates that the instance hasn't sent yet an
	// event to CoreRoller so it doesn't know in which state it is.
	InstanceStatusUndefined int = 1 + iota

	// InstanceStatusUpdateGranted indicates that the instance has been granted
	// an update (it should be reporting soon through events how is it going).
	InstanceStatusUpdateGranted

	// InstanceStatusError indicates that the instance reported an error while
	// processing the update.
	InstanceStatusError

	// InstanceStatusComplete indicates that the instance completed the update
	// process successfully.
	InstanceStatusComplete

	// InstanceStatusInstalled indicates that the instance has installed the
	// downloaded packages, but it hasn't applied it or restarted yet.
	InstanceStatusInstalled

	// InstanceStatusDownloaded indicates that the instance downloaded
	// successfully the update package.
	InstanceStatusDownloaded

	// InstanceStatusDownloading indicates that the instance started
	// downloading the update package.
	InstanceStatusDownloading

	// InstanceStatusOnHold indicates that the instance hasn't been granted an
	// update because one of the rollout policy limits has been reached.
	InstanceStatusOnHold
)

const (
	validityInterval = "1 days"
)

// Instance represents an instance running one or more applications for which
// CoreRoller can provide updates.
type Instance struct {
	ID          string              `db:"id" json:"id"`
	IP          string              `db:"ip" json:"ip"`
	CreatedTs   time.Time           `db:"created_ts" json:"created_ts"`
	Application InstanceApplication `db:"application" json:"application,omitempty"`
}

// InstanceApplication represents some details about an application running on
// a given instance: current version of the app, last time the instance checked
// for updates for this app, etc.
type InstanceApplication struct {
	InstanceID          string         `db:"instance_id" json:"instance_id,omitempty"`
	ApplicationID       string         `db:"application_id" json:"application_id"`
	GroupID             dat.NullString `db:"group_id" json:"group_id"`
	Version             string         `db:"version" json:"version"`
	CreatedTs           time.Time      `db:"created_ts" json:"created_ts"`
	Status              dat.NullInt64  `db:"status" json:"status"`
	LastCheckForUpdates time.Time      `db:"last_check_for_updates" json:"last_check_for_updates"`
	LastUpdateGrantedTs dat.NullTime   `db:"last_update_granted_ts" json:"last_update_granted_ts"`
	LastUpdateVersion   dat.NullString `db:"last_update_version" json:"last_update_version"`
	UpdateInProgress    bool           `db:"update_in_progress" json:"update_in_progress"`
}

// InstanceStatusHistoryEntry represents an entry in the instance status
// history.
type InstanceStatusHistoryEntry struct {
	ID            int       `db:"id" json:"-"`
	Status        int       `db:"status" json:"status"`
	Version       string    `db:"version" json:"version"`
	CreatedTs     time.Time `db:"created_ts" json:"created_ts"`
	InstanceID    string    `db:"instance_id" json:"-"`
	ApplicationID string    `db:"application_id" json:"-"`
	GroupID       string    `db:"group_id" json:"-"`
}

// InstancesQueryParams represents a helper structure used to pass a set of
// parameters when querying instances.
type InstancesQueryParams struct {
	ApplicationID string `json:"application_id"`
	GroupID       string `json:"group_id"`
	Status        int    `json:"status"`
	Version       string `json:"version"`
	Page          uint64 `json:"page"`
	PerPage       uint64 `json:"perpage"`
}

// RegisterInstance registers an instance into CoreRoller.
func (api *API) RegisterInstance(instanceID, instanceIP, instanceVersion, appID, groupID string) (*Instance, error) {
	if !isValidSemver(instanceVersion) {
		return nil, ErrInvalidSemver
	}

	var err error
	if appID, groupID, err = api.validateApplicationAndGroup(appID, groupID); err != nil {
		return nil, err
	}

	tx, err := api.dbR.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.AutoRollback()
	}()

	result, err := tx.
		Upsert("instance").
		Columns("id", "ip").
		Values(instanceID, instanceIP).
		Where("id = $1", instanceID).
		Exec()

	if err != nil || result.RowsAffected == 0 {
		return nil, err
	}

	result, err = tx.
		Upsert("instance_application").
		Columns("instance_id", "application_id", "group_id", "version", "last_check_for_updates").
		Values(instanceID, appID, groupID, instanceVersion, nowUTC).
		Where("instance_id = $1 AND application_id = $2", instanceID, appID).
		Exec()

	if err != nil || result.RowsAffected == 0 {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return api.GetInstance(instanceID, appID)
}

// GetInstance returns the instance identified by the id provided.
func (api *API) GetInstance(instanceID, appID string) (*Instance, error) {
	var instance Instance

	err := api.dbR.
		SelectDoc("*").
		From("instance").
		One("application", api.instanceAppQuery(appID)).
		Where("id = $1", instanceID).
		QueryStruct(&instance)

	if err != nil {
		return nil, err
	}

	return &instance, nil
}

// GetInstanceStatusHistory returns the status history of an instance in the
// context of the application/group provided.
func (api *API) GetInstanceStatusHistory(instanceID, appID, groupID string, limit uint64) ([]*InstanceStatusHistoryEntry, error) {
	var instanceStatusHistory []*InstanceStatusHistoryEntry

	err := api.instanceStatusHistoryQuery(instanceID, appID, groupID, limit).QueryStructs(&instanceStatusHistory)

	return instanceStatusHistory, err
}

// GetInstances returns all instances that match with the provided criteria.
func (api *API) GetInstances(p InstancesQueryParams) ([]*Instance, error) {
	var instances []*Instance

	err := api.instancesQuery(p).QueryStructs(&instances)

	return instances, err
}

// validateApplicationAndGroup validates if the group provided belongs to the
// provided application, returning the normalized uuid version of the appID and
// groupID provided if both are valid and the group belongs to the given
// application, or an error if something goes wrong.
func (api *API) validateApplicationAndGroup(appID, groupID string) (string, string, error) {
	appUUID, err := uuid.FromString(appID)
	if err != nil {
		return "", "", err
	}
	groupUUID, err := uuid.FromString(groupID)
	if err != nil {
		return "", "", err
	}

	group, err := api.GetGroup(groupID)
	if err != nil {
		return "", "", err
	}

	if group.ApplicationID != appUUID.String() {
		return "", "", ErrInvalidApplicationOrGroup
	}

	return appUUID.String(), groupUUID.String(), nil
}

// updateInstanceStatus updates the status for the provided instance in the
// context of the given application, storing it as well in the instance status
// history registry.
func (api *API) updateInstanceStatus(instanceID, appID string, newStatus int) error {
	instance, err := api.GetInstance(instanceID, appID)
	if err != nil {
		return err
	}
	if instance.Application.Status.Valid && instance.Application.Status.Int64 == int64(newStatus) {
		return nil
	}

	query := api.dbR.
		Update("instance_application").
		Set("status", newStatus).
		Where("instance_id = $1 AND application_id = $2", instanceID, appID).
		Returning("last_update_version", "group_id")

	if newStatus == InstanceStatusComplete {
		query.Set("version", dat.UnsafeString("CASE WHEN last_update_version IS NOT NULL THEN last_update_version ELSE version END"))
	}

	if newStatus == InstanceStatusComplete || newStatus == InstanceStatusError {
		query.Set("update_in_progress", false)
	}

	var lastUpdateVersion, groupID dat.NullString

	if err := query.QueryScalar(&lastUpdateVersion, &groupID); err != nil {
		return err
	}

	_, err = api.dbR.
		InsertInto("instance_status_history").
		Columns("status", "version", "instance_id", "application_id", "group_id").
		Values(newStatus, lastUpdateVersion, instanceID, appID, groupID).
		Exec()

	return err
}

// instanceAppQuery returns a SelectDocBuilder prepared to return the app status
// of the app identified by the application id provided for a given instance.
func (api *API) instanceAppQuery(appID string) *dat.SelectDocBuilder {
	return api.dbR.
		SelectDoc("version", "status", "last_check_for_updates", "last_update_version", "update_in_progress", "application_id", "group_id").
		From("instance_application").
		Where("instance_id = instance.id AND application_id = $1", appID).
		Where(fmt.Sprintf("last_check_for_updates > now() at time zone 'utc' - interval '%s'", validityInterval))
}

// instancesQuery returns a SelectDocBuilder prepared to return all instances
// that match the criteria provided in InstancesQueryParams.
func (api *API) instancesQuery(p InstancesQueryParams) *dat.SelectDocBuilder {
	p.Page, p.PerPage = validatePaginationParams(p.Page, p.PerPage)

	instancesSubquery := api.dbR.
		Select("instance_id").
		From("instance_application").
		Where("application_id = $1 AND group_id = $2", p.ApplicationID, p.GroupID).
		Where(fmt.Sprintf("last_check_for_updates > now() at time zone 'utc' - interval '%s'", validityInterval)).
		Paginate(p.Page, p.PerPage)

	if p.Status != 0 {
		instancesSubquery.Where("status = $1", p.Status)
	}

	if p.Version != "" {
		instancesSubquery.Where("version = $1", p.Version)
	}

	instancesSubquerySQL, instancesSubqueryParams := instancesSubquery.ToSQL()

	return api.dbR.
		SelectDoc("*").
		One("application", api.instanceAppQuery(p.ApplicationID)).
		From("instance").
		Where(fmt.Sprintf("id IN (%s)", instancesSubquerySQL), instancesSubqueryParams...)
}

// instanceStatusHistoryQuery returns a SelectDocBuilder prepared to return the
// status history of a given instance in the context of an application/group.
func (api *API) instanceStatusHistoryQuery(instanceID, appID, groupID string, limit uint64) *dat.SelectDocBuilder {
	if limit <= 0 {
		limit = 20
	}

	return api.dbR.
		SelectDoc("status", "version", "created_ts").
		From("instance_status_history").
		Where("instance_id = $1", instanceID).
		Where("application_id = $1", appID).
		Where("group_id = $1", groupID).
		OrderBy("created_ts DESC").
		Limit(limit)
}
