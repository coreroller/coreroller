package api

import (
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgutz/dat.v1"
)

func TestGetUpdatePackage(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tChannel2, _ := a.AddChannel(&Channel{Name: "test_channel2", Color: "green", ApplicationID: tApp2.ID})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := a.AddGroup(&Group{Name: "group2", ApplicationID: tApp2.ID, ChannelID: dat.NullStringFrom(tChannel2.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	_, err := a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "1.0.0", "invalidApplicationID", tGroup.ID)
	assert.Error(t, ErrInvalidApplicationOrGroup, "Invalid application id.")

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, "invalidGroupID")
	assert.Error(t, err, "Invalid group id.")

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "1.0.0", uuid.NewV4().String(), tGroup.ID)
	assert.Error(t, err, "Non existent application id.")

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, uuid.NewV4().String())
	assert.Error(t, err, "Non existent group id.")

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, tGroup2.ID)
	assert.Error(t, err, "Group doesn't belong to the application provided.")

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp2.ID, tGroup2.ID)
	assert.Equal(t, ErrNoPackageFound, err, "Group's channel has no package bound.")

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "12.1.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrNoUpdatePackageAvailable, err, "Instance version is up to date.")

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "1010.5.0+2016-05-27-1832", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrNoUpdatePackageAvailable, err, "Instance version is up to date.")
}

func TestGetUpdatePackage_GroupNoChannel(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, PolicyUpdatesEnabled: false, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	_, _ = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.Error(t, ErrNoPackageFound)
}

func TestGetUpdatePackage_UpdatesDisabled(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: false, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	_, err := a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrUpdatesDisabled, err)
}

func TestGetUpdatePackage_MaxUpdatesPerPeriodLimitReached_SafeMode(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	safeMode := true

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: safeMode, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	_, err := a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxUpdatesPerPeriodLimitReached, err, "Safe mode is enabled, first update should be completed before letting more through.")
}

func TestGetUpdatePackage_MaxUpdatesPerPeriodLimitReached_LimitUpdated(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 1, PolicyUpdateTimeout: "60 minutes"})

	instanceID := uuid.NewV4().String()
	_, err := a.GetUpdatePackage(instanceID, "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxUpdatesPerPeriodLimitReached, err, "Max 1 update per period, limit reached")

	tGroup.PolicyMaxUpdatesPerPeriod = 2
	_ = a.UpdateGroup(tGroup)

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)
}

func TestGetUpdatePackage_MaxUpdatesLimitsReached(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	maxUpdatesPerPeriod := 2
	periodInterval := "100 milliseconds"

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: periodInterval, PolicyMaxUpdatesPerPeriod: maxUpdatesPerPeriod, PolicyUpdateTimeout: "60 minutes"})

	newInstance1ID := uuid.NewV4().String()

	_, err := a.GetUpdatePackage(newInstance1ID, "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxUpdatesPerPeriodLimitReached, err)

	time.Sleep(100 * time.Millisecond)

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxConcurrentUpdatesLimitReached, err, "Period interval is over, but there are still two updates not completed or failed.")

	_ = a.updateInstanceStatus(newInstance1ID, tApp.ID, InstanceStatusComplete)

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)
}

func TestGetUpdatePackage_MaxTimedOutUpdatesLimitReached(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	maxUpdatesPerPeriod := 2
	periodInterval := "100 milliseconds"
	updateTimeout := "200 milliseconds"

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: periodInterval, PolicyMaxUpdatesPerPeriod: maxUpdatesPerPeriod, PolicyUpdateTimeout: updateTimeout})

	_, err := a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxConcurrentUpdatesLimitReached, err)

	time.Sleep(100 * time.Millisecond)

	_, err = a.GetUpdatePackage(uuid.NewV4().String(), "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrMaxTimedOutUpdatesLimitReached, err)
}

