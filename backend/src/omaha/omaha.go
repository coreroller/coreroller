package omaha

import (
	"encoding/xml"
	"errors"
	"io"
	"strconv"

	"api"

	omahaSpec "github.com/aquam8/go-omaha/omaha"
	"github.com/mgutz/logxi/v1"
	"github.com/satori/go.uuid"
)

var logger = log.New("omaha")

const (
	debug bool = true
)

var (
	coreosAppID, _    = uuid.FromString("e96281a6-d1af-4bde-9a0a-97b76e56dc57")
	coreosGroupAlpha  = "5b810680-e36a-4879-b98a-4f989e80b899"
	coreosGroupBeta   = "3fe10490-dd73-4b49-b72a-28ac19acfcdc"
	coreosGroupStable = "9a2deb70-37be-4026-853f-bfdd6b347bbe"

	// ErrMalformedRequest error indicates that the omaha request it has received is malformed
	ErrMalformedRequest = errors.New("omaha: request is malformed")

	// ErrMalformedResponse error indicates that the omaha response it wants to send is malformed
	ErrMalformedResponse = errors.New("omaha: response is malformed")
)

// HandleRequest ..
func HandleRequest(api *api.API, body io.Reader, bodyWriter *io.PipeWriter, ip string) error {
	defer func() {
		_ = bodyWriter.Close()
	}()

	omahaReq, err := readOmahaRequest(body)
	if err != nil {
		logger.Warn("HandleRequest problem with readOmahaRequest", "error", err.Error())
		return ErrMalformedRequest
	}

	omahaResp, err := buildOmahaResponse(api, omahaReq, ip)
	if err != nil {
		logger.Warn("HandleRequest problem with buildOmahaResponse", "error", err.Error())
		return ErrMalformedResponse
	}

	return writeXMLResponse(bodyWriter, omahaResp)
}

func getStatusMesssageFromRollerdResponse(err error) string {
	var errorMsg string
	var omahaError error
	errorMsg, omahaError = handleInvalidRequestErrors(err)
	if omahaError == nil {
		errorMsg, omahaError = handleRolloutPolicyErrors(err)
	}
	if omahaError == nil {
		logger.Warn("getStatusMesssageFromRollerdResponse", "error", err.Error())
		errorMsg = "error-failedToRetrieveUpdatePackageInfo"
	}

	return errorMsg
}

func handleInvalidRequestErrors(err error) (string, error) {
	switch {
	case err == api.ErrNoPackageFound:
		return "error-noPackageFound", err
	case err == api.ErrInvalidApplicationOrGroup:
		return "error-unknownApplicationOrGroup", err
	case err == api.ErrRegisterInstanceFailed:
		return "error-instanceRegistrationFailed", err
	}
	return "", nil
}

func handleRolloutPolicyErrors(err error) (string, error) {
	switch {
	case err == api.ErrMaxUpdatesPerPeriodLimitReached:
		return "error-maxUpdatesPerPeriodLimitReached", err
	case err == api.ErrMaxConcurrentUpdatesLimitReached:
		return "error-maxConcurrentUpdatesLimitReached", err
	case err == api.ErrMaxTimedOutUpdatesLimitReached:
		return "error-maxTimedOutUpdatesLimitReached", err
	case err == api.ErrUpdatesDisabled:
		return "error-updatesDisabled", err
	case err == api.ErrGetUpdatesStatsFailed:
		return "error-couldNotCheckUpdatesStats", err
	case err == api.ErrUpdateInProgressOnInstance:
		return "error-updateInProgressOnInstance", err
	}
	return "", nil
}

func getGroup(appID, appTrack string) string {
	appUUID, err := uuid.FromString(appID)
	if err == nil && appUUID == coreosAppID {
		switch appTrack {
		case "alpha":
			return coreosGroupAlpha
		case "beta":
			return coreosGroupBeta
		case "stable":
			return coreosGroupStable
		}
	}
	return appTrack
}

func buildOmahaResponse(a *api.API, omahaReq *omahaSpec.Request, ip string) (*omahaSpec.Response, error) {
	omahaResp := omahaSpec.NewResponse("coreroller")

	for _, reqApp := range omahaReq.Apps {

		respApp := omahaResp.AddApp(reqApp.Id)
		respApp.Status = "ok"

		// Let's add the track and version again in the response (as we got from request)
		respApp.Track = reqApp.Track
		respApp.Version = reqApp.Version

		// Get group
		group := getGroup(reqApp.Id, reqApp.Track)

		// If it has an event tag, we process it first.
		if reqApp.Events != nil {
			for _, event := range reqApp.Events {
				if err := processEvent(a, reqApp.MachineID, reqApp.Id, group, event); err != nil {
					logger.Warn("processEvent", "error", err.Error())
				}
				// Always acknowledge the event
				respEvent := respApp.AddEvent()
				respEvent.Status = "ok"
			}
		}

		// If it has an updatechek tag
		if reqApp.UpdateCheck != nil {
			appPackage, err := a.GetUpdatePackage(reqApp.MachineID, ip, reqApp.Version, reqApp.Id, group)
			if err != nil && err != api.ErrNoUpdatePackageAvailable {
				// If there is an error, we return it
				respApp.Status = getStatusMesssageFromRollerdResponse(err)
			} else {
				u := respApp.AddUpdateCheck()
				processUpdateCheck(a, appPackage, u)
			}
		}
	}

	return omahaResp, nil
}

