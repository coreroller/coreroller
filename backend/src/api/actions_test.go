package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddCoreosAction(t *testing.T) {
	a, _ := New(OptionInitDB)
	defer a.Close()

	tTeam, _ := a.AddTeam(&Team{Name: "test_team"})
	tApp, _ := a.AddApp(&Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&Package{Type: PkgTypeCoreos, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})

	coreosAction, err := a.AddCoreosAction(&CoreosAction{Event: "postinstall", Sha256: "fsdkjjfghsdakjfgaksdjfasd", PackageID: tPkg.ID})
	assert.NoError(t, err)

	coreosActionX, err := a.GetCoreosAction(tPkg.ID)
	assert.NoError(t, err)

	assert.Equal(t, coreosAction.Event, coreosActionX.Event)
	assert.Equal(t, coreosAction.Sha256, coreosActionX.Sha256)
}
