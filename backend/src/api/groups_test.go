package api

import (
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgutz/dat.v1"
)

func TestAddGroup(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tPkg2, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp2.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tChannel2, _ := a.AddChannel(&Channel{Name: "test_channel2", Color: "yellow", ApplicationID: tApp2.ID, PackageID: dat.NullStringFrom(tPkg2.ID)})

	group := &Group{
		Name:                      "group1",
		Description:               "description",
		ApplicationID:             tApp.ID,
		ChannelID:                 dat.NullStringFrom(tChannel.ID),
		PolicyUpdatesEnabled:      true,
		PolicySafeMode:            true,
		PolicyPeriodInterval:      "15 minutes",
		PolicyMaxUpdatesPerPeriod: 2,
		PolicyUpdateTimeout:       "60 minutes",
	}
	group, err := a.AddGroup(group)
	assert.NoError(t, err)
	assert.Equal(t, true, group.PolicyUpdatesEnabled)

	groupX, err := a.GetGroup(group.ID)
	assert.Equal(t, group.Name, groupX.Name)
	assert.Equal(t, group.Description, groupX.Description)
	assert.Equal(t, group.PolicyUpdatesEnabled, groupX.PolicyUpdatesEnabled)
	assert.Equal(t, group.PolicySafeMode, groupX.PolicySafeMode)
	assert.Equal(t, group.PolicyPeriodInterval, groupX.PolicyPeriodInterval)
	assert.Equal(t, group.PolicyMaxUpdatesPerPeriod, groupX.PolicyMaxUpdatesPerPeriod)
	assert.Equal(t, group.PolicyUpdateTimeout, groupX.PolicyUpdateTimeout)
	assert.Equal(t, tApp.ID, groupX.ApplicationID)
	assert.Equal(t, dat.NullStringFrom(tChannel.ID), groupX.ChannelID)
	assert.Equal(t, tChannel.Name, groupX.Channel.Name)
	assert.Equal(t, tPkg.Version, groupX.Channel.Package.Version)

	_, err = a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel2.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	assert.Equal(t, ErrInvalidChannel, err, "Channel id used doesn't belong to the application id that this group will be bound to and it should.")
}

func TestUpdateGroup(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp1, _ := a.AddApp(&Application{Name: "test_app1", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp1.ID})
	tChannel1, _ := a.AddChannel(&Channel{Name: "test_channel1", Color: "blue", ApplicationID: tApp1.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tChannel2, _ := a.AddChannel(&Channel{Name: "test_channel2", Color: "green", ApplicationID: tApp1.ID})
	tChannel3, _ := a.AddChannel(&Channel{Name: "test_channel3", Color: "red", ApplicationID: tApp2.ID})

	group, _ := a.AddGroup(&Group{Name: "group1", ApplicationID: tApp1.ID, ChannelID: dat.NullStringFrom(tChannel1.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	group.Name = "group1_updated"
	group.PolicyUpdatesEnabled = true
	group.ChannelID = dat.NullStringFrom(tChannel2.ID)
	err := a.UpdateGroup(group)
	assert.NoError(t, err)

	groupX, _ := a.GetGroup(group.ID)
	assert.Equal(t, group.Name, groupX.Name)
	assert.Equal(t, group.PolicyPeriodInterval, groupX.PolicyPeriodInterval)
	assert.Equal(t, group.PolicyUpdatesEnabled, groupX.PolicyUpdatesEnabled)
	assert.Equal(t, tChannel2.Name, groupX.Channel.Name)

	groupX.ApplicationID = tApp2.ID
	err = a.UpdateGroup(groupX)
	assert.NoError(t, err, "Application id cannot be updated, but it won't produce an error.")

	groupX, _ = a.GetGroup(group.ID)
	assert.Equal(t, tApp1.ID, groupX.ApplicationID)

	groupX.ChannelID = dat.NullStringFrom(tChannel3.ID)
	err = a.UpdateGroup(groupX)
	assert.Equal(t, ErrInvalidChannel, err, "Channel id used doesn't belong to the application id that this group is bound to and it should.")
}

func TestDeleteGroup(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	err := a.DeleteGroup(tGroup.ID)
	assert.NoError(t, err)

	_, err = a.GetGroup(tGroup.ID)
	assert.Error(t, err, "Trying to get deleted group.")
}

func TestGetGroup(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	group, err := a.GetGroup(tGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, tGroup.Name, group.Name)
	assert.Equal(t, tApp.ID, group.ApplicationID)
	assert.Equal(t, tChannel.Name, group.Channel.Name)
	assert.Equal(t, tPkg.Version, group.Channel.Package.Version)

	_, err = a.GetGroup(uuid.NewV4().String())
	assert.Error(t, err, "Trying to get non existent group.")
}

func TestGetGroups(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup1, _ := a.AddGroup(&Group{Name: "test_group1", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	tGroup2, _ := a.AddGroup(&Group{Name: "test_group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	groups, err := a.GetGroups(tApp.ID, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(groups))
	assert.Equal(t, tGroup2.Name, groups[0].Name)
	assert.Equal(t, tGroup1.Name, groups[1].Name)
	assert.Equal(t, tChannel.Name, groups[1].Channel.Name)
	assert.Equal(t, tPkg.ID, groups[1].Channel.PackageID.String)
	assert.Equal(t, tPkg.Version, groups[1].Channel.Package.Version)
}
