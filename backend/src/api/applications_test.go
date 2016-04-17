package api

import (
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgutz/dat.v1"
)

func TestAddApp(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})

	newApp, err := a.AddApp(&Application{Name: "app1", TeamID: tTeam.ID})
	assert.NoError(t, err)

	newAppX, err := a.GetApp(newApp.ID)
	assert.NoError(t, err)
	assert.Equal(t, "app1", newAppX.Name)

	_, err = a.AddApp(&Application{Name: "app1", TeamID: tTeam.ID})
	assert.Error(t, err, "App name must be unique per team.")

	_, err = a.AddApp(&Application{TeamID: tTeam.ID})
	assert.Error(t, err, "App name is required.")

	_, err = a.AddApp(&Application{Name: "app2"})
	assert.Error(t, err, "Team id is required.")

	_, err = a.AddApp(&Application{Name: "app2", TeamID: uuid.NewV4().String()})
	assert.Error(t, err, "Team id used must exist.")

	_, err = a.AddApp(&Application{Name: "app2", TeamID: "invalidTeamID"})
	assert.Error(t, err, "Team id must be a valid uuid.")
}

func TestAddAppCloning(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	_, _ = a.AddGroup(&Group{Name: "group1", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	_, _ = a.AddGroup(&Group{Name: "group2", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	clonedApp, err := a.AddAppCloning(&Application{Name: "app1", TeamID: tTeam.ID}, tApp.ID)
	assert.NoError(t, err)

	sourceApp, _ := a.GetApp(tApp.ID)
	clonedAppX, _ := a.GetApp(clonedApp.ID)
	assert.Equal(t, len(sourceApp.Groups), len(clonedAppX.Groups))
	assert.Equal(t, len(sourceApp.Channels), len(clonedAppX.Channels))

	// TODO: test specific fields in groups and channels (do not forget channel id in group!)

	_, err = a.AddAppCloning(&Application{Name: "app2", TeamID: tTeam.ID}, "")
	assert.NoError(t, err, "Using an empty source app id when cloning has the same effect as not cloning.")
}

func TestUpdateApp(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", Description: "description", TeamID: tTeam.ID})

	err := a.UpdateApp(&Application{ID: tApp.ID, Name: "test_app_updated"})
	assert.NoError(t, err)

	app, _ := a.GetApp(tApp.ID)
	assert.Equal(t, "test_app_updated", app.Name)
	assert.Equal(t, "", app.Description, "Description set to empty string in last update as it wasn't provided")

	err = a.UpdateApp(&Application{ID: tApp.ID, Name: "test_app", Description: "description_updated"})
	assert.NoError(t, err)

	app, _ = a.GetApp(tApp.ID)
	assert.Equal(t, "test_app", app.Name)
	assert.Equal(t, "description_updated", app.Description)

	err = a.UpdateApp(&Application{Name: "test_app_updated_again"})
	assert.Error(t, err, "App id is required.")

	err = a.UpdateApp(&Application{ID: "invalidAppID", Name: "test_app_updated_again"})
	assert.Error(t, err, "App id must be a valid uuid.")
}

func TestDeleteApp(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})

	err := a.DeleteApp(tApp.ID)
	assert.NoError(t, err)

	_, err = a.GetApp(tApp.ID)
	assert.Error(t, err, "Trying to get deleted app.")
}

func TestGetApp(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID})

	app, err := a.GetApp(tApp.ID)
	assert.NoError(t, err)
	assert.Equal(t, tApp.Name, app.Name)
	assert.Equal(t, tChannel.Name, app.Channels[0].Name)

	_, err = a.GetApp(uuid.NewV4().String())
	assert.Error(t, err, "Trying to get non existent app.")
}

func TestGetApps(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp1, _ := a.AddApp(&Application{Name: "test_app1", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp1.ID})

	apps, err := a.GetApps(tTeam.ID, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(apps))
	assert.Equal(t, tApp1.Name, apps[1].Name)
	assert.Equal(t, tApp2.Name, apps[0].Name)
	assert.Equal(t, tChannel.Name, apps[1].Channels[0].Name)

	_, err = a.GetApps(uuid.NewV4().String(), 0, 0)
	assert.Error(t, err, "Trying to get apps of inexisting team.")
}