func processUpdateCheck(a *api.API, appPackage *api.Package, u *omahaSpec.UpdateCheck) {
	// No error, it could be that appPackage is nil (no update package), or that it is the new package to return.
	// In any case, response app status must be 'ok'.

	if appPackage == nil {
		u.Status = "noupdate"
		return
	}

	// Create a manifest, but do not add it to UpdateCheck response until it's successful
	manifest := &omahaSpec.Manifest{Version: appPackage.Version}
	if err := addExtraInfo(a, appPackage, manifest); err != nil {
		u.Status = "err-internal"
		return
	}

	// On a status value of "ok", additional children can be included
	u.Status = "ok"

	addPackage(appPackage, manifest)
	u.Manifest = manifest

	u.AddUrl(appPackage.URL)
}

func addPackage(pkg *api.Package, m *omahaSpec.Manifest) {
	var filename, size, hash string
	if sizeValue, err := pkg.Size.Value(); err == nil {
		size, _ = sizeValue.(string)
	}

	if filenameValue, err := pkg.Filename.Value(); err == nil {
		filename, _ = filenameValue.(string)
	}

	if hashValue, err := pkg.Hash.Value(); err == nil {
		hash, _ = hashValue.(string)
	}

	if hash != "" || filename != "" || size != "" {
		m.AddPackage(hash, filename, size, true)
	}
}

func isPackageAContainer(packageType int) bool {
	return packageType == api.PkgTypeDocker || packageType == api.PkgTypeRocket
}

func isPackageCoreos(packageType int) bool {
	return packageType == api.PkgTypeCoreos
}

func addExtraInfo(a *api.API, pkg *api.Package, m *omahaSpec.Manifest) error {
	if isPackageAContainer(pkg.Type) {
		return addContainerExtraInfo(pkg, m)
	}

	if isPackageCoreos(pkg.Type) {
		return addCoreosExtraInfo(a, pkg, m)
	}
	return nil
}

func addContainerExtraInfo(pkg *api.Package, m *omahaSpec.Manifest) error {
	action := m.AddAction("update_app")
	// Send the application type when it's a container (for containers_updater)
	action.ChromeOSVersion = strconv.Itoa(pkg.Type)
	return nil
}

func addCoreosExtraInfo(a *api.API, pkg *api.Package, m *omahaSpec.Manifest) error {
	c, err := a.GetCoreosAction(pkg.ID)
	if err != nil {
		return err
	}
	action := m.AddAction(c.Event)
	action.ChromeOSVersion = c.ChromeOSVersion
	action.Sha256 = c.Sha256
	action.NeedsAdmin = c.NeedsAdmin
	action.IsDelta = c.IsDelta
	action.DisablePayloadBackoff = c.DisablePayloadBackoff
	action.MetadataSignatureRsa = c.MetadataSignatureRsa
	action.MetadataSize = c.MetadataSize
	action.Deadline = c.Deadline
	return nil
}

func processEvent(a *api.API, machineID string, appID string, group string, event *omahaSpec.Event) error {
	logger.Info("processEvent", "appID", appID, "group", group, "event", event.Type+"."+event.Result, "eventError", event.ErrorCode, "previousVersion", event.PreviousVersion)
	eventType, err := strconv.Atoi(event.Type)
	if err != nil {
		return err
	}
	eventResult, err := strconv.Atoi(event.Result)
	if err != nil {
		return err
	}
	if err := a.RegisterEvent(machineID, appID, group, eventType, eventResult, event.PreviousVersion, event.ErrorCode); err != nil {
		return err
	}
	return nil
}

// ReadOmahaRequest ...
func readOmahaRequest(body io.Reader) (*omahaSpec.Request, error) {
	var omahaReq *omahaSpec.Request
	decoder := xml.NewDecoder(body)
	err := decoder.Decode(&omahaReq)
	if err != nil {
		return nil, err
	}

	// Debug helper
	encodeToXMLAndPrint(omahaReq, debug)

	return omahaReq, nil
}

func writeXMLResponse(w *io.PipeWriter, v interface{}) error {
	// Debug helper
	encodeToXMLAndPrint(v, debug)

	encoder := xml.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		return ErrMalformedResponse
	}
	return nil
}

func encodeToXMLAndPrint(v interface{}, debug bool) {
	if debug {
		raw, err := xml.MarshalIndent(v, "", " ")
		if err != nil {
			_ = logger.Error(err.Error())
			return
		}
		logger.Debug("OMAHA TRACE", "XML", string(raw))
		//fmt.Printf("\n%s\n", raw)
	}
}
