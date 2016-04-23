package api

import (
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgutz/dat.v1"
)

func TestAddChannel(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tPkg2, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp2.ID})

	channel, err := a.AddChannel(&Channel{Name: "channel1", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	assert.NoError(t, err)

	channelX, err := a.GetChannel(channel.ID)
	assert.NoError(t, err)
	assert.Equal(t, "channel1", channelX.Name)
	assert.Equal(t, "blue", channelX.Color)
	assert.Equal(t, tApp.ID, channelX.ApplicationID)
	assert.Equal(t, tPkg.ID, channelX.PackageID.String)

	channel2, err := a.AddChannel(&Channel{Name: "channel2", Color: "green", ApplicationID: tApp.ID})
	assert.NoError(t, err, "A channel may not have a package associated yet.")
	assert.Equal(t, "channel2", channel2.Name)
	assert.Equal(t, "green", channel2.Color)
	assert.Equal(t, dat.NullString{}, channel2.PackageID)

	_, err = a.AddChannel(&Channel{Name: "channel3"})
	assert.Error(t, err, "App id is required")

	_, err = a.AddChannel(&Channel{Name: "channel3", ApplicationID: "invalidAppID"})
	assert.Error(t, err, "App id must be a valid uuid.")

	_, err = a.AddChannel(&Channel{Name: "channel3", ApplicationID: uuid.NewV4().String()})
	assert.Error(t, err, "App used must exist.")

	_, err = a.AddChannel(&Channel{Name: "channel3", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom("invalidPackageID")})
	assert.Error(t, err, "Package id must be a valid uuid.")

	_, err = a.AddChannel(&Channel{Name: "channel3", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(uuid.NewV4().String())})
	assert.Error(t, err, "Package used must exist.")

	_, err = a.AddChannel(&Channel{Name: "channel3", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg2.ID)})
	assert.Equal(t, ErrInvalidPackage, err, "Package used must belong to the same application as the channel.")
}

func TestUpdateChannel(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tApp2, _ := a.AddApp(&Application{Name: "test_app2", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tPkg2, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.1", ApplicationID: tApp.ID})
	tPkg3, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.2", ApplicationID: tApp2.ID})
	tPkg4, _ := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.3", ApplicationID: tApp.ID, ChannelsBlacklist: []string{tChannel.ID}})

	err := a.UpdateChannel(&Channel{ID: tChannel.ID, Name: "test_channel_updated", PackageID: dat.NullStringFrom(tPkg2.ID)})
	assert.NoError(t, err)

	channel, err := a.GetChannel(tChannel.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test_channel_updated", channel.Name)
	assert.Equal(t, "", channel.Color, "Color was set to an empty string in the last update.")
	assert.Equal(t, "12.1.1", channel.Package.Version)

	err = a.UpdateChannel(&Channel{ID: tChannel.ID, Name: "test_channel_updated", PackageID: dat.NullStringFrom(tPkg3.ID)})
	assert.Equal(t, ErrInvalidPackage, err, "Package used must belong to the same application as the channel.")

	err = a.UpdateChannel(&Channel{ID: tChannel.ID, PackageID: dat.NullStringFrom(tPkg4.ID)})
	assert.Equal(t, ErrBlacklistedChannel, err, "Package used must not have blacklisted this channel.")
}

func TestDeleteChannel(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel, err := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID})

	err = a.DeleteChannel(tChannel.ID)
	assert.NoError(t, err)

	_, err = a.GetChannel(tChannel.ID)
	assert.Error(t, err, "Trying to get deleted channel.")

	err = a.DeleteChannel("invalidChannelID")
	assert.Error(t, err, "Channel id must be a valid uuid.")
}

func TestGetChannel(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel, err := a.AddChannel(&Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID})

	channel, err := a.GetChannel(tChannel.ID)
	assert.NoError(t, err)
	assert.Equal(t, tChannel.Name, channel.Name)
	assert.Equal(t, tChannel.Color, channel.Color)
	assert.Equal(t, tApp.ID, channel.ApplicationID)

	_, err = a.GetChannel("invalidChannelID")
	assert.Error(t, err, "Channel id must be a valid uuid.")

	_, err = a.GetChannel(uuid.NewV4().String())
	assert.Error(t, err, "Channel id must exist.")
}

func TestGetChannels(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	_, err := a.AddChannel(&Channel{Name: "test_channel1", Color: "blue", ApplicationID: tApp.ID})
	_, err = a.AddChannel(&Channel{Name: "test_channel2", Color: "green", ApplicationID: tApp.ID})
	_, err = a.AddChannel(&Channel{Name: "test_channel3", Color: "red", ApplicationID: tApp.ID})

	channels, err := a.GetChannels(tApp.ID, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(channels))

	_, err = a.GetChannels("invalidAppID", 0, 0)
	assert.Error(t, err, "Add id must be a valid uuid.")

	_, err = a.GetChannels(uuid.NewV4().String(), 0, 0)
	assert.Error(t, err, "App id used must exist.")
}
