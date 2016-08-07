package api

import (
	"testing"

	"gopkg.in/mgutz/dat.v1"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddPackage(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel1, _ := a.AddChannel(&Channel{Name: "test_channel1", Color: "blue", ApplicationID: tApp.ID})
	tChannel2, _ := a.AddChannel(&Channel{Name: "test_channel2", Color: "green", ApplicationID: tApp.ID})

	pkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{tChannel1.ID, tChannel2.ID}})
	assert.NoError(t, err)

	pkgX, err := a.GetPackage(pkg.ID)
	assert.NoError(t, err)
	assert.Equal(t, PkgTypeOther, pkgX.Type)
	assert.Equal(t, "http://sample.url/pkg", pkgX.URL)
	assert.Equal(t, "12.1.0", pkgX.Version)
	assert.Equal(t, tApp.ID, pkgX.ApplicationID)
	assert.Contains(t, pkgX.ChannelsBlacklist, tChannel1.ID)
	assert.Contains(t, pkgX.ChannelsBlacklist, tChannel2.ID)

	_, err = a.AddPackage(&Package{URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	assert.Error(t, err, "Package type is required.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, Version: "12.1.0", ApplicationID: tApp.ID})
	assert.Error(t, err, "Package url is required.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", ApplicationID: tApp.ID})
	assert.Error(t, err, "Package version is required.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "aaa12.1.0"})
	assert.Equal(t, ErrInvalidSemver, err, "Package version must be a valid semver.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0"})
	assert.Error(t, err, "App id is required and must be a valid uuid.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{uuid.NewV4().String()}})
	assert.Error(t, err, "Blacklisted channels must be existing channels ids.")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{"invalidChannelID"}})
	assert.Error(t, err, "Blacklisted channels must be valid existing channels ids.")
}

func TestAddPackageCoreos(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	pkg := &Package{
		Type:          PkgTypeCoreos,
		URL:           "https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/766.3.0/",
		Filename:      dat.NullStringFrom("update.gz"),
		Version:       "2016.6.6",
		Size:          dat.NullStringFrom("123456"),
		Hash:          dat.NullStringFrom("sha1:blablablabla"),
		ApplicationID: coreosAppID,
		CoreosAction: &CoreosAction{
			Sha256: "sha256:blablablabla",
		},
	}
	_, err := a.AddPackage(pkg)
	assert.NoError(t, err)
	assert.Equal(t, "postinstall", pkg.CoreosAction.Event)
	assert.Equal(t, false, pkg.CoreosAction.NeedsAdmin)
	assert.Equal(t, false, pkg.CoreosAction.IsDelta)
	assert.Equal(t, true, pkg.CoreosAction.DisablePayloadBackoff)
	assert.Equal(t, "sha256:blablablabla", pkg.CoreosAction.Sha256)
}

func TestUpdatePackage(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel1, _ := a.AddChannel(&Channel{Name: "test_channel1", Color: "blue", ApplicationID: tApp.ID})
	tPkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{tChannel1.ID}})
	tChannel2, _ := a.AddChannel(&Channel{Name: "test_channel2", Color: "green", ApplicationID: tApp.ID})
	tChannel3, _ := a.AddChannel(&Channel{Name: "test_channel3", Color: "red", ApplicationID: tApp.ID})
	tChannel4, _ := a.AddChannel(&Channel{Name: "test_channel4", Color: "yellow", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})

	err = a.UpdatePackage(&Package{ID: tPkg.ID, Type: PkgTypeOther, URL: "http://sample.url/pkg_updated", Version: "12.2.0", ChannelsBlacklist: []string{tChannel2.ID, tChannel3.ID}})
	assert.NoError(t, err)

	pkg, err := a.GetPackage(tPkg.ID)
	assert.NoError(t, err)
	assert.Equal(t, "http://sample.url/pkg_updated", pkg.URL)
	assert.Equal(t, "12.2.0", pkg.Version)
	assert.NotContains(t, pkg.ChannelsBlacklist, tChannel1.ID)
	assert.Contains(t, pkg.ChannelsBlacklist, tChannel2.ID)
	assert.Contains(t, pkg.ChannelsBlacklist, tChannel3.ID)

	err = a.UpdatePackage(&Package{ID: tPkg.ID, Type: PkgTypeOther, URL: "http://sample.url/pkg_updated", Version: "12.2.0", ChannelsBlacklist: []string{tChannel4.ID}})
	assert.Equal(t, ErrBlacklistingChannel, err)

	err = a.UpdatePackage(&Package{ID: tPkg.ID, Type: PkgTypeOther, URL: "http://sample.url/pkg_updated", Version: "12.2.0", ChannelsBlacklist: nil})
	assert.NoError(t, err)
	pkg, _ = a.GetPackage(tPkg.ID)
	assert.Len(t, pkg.ChannelsBlacklist, 0)
}

