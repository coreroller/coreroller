package api

import (
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddPackage(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})

	pkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	assert.NoError(t, err)

	pkgX, err := a.GetPackage(pkg.ID)
	assert.NoError(t, err)
	assert.Equal(t, PkgTypeOther, pkgX.Type)
	assert.Equal(t, "http://sample.url/pkg", pkgX.URL)
	assert.Equal(t, "12.1.0", pkgX.Version)
	assert.Equal(t, tApp.ID, pkgX.ApplicationID)

	_, err = a.AddPackage(&Package{URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	assert.Error(t, err, "Package type is required")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, Version: "12.1.0", ApplicationID: tApp.ID})
	assert.Error(t, err, "Package url is required")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", ApplicationID: tApp.ID})
	assert.Error(t, err, "Package version is required")

	_, err = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0"})
	assert.Error(t, err, "App id is required and must be a valid uuid.")
}

func TestUpdatePackage(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})

	err = a.UpdatePackage(&Package{ID: tPkg.ID, Type: PkgTypeOther, URL: "http://sample.url/pkg_updated", Version: "12.2.0"})
	assert.NoError(t, err)

	pkg, err := a.GetPackage(tPkg.ID)
	assert.NoError(t, err)
	assert.Equal(t, "http://sample.url/pkg_updated", pkg.URL)
	assert.Equal(t, "12.2.0", pkg.Version)
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
	tPkg, err := a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})

	pkg, err := a.GetPackage(tPkg.ID)
	assert.NoError(t, err)
	assert.Equal(t, PkgTypeOther, pkg.Type)
	assert.Equal(t, "http://sample.url/pkg", pkg.URL)
	assert.Equal(t, "12.1.0", pkg.Version)
	assert.Equal(t, tApp.ID, pkg.ApplicationID)

	_, err = a.GetPackage("invalidPackageID")
	assert.Error(t, err, "Package id must be a valid uuid.")

	_, err = a.GetPackage(uuid.NewV4().String())
	assert.Error(t, err, "Package id must exist.")
}

func TestGetPackages(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	_, _ = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg1", Version: "12.1.0", ApplicationID: tApp.ID})
	_, _ = a.AddPackage(&Package{Type: PkgTypeOther, URL: "http://sample.url/pkg2", Version: "14.1.0", ApplicationID: tApp.ID})

	pkgs, err := a.GetPackages(tApp.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(pkgs))
	assert.Equal(t, "http://sample.url/pkg2", pkgs[0].URL)
	assert.Equal(t, "http://sample.url/pkg1", pkgs[1].URL)

	_, err = a.GetPackages("invalidAppID")
	assert.Error(t, err, "Add id must be a valid uuid.")

	_, err = a.GetPackages(uuid.NewV4().String())
	assert.Error(t, err, "App id used must exist.")
}
