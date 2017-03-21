package api

import (
	"errors"
	"time"

	"gopkg.in/mgutz/dat.v1"
)

const (
	// EventUpdateComplete indicates that the update process completed. It could
	// mean a successful or failed updated, depending on the result attached to
	// the event. This applies to all events.
	EventUpdateComplete = 3

	// EventUpdateDownloadStarted indicates that the instance started
	// downloading the update package.
	EventUpdateDownloadStarted = 13

	// EventUpdateDownloadFinished indicates that the update package was
	// downloaded.
	EventUpdateDownloadFinished = 14

	// EventUpdateInstalled indicates that the update package was installed.
	EventUpdateInstalled = 800
)

const (
	// ResultFailed indicates that the operation associated with the event
	// posted failed.
	ResultFailed = 0

	// ResultSuccess indicates that the operation associated with the event
	// posted succeeded.
	ResultSuccess = 1

	// ResultSuccessReboot also indicates a successful operation, but it's
	// meant only to be used along with events of EventUpdateComplete type.
	// It's important that instances use EventUpdateComplete events in
	// combination with ResultSuccessReboot to communicate a successful update
	// completed as it has a special meaning for CoreRoller in order to adjust
	// properly the rollout policies and create activity entries.
	ResultSuccessReboot = 2
)

var (
	// ErrInvalidInstance indicates that the instance provided is not valid or
	// it doesn't exist.
	ErrInvalidInstance = errors.New("coreroller: invalid instance")

	// ErrInvalidApplicationOrGroup indicates that the application or group id
	// provided are not valid or related to each other.
	ErrInvalidApplicationOrGroup = errors.New("coreroller: invalid application or group")

	// ErrInvalidEventTypeOrResult indicates that the event or result provided
	// are not valid (CoreRoller only implements a subset of the Omaha protocol
	// events).
	ErrInvalidEventTypeOrResult = errors.New("coreroller: invalid event type or result")

	// ErrEventRegistrationFailed indicates that the event registration into
	// CoreRoller failed.
	ErrEventRegistrationFailed = errors.New("coreroller: event registration failed")

	// ErrNoUpdateInProgress indicates that an event was received but there
	// wasn't an update in progress for the provided instance/application, so
	// it was rejected.
	ErrNoUpdateInProgress = errors.New("coreroller: no update in progress")

	// ErrCoreosEventIgnored indicates that a CoreOS updater event was ignored.
	// This is a temporary solution to handle CoreOS specific behaviour.
	ErrCoreosEventIgnored = errors.New("coreroller: coreos event ignored")
)

// Event represents an event posted by an instance to CoreRoller.
type Event struct {
	ID              int            `db:"id" json:"id"`
	CreatedTs       time.Time      `db:"created_ts" json:"created_ts"`
	PreviousVersion dat.NullString `db:"previous_version" json:"previous_version"`
	ErrorCode       dat.NullString `db:"error_code" json:"error_code"`
	InstanceID      string         `db:"instance_id" json:"instance_id"`
	ApplicationID   string         `db:"application_id" json:"application_id"`
	EventTypeID     string         `db:"event_type_id" json:"event_type_id"`
}

// RegisterEvent registers an event posted by an instance in CoreRoller. The
// event will be bound to an application/group combination.
func (api *API) RegisterEvent(instanceID, appID, groupID string, etype, eresult int, previousVersion, errorCode string) error {
	var err error
	if appID, groupID, err = api.validateApplicationAndGroup(appID, groupID); err != nil {
		return err
	}

	instance, err := api.GetInstance(instanceID, appID)
	if err != nil {
		return ErrInvalidInstance
	}
	if instance.Application.ApplicationID != appID {
		return ErrInvalidApplicationOrGroup
	}
	if !instance.Application.UpdateInProgress {
		return ErrNoUpdateInProgress
	}

	// Temporary hack to handle CoreOS updater specific behaviour
	if appID == coreosAppID && etype == EventUpdateComplete && eresult == ResultSuccessReboot {
		if previousVersion == "" || previousVersion == "0.0.0.0" || previousVersion != instance.Application.Version {
			return ErrCoreosEventIgnored
		}
	}

	var eventTypeID int
	err = api.dbR.
		Select("id").
		From("event_type").
		Where("type = $1 and result = $2", etype, eresult).
		QueryScalar(&eventTypeID)
	if err != nil {
		return ErrInvalidEventTypeOrResult
	}

	_, err = api.dbR.
		InsertInto("event").
		Columns("event_type_id", "instance_id", "application_id", "previous_version", "error_code").
		Values(eventTypeID, instanceID, appID, previousVersion, errorCode).
		Exec()

	if err != nil {
		return ErrEventRegistrationFailed
	}

	lastUpdateVersion := instance.Application.LastUpdateVersion.String
	_ = api.triggerEventConsequences(instanceID, appID, groupID, lastUpdateVersion, etype, eresult)

	return nil
}

// triggerEventConsequences is in charge of triggering the consequences of a
// given event. Depending on the type of the event and its result, the status
// of the instance may be updated, new activity entries could be created, etc.
func (api *API) triggerEventConsequences(instanceID, appID, groupID, lastUpdateVersion string, etype, result int) error {
	group, err := api.GetGroup(groupID)
	if err != nil {
		return err
	}

	// TODO: should we also consider ResultSuccess in the next check? CoreOS ~ generic conflicts?
	if etype == EventUpdateComplete && result == ResultSuccessReboot {
		_ = api.updateInstanceStatus(instanceID, appID, InstanceStatusComplete)

		updatesStats, err := api.getGroupUpdatesStats(group)
		if err != nil {
			return err
		}
		if updatesStats.UpdatesToCurrentVersionSucceeded == updatesStats.TotalInstances {
			_ = api.setGroupRolloutInProgress(groupID, false)
			_ = api.newGroupActivityEntry(activityRolloutFinished, activitySuccess, lastUpdateVersion, appID, groupID)
		}
	}

	if etype == EventUpdateDownloadStarted && result == ResultSuccess {
		_ = api.updateInstanceStatus(instanceID, appID, InstanceStatusDownloading)
	}

	if etype == EventUpdateDownloadFinished && result == ResultSuccess {
		_ = api.updateInstanceStatus(instanceID, appID, InstanceStatusDownloaded)
	}

	if etype == EventUpdateInstalled && result == ResultSuccess {
		_ = api.updateInstanceStatus(instanceID, appID, InstanceStatusInstalled)
	}

	if result == ResultFailed {
		_ = api.updateInstanceStatus(instanceID, appID, InstanceStatusError)
		_ = api.newInstanceActivityEntry(activityInstanceUpdateFailed, activityError, lastUpdateVersion, appID, groupID, instanceID)

		updatesStats, err := api.getGroupUpdatesStats(group)
		if err != nil {
			return err
		}
		if updatesStats.UpdatesToCurrentVersionAttempted == 1 {
			_ = api.disableUpdates(groupID)
			_ = api.setGroupRolloutInProgress(groupID, false)
			_ = api.newGroupActivityEntry(activityRolloutFailed, activityError, lastUpdateVersion, appID, groupID)
		}
	}

	return nil
}
