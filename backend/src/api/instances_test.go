package api

import (
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgutz/dat.v1"
)

func TestRegisterInstance(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := a.AddGroup(&Group{Name: "group2", ApplicationID: tApp2.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup3, _ := a.AddGroup(&Group{Name: "group3", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	instanceID := uuid.NewV4().String()

	_, err := a.RegisterInstance("", "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	assert.Error(t, err, "Using empty string as instance id.")

	_, err = a.RegisterInstance(instanceID, "invalidIP", "1.0.0", tApp.ID, tGroup.ID)
	assert.Error(t, err, "Using an invalid instance ip.")

	_, err = a.RegisterInstance(instanceID, "10.0.0.1", "1.0.0", "invalidAppID", tGroup.ID)
	assert.Error(t, err, "Using an invalid application id.")

	_, err = a.RegisterInstance(instanceID, "10.0.0.1", "1.0.0", tApp.ID, "invalidGroupID")
	assert.Error(t, err, "Using an invalid group id.")

	_, err = a.RegisterInstance(instanceID, "10.0.0.1", "", tApp.ID, "invalidGroupID")
	assert.Error(t, err, "Using an empty instance version.")

	_, err = a.RegisterInstance(instanceID, "10.0.0.1", "aaa1.0.0", tApp.ID, "invalidGroupID")
	assert.Equal(t, ErrInvalidSemver, err, "Using an invalid instance version.")

	_, err = a.RegisterInstance(instanceID, "10.0.0.1", "1.0.0", tApp.ID, tGroup2.ID)
	assert.Equal(t, ErrInvalidApplicationOrGroup, err, "The group provided doesn't belong to the application provided.")

	instance, err := a.RegisterInstance(instanceID, "10.0.0.1", "1.0.0", "{"+tApp.ID+"}", "{"+tGroup.ID+"}")
	assert.NoError(t, err)
	assert.Equal(t, instanceID, instance.ID)
	assert.Equal(t, "10.0.0.1", instance.IP)

	instance, err = a.RegisterInstance(instanceID, "10.0.0.2", "1.0.2", tApp.ID, tGroup.ID)
	assert.NoError(t, err, "Registering an already registered instance with some updates, that's fine.")
	assert.Equal(t, "10.0.0.2", instance.IP)
	assert.Equal(t, "1.0.2", instance.Application.Version)

	_, err = a.RegisterInstance(instanceID, "10.0.0.2", "1.0.2", tApp2.ID, tGroup.ID)
	assert.Error(t, err, "Application id cannot be updated.")

	instance, err = a.RegisterInstance(instanceID, "10.0.0.3", "1.0.3", tApp.ID, tGroup3.ID)
	assert.NoError(t, err, "Registering an already registered instance using a different group, that's fine.")
	assert.Equal(t, "10.0.0.3", instance.IP)
	assert.Equal(t, "1.0.3", instance.Application.Version)
	assert.Equal(t, dat.NullStringFrom(tGroup3.ID), instance.Application.GroupID)
}

func TestGetInstance(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)

	_, err := a.GetInstance(uuid.NewV4().String(), tApp.ID)
	assert.Error(t, err, "Using non existent instance id.")

	_, err = a.GetInstance("invalidInstanceID", tApp.ID)
	assert.Error(t, err, "Using invalid instance id.")

	_, err = a.GetInstance(tInstance.ID, "invalidApplicationID")
	assert.Error(t, err, "Using invalid application id.")

	instance, err := a.GetInstance(tInstance.ID, tApp.ID)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.1", instance.IP)
	assert.Equal(t, tApp.ID, instance.Application.ApplicationID)
	assert.Equal(t, dat.NullStringFrom(tGroup.ID), instance.Application.GroupID)
	assert.Equal(t, "1.0.0", instance.Application.Version)
}

func TestGetInstances(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := a.AddGroup(&Group{Name: "group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tInstance, _ := a.RegisterInstance(uuid.NewV4().String(), "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	_, _ = a.RegisterInstance(uuid.NewV4().String(), "10.0.0.2", "1.0.1", tApp.ID, tGroup.ID)
	_, _ = a.RegisterInstance(uuid.NewV4().String(), "10.0.0.3", "1.0.2", tApp.ID, tGroup2.ID)

	instances, err := a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Version: "1.0.0", Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(instances))

	instances, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(instances))

	instances, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Page: 1, PerPage: 1})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(instances))

	instances, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup2.ID, Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(instances))

	_, _ = a.GetUpdatePackage(tInstance.ID, "10.0.0.1", "1.0.0", tApp.ID, tGroup.ID)
	_ = a.RegisterEvent(tInstance.ID, tApp.ID, tGroup.ID, EventUpdateComplete, ResultSuccessReboot, "", "")

	instances, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, GroupID: tGroup.ID, Status: InstanceStatusComplete, Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(instances))

	_, err = a.GetInstances(InstancesQueryParams{GroupID: tGroup.ID, Version: "1.0.0", Page: 1, PerPage: 10})
	assert.Error(t, err, "Application id must be provided.")

	_, err = a.GetInstances(InstancesQueryParams{ApplicationID: tApp.ID, Version: "1.0.0", Page: 1, PerPage: 10})
	assert.Error(t, err, "Group id must be provided.")

	_, err = a.GetInstances(InstancesQueryParams{Version: "1.0.0", Page: 1, PerPage: 10})
	assert.Error(t, err, "Application id and group id are required and must be valid uuids.")

	_, err = a.GetInstances(InstancesQueryParams{ApplicationID: "invalidApplicationID", GroupID: "invalidGroupID", Version: "1.0.0", Page: 1, PerPage: 10})
	assert.Error(t, err, "Application id and group id are required and must be valid uuids.")
}
