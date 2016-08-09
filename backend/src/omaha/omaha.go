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

var (
	logger = log.New("omaha")

	coreosAppID = api.CoreosAppID()

	// ErrMalformedRequest error indicates that the omaha request it has
	// received is malformed.
	ErrMalformedRequest = errors.New("omaha: request is malformed")

	// ErrMalformedResponse error indicates that the omaha response it wants to
	// send is malformed.
	ErrMalformedResponse = errors.New("omaha: response is malformed")
)

// Handler represents a component capable of processing Omaha requests. It uses
// the CoreRoller API to get packages updates, process events, etc.
type Handler struct {
	crApi *api.API
}

// NewHandler creates a new Handler instance.
func NewHandler(crApi *api.API) *Handler {
	return &Handler{
		crApi: crApi,
	}
}

// Handle is in charge of processing an Omaha request.
func (h *Handler) Handle(rawReq io.Reader, respWriter io.Writer, ip string) error {
	var omahaReq *omahaSpec.Request

	if err := xml.NewDecoder(rawReq).Decode(&omahaReq); err != nil {
		logger.Warn("Handle - malformed omaha request", "error", err.Error())
		return ErrMalformedRequest
	}
	trace(omahaReq)

	omahaResp, err := h.buildOmahaResponse(omahaReq, ip)
	if err != nil {
		logger.Warn("Handle - error building omaha response", "error", err.Error())
		return ErrMalformedResponse
	}
	trace(omahaResp)

	return xml.NewEncoder(respWriter).Encode(omahaResp)
}

func (h *Handler) buildOmahaResponse(omahaReq *omahaSpec.Request, ip string) (*omahaSpec.Response, error) {
	omahaResp := omahaSpec.NewResponse("coreroller")

	for _, reqApp := range omahaReq.Apps {
		respApp := omahaResp.AddApp(reqApp.Id)
		respApp.Status = "ok"
		respApp.Track = reqApp.Track
		respApp.Version = reqApp.Version

		// Use Track field as the group to ask CR for updates. For the CoreOS
		// app, map group name to its id if available.
		group := reqApp.Track
		if reqAppUUID, err := uuid.FromString(reqApp.Id); err == nil {
			if reqAppUUID.String() == coreosAppID {
				if coreosGroupID, ok := api.CoreosGroupID(group); ok {
					group = coreosGroupID
				}
			}
		}

		if reqApp.Events != nil {
			for _, event := range reqApp.Events {
				if err := h.processEvent(reqApp.MachineID, reqApp.Id, group, event); err != nil {
					logger.Warn("processEvent", "error", err.Error())
				}
				respEvent := respApp.AddEvent()
				respEvent.Status = "ok"
			}
		}

		if reqApp.UpdateCheck != nil {
			pkg, err := h.crApi.GetUpdatePackage(reqApp.MachineID, ip, reqApp.Version, reqApp.Id, group)
			if err != nil && err != api.ErrNoUpdatePackageAvailable {
				respApp.Status = h.getStatusMessage(err)
			} else {
				respApp.UpdateCheck = h.prepareUpdateCheck(pkg)
			}
		}
	}

	return omahaResp, nil
}

func (h *Handler) processEvent(machineID string, appID string, group string, event *omahaSpec.Event) error {
	logger.Info("processEvent", "appID", appID, "group", group, "event", event.Type+"."+event.Result, "eventError", event.ErrorCode, "previousVersion", event.PreviousVersion)

	eventType, err := strconv.Atoi(event.Type)
	if err != nil {
		return err
	}
	eventResult, err := strconv.Atoi(event.Result)
	if err != nil {
		return err
	}

	return h.crApi.RegisterEvent(machineID, appID, group, eventType, eventResult, event.PreviousVersion, event.ErrorCode)
}

func (h *Handler) getStatusMessage(crErr error) string {
	switch crErr {
	case api.ErrNoPackageFound:
		return "error-noPackageFound"
	case api.ErrInvalidApplicationOrGroup:
		return "error-unknownApplicationOrGroup"
	case api.ErrRegisterInstanceFailed:
		return "error-instanceRegistrationFailed"
	case api.ErrMaxUpdatesPerPeriodLimitReached:
		return "error-maxUpdatesPerPeriodLimitReached"
	case api.ErrMaxConcurrentUpdatesLimitReached:
		return "error-maxConcurrentUpdatesLimitReached"
	case api.ErrMaxTimedOutUpdatesLimitReached:
		return "error-maxTimedOutUpdatesLimitReached"
	case api.ErrUpdatesDisabled:
		return "error-updatesDisabled"
	case api.ErrGetUpdatesStatsFailed:
		return "error-couldNotCheckUpdatesStats"
	case api.ErrUpdateInProgressOnInstance:
		return "error-updateInProgressOnInstance"
	}

	logger.Warn("getStatusMessage", "error", crErr.Error())

	return "error-failedToRetrieveUpdatePackageInfo"
}

func (h *Handler) prepareUpdateCheck(pkg *api.Package) *omahaSpec.UpdateCheck {
	updateCheck := &omahaSpec.UpdateCheck{}

	if pkg == nil {
		updateCheck.Status = "noupdate"
		return updateCheck
	}

	// Create a manifest, but do not add it to UpdateCheck until it's successful
	manifest := &omahaSpec.Manifest{Version: pkg.Version}
	manifest.AddPackage(pkg.Hash.String, pkg.Filename.String, pkg.Size.String, true)

	switch pkg.Type {
	case api.PkgTypeCoreos:
		cra, err := h.crApi.GetCoreosAction(pkg.ID)
		if err != nil {
			updateCheck.Status = "err-internal"
			return updateCheck
		}
		a := manifest.AddAction(cra.Event)
		a.ChromeOSVersion = cra.ChromeOSVersion
		a.Sha256 = cra.Sha256
		a.NeedsAdmin = cra.NeedsAdmin
		a.IsDelta = cra.IsDelta
		a.DisablePayloadBackoff = cra.DisablePayloadBackoff
		a.MetadataSignatureRsa = cra.MetadataSignatureRsa
		a.MetadataSize = cra.MetadataSize
		a.Deadline = cra.Deadline
	}

	updateCheck.Status = "ok"
	updateCheck.Manifest = manifest
	updateCheck.AddUrl(pkg.URL)

	return updateCheck
}

func trace(v interface{}) {
	if logger.IsDebug() {
		raw, err := xml.MarshalIndent(v, "", " ")
		if err != nil {
			_ = logger.Error(err.Error())
			return
		}
		logger.Debug("Omaha trace", "XML", string(raw))
	}
}