func TestGetUpdatePackage_RolloutStats(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 4, PolicyUpdateTimeout: "60 minutes"})

	instance1, _ := a.RegisterInstance(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	instance2, _ := a.RegisterInstance(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	instance3, _ := a.RegisterInstance(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	_, _ = a.GetUpdatePackage(instance1.ID, "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	_, _ = a.GetUpdatePackage(instance2.ID, "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	_, _ = a.GetUpdatePackage(instance3.ID, "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)

	group, _ := a.GetGroup(tGroup.ID)
	assert.True(t, group.RolloutInProgress)
	assert.Equal(t, 3, group.InstancesStats.Total)
	assert.Equal(t, 1, group.InstancesStats.UpdateGranted)
	assert.Equal(t, 2, group.InstancesStats.OnHold)

	_ = a.RegisterEvent(instance1.ID, tApp.ID, tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "", "")

	group, _ = a.GetGroup(tGroup.ID)
	assert.True(t, group.RolloutInProgress)
	assert.Equal(t, 1, group.InstancesStats.Downloading)
	assert.Equal(t, 2, group.InstancesStats.OnHold)

	_ = a.RegisterEvent(instance1.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")
	_, _ = a.GetUpdatePackage(instance2.ID, "10.0.0.2", "12.0.0", tApp.ID, tGroup.ID)
	_, _ = a.GetUpdatePackage(instance3.ID, "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)

	group, _ = a.GetGroup(tGroup.ID)
	assert.True(t, group.RolloutInProgress)
	assert.Equal(t, 1, group.InstancesStats.Complete)
	assert.Equal(t, 2, group.InstancesStats.UpdateGranted)

	_ = a.RegisterEvent(instance2.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")
	_ = a.RegisterEvent(instance3.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultFailed, "", "")

	group, _ = a.GetGroup(tGroup.ID)
	assert.True(t, group.RolloutInProgress)
	assert.Equal(t, 2, group.InstancesStats.Complete)
	assert.Equal(t, 1, group.InstancesStats.Error)

	_, _ = a.GetUpdatePackage(instance3.ID, "10.0.0.3", "12.0.0", tApp.ID, tGroup.ID)
	_ = a.RegisterEvent(instance3.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")

	group, _ = a.GetGroup(tGroup.ID)
	assert.False(t, group.RolloutInProgress)
	assert.Equal(t, 3, group.InstancesStats.Complete)
}

func TestGetUpdatePackage_UpdateInProgressOnInstance(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: false, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	instanceID := uuid.NewV4().String()

	_, err := a.GetUpdatePackage(instanceID, "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetUpdatePackage(instanceID, "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	assert.Equal(t, ErrUpdateInProgressOnInstance, err)
}

func TestGetUpdatePackage_InstanceStatusHistory(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 3, PolicyUpdateTimeout: "60 minutes"})

	instance1, _ := a.RegisterInstance(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	_, _ = a.GetUpdatePackage(instance1.ID, "10.0.0.1", "12.0.0", tApp.ID, tGroup.ID)
	_ = a.RegisterEvent(instance1.ID, tApp.ID, tGroup.ID, EventUpdateDownloadStarted, ResultSuccess, "", "")
	_ = a.RegisterEvent(instance1.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")

	instanceStatusHistory, err := a.GetInstanceStatusHistory(instance1.ID, tApp.ID, tGroup.ID, 5)
	assert.NoError(t, err)
	assert.Equal(t, InstanceStatusComplete, instanceStatusHistory[0].Status)
	assert.Equal(t, tPkg.Version, instanceStatusHistory[0].Version)
	assert.Equal(t, InstanceStatusDownloading, instanceStatusHistory[1].Status)
	assert.Equal(t, tPkg.Version, instanceStatusHistory[1].Version)
	assert.Equal(t, InstanceStatusUpdateGranted, instanceStatusHistory[2].Status)
	assert.Equal(t, tPkg.Version, instanceStatusHistory[2].Version)
}
