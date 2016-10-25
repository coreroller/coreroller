package omaha

import (
	"bytes"
	"encoding/xml"
	"log"
	"os"
	"testing"

	"api"

	omahaSpec "github.com/aquam8/go-omaha/omaha"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgutz/dat.v1"
)

const (
	testsDbURL string = "postgres://postgres@127.0.0.1:5432/coreroller_tests?sslmode=disable&connect_timeout=10"

	reqVersion  string = "3"
	reqPlatform string = "coreos"
	reqSp       string = "linux"
	reqArch     string = ""
)

func TestMain(m *testing.M) {
	os.Setenv("COREROLLER_DB_URL", testsDbURL)

	a, err := api.New(api.OptionInitDB)
	defer a.Close()

	if err != nil {
		log.Println("These tests require PostgreSQL running and a tests database created, please adjust testsDbUrl as needed.")
		log.Println("Default: postgres://postgres@127.0.0.1:5432/coreroller_tests?sslmode=disable")
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestInvalidRequests(t *testing.T) {
	a, _ := api.New(api.OptionInitDB)
	defer a.Close()
	h := NewHandler(a)

	tTeam, _ := a.AddTeam(&api.Team{Name: "test_team"})
	tApp, _ := a.AddApp(&api.Application{Name: "test_app", Description: "Test app", TeamID: tTeam.ID})
	tPkg, _ := a.AddPackage(&api.Package{Type: api.PkgTypeCoreos, URL: "http://sample.url/pkg", Version: "640.0.0", ApplicationID: tApp.ID})
	tChannel, _ := a.AddChannel(&api.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: dat.NullStringFrom(tPkg.ID)})
	tGroup, _ := a.AddGroup(&api.Group{Name: "test_group", ApplicationID: tApp.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	validUnregisteredIP := "127.0.0.1"
	validUnregisteredMachineID := "some-id"
	validUnverifiedAppVersion := "100.0.1"
	addPing := false
	updateCheck := true
	noEventType := ""
	noEventResult := ""
	eventPreviousVersion := ""

	omahaResp := doOmahaRequest(t, h, tApp.ID, validUnverifiedAppVersion, validUnregisteredMachineID, "invalid-track", validUnregisteredIP, addPing, updateCheck, noEventType, noEventResult, eventPreviousVersion)
	checkOmahaResponse(t, omahaResp, tApp.ID, "error-instanceRegistrationFailed")

	omahaResp = doOmahaRequest(t, h, tApp.ID, validUnverifiedAppVersion, validUnregisteredMachineID, tGroup.ID, "invalid-ip", addPing, updateCheck, noEventType, noEventResult, eventPreviousVersion)
	checkOmahaResponse(t, omahaResp, tApp.ID, "error-instanceRegistrationFailed")

	omahaResp = doOmahaRequest(t, h, "invalid-app-uuid", validUnverifiedAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, updateCheck, noEventType, noEventResult, eventPreviousVersion)
	checkOmahaResponse(t, omahaResp, "invalid-app-uuid", "error-instanceRegistrationFailed")

	omahaResp = doOmahaRequest(t, h, tApp.ID, "", validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, updateCheck, noEventType, noEventResult, eventPreviousVersion)
	checkOmahaResponse(t, omahaResp, tApp.ID, "error-instanceRegistrationFailed")
}

func TestAppNoUpdateForAppWithChannelAndPackageName(t *testing.T) {
	a, _ := api.New(api.OptionInitDB)
	defer a.Close()
	h := NewHandler(a)

	tAppCoreos, _ := a.GetApp(coreosAppID)
	tPkgCoreos640, _ := a.AddPackage(&api.Package{Type: api.PkgTypeCoreos, URL: "http://sample.url/pkg", Version: "640.0.0", ApplicationID: tAppCoreos.ID})
	tChannel, _ := a.AddChannel(&api.Channel{Name: "mychannel", Color: "white", ApplicationID: tAppCoreos.ID, PackageID: dat.NullStringFrom(tPkgCoreos640.ID)})
	tGroup, _ := a.AddGroup(&api.Group{Name: "Production", ApplicationID: tAppCoreos.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	validUnregisteredIP := "127.0.0.1"
	validUnregisteredMachineID := "65e1266d-6f54-4b87-9080-23b99ca9c12f"
	expectedAppVersion := "640.0.0"
	updateCheck := true
	addPing := true

	// Now with an error event tag, no updatecheck tag
	omahaResp := doOmahaRequest(t, h, tAppCoreos.ID, expectedAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, !addPing, !updateCheck, "3", "0", "268437959")
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaEventResponse(t, omahaResp, tAppCoreos.ID, 1)
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, !addPing)
	checkOmahaNoUpdateResponse(t, omahaResp)

	// Now updatetag, successful event, no previous version
	omahaResp = doOmahaRequest(t, h, tAppCoreos.ID, expectedAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, !addPing, updateCheck, "3", "2", "0.0.0.0")
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaEventResponse(t, omahaResp, tAppCoreos.ID, 1)
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, !addPing)
	checkOmahaUpdateResponse(t, omahaResp, expectedAppVersion, "", "", "noupdate")

	// Now updatetag, successful event, no previous version
	omahaResp = doOmahaRequest(t, h, tAppCoreos.ID, expectedAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, updateCheck, "3", "2", "")
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaEventResponse(t, omahaResp, tAppCoreos.ID, 1)
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
	checkOmahaUpdateResponse(t, omahaResp, expectedAppVersion, "", "", "noupdate")

	// Now updatetag, successful event, with previous version
	omahaResp = doOmahaRequest(t, h, tAppCoreos.ID, expectedAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, updateCheck, "3", "2", "614.0.0")
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaEventResponse(t, omahaResp, tAppCoreos.ID, 1)
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
	checkOmahaUpdateResponse(t, omahaResp, expectedAppVersion, "", "", "noupdate")

	// Now updatetag, successful event, with previous version, greater than current active version
	omahaResp = doOmahaRequest(t, h, tAppCoreos.ID, "666.0.0", validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, updateCheck, "3", "2", "614.0.0")
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaEventResponse(t, omahaResp, tAppCoreos.ID, 1)
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
	checkOmahaUpdateResponse(t, omahaResp, expectedAppVersion, "", "", "noupdate")
}

func TestAppRegistrationForAppWithChannelAndPackageName(t *testing.T) {
	a, _ := api.New(api.OptionInitDB)
	defer a.Close()
	h := NewHandler(a)

	tAppCoreos, _ := a.GetApp(coreosAppID)
	tPkgCoreos640, _ := a.AddPackage(&api.Package{Type: api.PkgTypeCoreos, URL: "http://sample.url/pkg", Version: "640.0.0", ApplicationID: tAppCoreos.ID})
	tChannel, _ := a.AddChannel(&api.Channel{Name: "mychannel", Color: "white", ApplicationID: tAppCoreos.ID, PackageID: dat.NullStringFrom(tPkgCoreos640.ID)})
	tGroup, _ := a.AddGroup(&api.Group{Name: "Production", ApplicationID: tAppCoreos.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})

	validUnregisteredIP := "127.0.0.1"
	validUnregisteredMachineID := "65e1266d-6f54-4b87-9080-23b99ca9c12f"
	expectedAppVersion := "640.0.0"
	updateCheck := true
	noEventType := ""
	noEventResult := ""
	completedEventType := "3"
	sucessEventResult := "1"
	eventPreviousVersion := ""
	addPing := true

	omahaResp := doOmahaRequest(t, h, tAppCoreos.ID, expectedAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, updateCheck, noEventType, noEventResult, eventPreviousVersion)
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
	checkOmahaUpdateResponse(t, omahaResp, expectedAppVersion, "", "", "noupdate")

	omahaResp = doOmahaRequest(t, h, tAppCoreos.ID, expectedAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, !updateCheck, completedEventType, sucessEventResult, eventPreviousVersion)
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
}

func TestAppUpdateForAppWithChannelAndPackageName(t *testing.T) {
	a, _ := api.New(api.OptionInitDB)
	defer a.Close()
	h := NewHandler(a)

	tAppCoreos, _ := a.GetApp(coreosAppID)
	tFilenameCoreos := "coreosupdate.tgz"
	tPkgCoreos640, _ := a.AddPackage(&api.Package{Type: api.PkgTypeCoreos, URL: "http://sample.url/pkg", Filename: dat.NullStringFrom(tFilenameCoreos), Version: "99640.0.0", ApplicationID: tAppCoreos.ID})
	tChannel, _ := a.AddChannel(&api.Channel{Name: "mychannel", Color: "white", ApplicationID: tAppCoreos.ID, PackageID: dat.NullStringFrom(tPkgCoreos640.ID)})
	tGroup, _ := a.AddGroup(&api.Group{Name: "Production", ApplicationID: tAppCoreos.ID, ChannelID: dat.NullStringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	coreosAction, _ := a.AddCoreosAction(&api.CoreosAction{Event: "postinstall", Sha256: "fsdkjjfghsdakjfgaksdjfasd", PackageID: tPkgCoreos640.ID})

	validUnregisteredIP := "127.0.0.1"
	validUnregisteredMachineID := "65e1266d-6f54-4b87-9080-23b99ca9c12f"
	oldAppVersion := "610.0.0"
	updateCheck := true
	addPing := true

	omahaResp := doOmahaRequest(t, h, tAppCoreos.ID, oldAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, updateCheck, "", "", "")
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaUpdateResponse(t, omahaResp, tPkgCoreos640.Version, tFilenameCoreos, tPkgCoreos640.URL, "ok")
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
	checkOmahaCoreosAction(t, coreosAction, omahaResp.Apps[0].UpdateCheck.Manifest.Actions.Actions[0])

	// Send download started
	omahaResp = doOmahaRequest(t, h, tAppCoreos.ID, oldAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, !updateCheck, "13", "1", "")
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
	checkOmahaNoUpdateResponse(t, omahaResp)

	// Send download finished
	omahaResp = doOmahaRequest(t, h, tAppCoreos.ID, oldAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, !updateCheck, "14", "1", "")
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
	checkOmahaNoUpdateResponse(t, omahaResp)

	// Send complete
	omahaResp = doOmahaRequest(t, h, tAppCoreos.ID, oldAppVersion, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, !updateCheck, "3", "1", "")
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
	checkOmahaNoUpdateResponse(t, omahaResp)

	// Send rebooted
	omahaResp = doOmahaRequest(t, h, tAppCoreos.ID, tPkgCoreos640.Version, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, updateCheck, "3", "2", oldAppVersion)
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
	checkOmahaUpdateResponse(t, omahaResp, tPkgCoreos640.Version, "", "", "noupdate")

	// Expect no update
	omahaResp = doOmahaRequest(t, h, tAppCoreos.ID, tPkgCoreos640.Version, validUnregisteredMachineID, tGroup.ID, validUnregisteredIP, addPing, updateCheck, "", "", "")
	checkOmahaResponse(t, omahaResp, tAppCoreos.ID, "ok")
	checkOmahaPingResponse(t, omahaResp, tAppCoreos.ID, addPing)
	checkOmahaUpdateResponse(t, omahaResp, tPkgCoreos640.Version, "", "", "noupdate")
}

func TestCoreosGroupNamesConversionToIds(t *testing.T) {
	a, _ := api.New(api.OptionInitDB)
	defer a.Close()
	h := NewHandler(a)

	coreosAppIDWithCurlyBraces := "{" + coreosAppID + "}"
	machineID := "65e1266d-6f54-4b87-9080-23b99ca9c12f"
	machineIP := "10.0.0.1"

	omahaResp := doOmahaRequest(t, h, coreosAppID, "2000.0.0", machineID, "invalid-group", machineIP, false, true, "", "", "")
	checkOmahaResponse(t, omahaResp, coreosAppID, "error-instanceRegistrationFailed")

	omahaResp = doOmahaRequest(t, h, coreosAppID, "2000.0.0", machineID, "alpha", machineIP, false, true, "", "", "")
	checkOmahaResponse(t, omahaResp, coreosAppID, "ok")

	omahaResp = doOmahaRequest(t, h, coreosAppIDWithCurlyBraces, "2000.0.0", machineID, "alpha", machineIP, false, true, "", "", "")
	checkOmahaResponse(t, omahaResp, coreosAppIDWithCurlyBraces, "ok")
}

func doOmahaRequest(t *testing.T, h *Handler, appID, appVersion, appMachineID, appTrack, ip string, addPing, updateCheck bool, eventType, eventResult, eventPreviousVersion string) *omahaSpec.Response {
	omahaReq := omahaSpec.NewRequest(reqVersion, reqPlatform, reqSp, reqArch)
	app := omahaReq.AddApp(appID, appVersion)
	app.MachineID = appMachineID
	app.Track = appTrack
	if updateCheck {
		app.AddUpdateCheck()
	}
	if eventType != "" {
		e := app.AddEvent()
		e.Type = eventType
		e.Result = eventResult
		e.PreviousVersion = eventPreviousVersion
	}
	if addPing {
		app.AddPing()
	}

	omahaReqXML, err := xml.Marshal(omahaReq)
	assert.NoError(t, err)

	omahaRespXML := new(bytes.Buffer)
	h.Handle(bytes.NewReader(omahaReqXML), omahaRespXML, ip)

	var omahaResp *omahaSpec.Response
	err = xml.NewDecoder(omahaRespXML).Decode(&omahaResp)
	assert.NoError(t, err)

	return omahaResp
}

func checkOmahaResponse(t *testing.T, omahaResp *omahaSpec.Response, expectedAppID, expectedError string) {
	appResp := omahaResp.Apps[0]

	assert.Equal(t, expectedError, appResp.Status)
	assert.Equal(t, expectedAppID, appResp.Id)
}

func checkOmahaNoUpdateResponse(t *testing.T, omahaResp *omahaSpec.Response) {
	appResp := omahaResp.Apps[0]

	assert.Nil(t, appResp.UpdateCheck)
}

func checkOmahaUpdateResponse(t *testing.T, omahaResp *omahaSpec.Response, expectedVersion, expectedPackageName, expectedUpdateURL, expectedError string) {
	appResp := omahaResp.Apps[0]

	assert.NotNil(t, appResp.UpdateCheck)
	assert.Equal(t, expectedError, appResp.UpdateCheck.Status)

	if appResp.UpdateCheck.Manifest != nil {
		assert.True(t, appResp.UpdateCheck.Manifest.Version >= expectedVersion)
		assert.Equal(t, expectedPackageName, appResp.UpdateCheck.Manifest.Packages.Packages[0].Name)
	}

	if appResp.UpdateCheck.Urls != nil {
		assert.Equal(t, 1, len(appResp.UpdateCheck.Urls.Urls))
		assert.Equal(t, expectedUpdateURL, appResp.UpdateCheck.Urls.Urls[0].CodeBase)
	}
}

func checkOmahaEventResponse(t *testing.T, omahaResp *omahaSpec.Response, expectedAppID string, expectedEventCount int) {
	appResp := omahaResp.Apps[0]

	assert.Equal(t, expectedAppID, appResp.Id)
	assert.Equal(t, expectedEventCount, len(appResp.Events))
	for i := 0; i < expectedEventCount; i++ {
		assert.Equal(t, "ok", appResp.Events[i].Status)
	}
}

func checkOmahaPingResponse(t *testing.T, omahaResp *omahaSpec.Response, expectedAppID string, expectedPingResponse bool) {
	appResp := omahaResp.Apps[0]

	assert.Equal(t, expectedAppID, appResp.Id)
	if expectedPingResponse {
		assert.Equal(t, "ok", appResp.Ping.Status)
		assert.NotNil(t, appResp.Ping)
	} else {
		assert.Nil(t, appResp.Ping)
	}
}

func checkOmahaCoreosAction(t *testing.T, c *api.CoreosAction, r *omahaSpec.Action) {
	assert.Equal(t, c.Event, r.Event)
	assert.Equal(t, c.Sha256, r.Sha256)
	assert.Equal(t, c.IsDelta, r.IsDelta)
	assert.Equal(t, c.Deadline, r.Deadline)
	assert.Equal(t, c.DisablePayloadBackoff, r.DisablePayloadBackoff)
	assert.Equal(t, c.ChromeOSVersion, r.ChromeOSVersion)
	assert.Equal(t, c.MetadataSize, r.MetadataSize)
	assert.Equal(t, c.NeedsAdmin, r.NeedsAdmin)
	assert.Equal(t, c.MetadataSignatureRsa, r.MetadataSignatureRsa)
}