func TestUpdatePackageCoreos(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	pkg := &Package{
		Type:          PkgTypeCoreos,
		URL:           "https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/766.3.0/",
		Filename:      dat.NullStringFrom("update.gz"),
		Version:       "2016.6.6",
		Size:          dat.NullStringFrom("123456"),
		Hash:          dat.NullStringFrom("sha1:blablablabla"),
		ApplicationID: coreosAppID,
	}
	_, err := a.AddPackage(pkg)
	assert.NoError(t, err)
	assert.Nil(t, pkg.CoreosAction)

	pkg.CoreosAction = &CoreosAction{
		Sha256: "sha256:blablablabla",
	}
	err = a.UpdatePackage(pkg)
	assert.NoError(t, err)
	assert.Equal(t, "postinstall", pkg.CoreosAction.Event)
	assert.Equal(t, false, pkg.CoreosAction.NeedsAdmin)
	assert.Equal(t, false, pkg.CoreosAction.IsDelta)
	assert.Equal(t, true, pkg.CoreosAction.DisablePayloadBackoff)
	assert.Equal(t, "sha256:blablablabla", pkg.CoreosAction.Sha256)

	pkg.CoreosAction.Sha256 = "sha256:bleblebleble"
	err = a.UpdatePackage(pkg)
	assert.NoError(t, err)
	assert.Equal(t, "sha256:bleblebleble", pkg.CoreosAction.Sha256)
}

func TestDeletePackage(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})

	err = a.DeletePackage(tPkg.ID)
	assert.NoError(t, err)

	_, err = a.GetPackage(tPkg.ID)
	assert.Error(t, err, "Trying to get deleted package.")

	err = a.DeletePackage("invalidPackageID")
	assert.Error(t, err, "Package id must be a valid uuid.")
}

func TestGetPackage(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tChannel, _ := a.AddChannel(&Channel{Name: "test_channel1", Color: "blue", ApplicationID: tApp.ID})
	tPkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID, ChannelsBlacklist: []string{tChannel.ID}})

	pkg, err := a.GetPackage(tPkg.ID)
	assert.NoError(t, err)
	assert.Equal(t, PkgTypeOther, pkg.Type)
	assert.Equal(t, "http://sample.url/pkg", pkg.URL)
	assert.Equal(t, "12.1.0", pkg.Version)
	assert.Equal(t, tApp.ID, pkg.ApplicationID)
	assert.Equal(t, []string{tChannel.ID}, pkg.ChannelsBlacklist)

	_, err = a.GetPackage("invalidPackageID")
	assert.Error(t, err, "Package id must be a valid uuid.")

	_, err = a.GetPackage(uuid.NewV4().String())
	assert.Error(t, err, "Package id must exist.")
}

func TestGetPackageByVersion(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})

	pkg, err := a.GetPackageByVersion(tApp.ID, tPkg.Version)
	assert.NoError(t, err)
	assert.Equal(t, PkgTypeOther, pkg.Type)
	assert.Equal(t, "http://sample.url/pkg", pkg.URL)
	assert.Equal(t, "12.1.0", pkg.Version)
	assert.Equal(t, tApp.ID, pkg.ApplicationID)

	_, err = a.GetPackageByVersion("invalidAppID", "12.1.0")
	assert.Error(t, err, "Application id must be a valid uuid.")

	_, err = a.GetPackageByVersion(uuid.NewV4().String(), "12.1.0")
	assert.Error(t, err, "Application id must exist.")

	_, err = a.GetPackageByVersion(tApp.ID, "hola")
	assert.Error(t, err, "Version must be a valid semver value.")
}

func TestGetPackages(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	_, _ = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg1", Version: "1010.5.0+2016-05-27-1832", ApplicationID: tApp.ID})
	_, _ = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg2", Version: "12.1.0", ApplicationID: tApp.ID})
	_, _ = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg3", Version: "14.1.0", ApplicationID: tApp.ID})
	_, _ = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg4", Version: "1010.6.0-blabla", ApplicationID: tApp.ID})

	pkgs, err := a.GetPackages(tApp.ID, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(pkgs))
	assert.Equal(t, "http://sample.url/pkg4", pkgs[0].URL)
	assert.Equal(t, "http://sample.url/pkg1", pkgs[1].URL)
	assert.Equal(t, "http://sample.url/pkg3", pkgs[2].URL)
	assert.Equal(t, "http://sample.url/pkg2", pkgs[3].URL)

	_, err = a.GetPackages("invalidAppID", 0, 0)
	assert.Error(t, err, "Add id must be a valid uuid.")

	_, err = a.GetPackages(uuid.NewV4().String(), 0, 0)
	assert.Error(t, err, "App id used must exist.")
}
