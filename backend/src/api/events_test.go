package api

import (
	"database/sql"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgutz/dat.v1"
)

func TestRegisterEvent_InvalidParams(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	err := a.RegisterEvent(uuid.NewV4().String(), tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")
	assert.Equal(t, ErrInvalidInstance, err)

	err = a.RegisterEvent(tInstance.ID, uuid.NewV4().String(), tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")
	assert.Equal(t, ErrInvalidApplicationOrGroup, err)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, uuid.NewV4().String(), EventUpdateComplete, ResultSuccessReboot, "", "")
	assert.Equal(t, sql.ErrNoRows, err)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "", "")
	assert.Equal(t, ErrNoUpdateInProgress, err)

	_, _ = a.GetUpdatePackage(tInstance.ID, "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, 1000, ResultSuccess, "", "")
	assert.Equal(t, ErrInvalidEventTypeOrResult, err)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, 1000, "", "")
	assert.Equal(t, ErrInvalidEventTypeOrResult, err)
}

func TestRegisterEvent_TriggerEventConsequences(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	tInstance2, _ := a.RegisterInstance(uuid.NewV4().String(), "10.0.0.2", "1.0.0", tApp.ID, tGroup.ID)

	_, err := a.GetUpdatePackage(tInstance.ID, "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	err = a.RegisterEvent(tInstance.ID, "{"+tApp.ID+"}", tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "", "")
	assert.NoError(t, err)
	instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, dat.NullInt64From(int64(InstanceStatusDownloading)), instance.Application.Status)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, "{"+tGroup.ID+"}", EventUpdateDownloadFinished, ResultSuccess, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, dat.NullInt64From(int64(InstanceStatusDownloaded)), instance.Application.Status)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateInstalled, ResultSuccess, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, dat.NullInt64From(int64(InstanceStatusInstalled)), instance.Application.Status)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, dat.NullInt64From(int64(InstanceStatusComplete)), instance.Application.Status)

	_, err = a.GetUpdatePackage(tInstance2.ID, "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	err = a.RegisterEvent(tInstance2.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultFailed, "", "")
	assert.NoError(t, err)
	instance, _ = a.GetInstance(tInstance2.ID, tApp.ID)
	assert.Equal(t, dat.NullInt64From(int64(InstanceStatusError)), instance.Application.Status)
	group, _ := a.GetGroup(tGroup.ID)
	assert.Equal(t, true, group.PolicyUpdatesEnabled, "It wasn't the first update the one that failed.")
}

func TestRegisterEvent_TriggerEventConsequences_FirstUpdateAttemptFailed(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	_, err := a.GetUpdatePackage(tInstance.ID, "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	err = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultFailed, "", "")
	assert.NoError(t, err)
	instance, _ := a.GetInstance(tInstance.ID, tApp.ID)
	assert.Equal(t, dat.NullInt64From(int64(InstanceStatusError)), instance.Application.Status)
	group, _ := a.GetGroup(tGroup.ID)
	assert.Equal(t, false, group.PolicyUpdatesEnabled, "First update attempt failed.")
}
