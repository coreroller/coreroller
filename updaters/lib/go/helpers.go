package helpers

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/coreos/go-omaha/omaha"
)

// Update represents some information about an update received from CR.
type Update struct {
	Version  string
	URL      string
	Filename string
	Hash     string
}

const (
	defaultOmahaURL = "http://localhost:8000/omaha/"

	// Event types
	eventUpdateComplete         = 3
	eventUpdateDownloadStarted  = 13
	eventUpdateDownloadFinished = 14
	eventUpdateInstalled        = 800

	// Event results
	resultFailed        = 0
	resultSuccess       = 1
	resultSuccessReboot = 2
)

var (
	// ErrInvalidOmahaResponse error indicates that the omaha response received
	// from CR was invalid.
	ErrInvalidOmahaResponse = errors.New("invalid omaha response")

	// ErrNoUpdate error indicates that there wasn't any update available for
	// instance requesting it in the context of the appID/groupID provided.
	ErrNoUpdate = errors.New("no update available")
)

// GetUpdate asks CR for an update for the given instance in the context of the
// application and group provided.
func GetUpdate(instanceID, appID, groupID, version string) (*Update, error) {
	req := buildOmahaUpdateRequest(instanceID, appID, groupID, version)
	resp, err := doOmahaRequest(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.Apps) != 1 {
		return nil, ErrInvalidOmahaResponse
	}

	app := resp.Apps[0]
	if app.Status != "ok" {
		return nil, errors.New(app.Status)
	}

	status := app.UpdateCheck.Status
	switch status {
	case "ok":
		update := &Update{
			Version:  app.UpdateCheck.Manifest.Version,
			URL:      app.UpdateCheck.Urls.Urls[0].CodeBase,
			Filename: app.UpdateCheck.Manifest.Packages.Packages[0].Name,
			Hash:     app.UpdateCheck.Manifest.Packages.Packages[0].Hash,
		}
		return update, nil
	case "noupdate":
		return nil, ErrNoUpdate
	default:
		return nil, errors.New(status)
	}
}

// EventDownloadStarted posts an event to CR to indicate that the download of
// the update has started.
func EventDownloadStarted(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, eventUpdateDownloadStarted, resultSuccess)
	_, err := doOmahaRequest(req)

	return err
}

// EventDownloadFinished posts an event to CR to indicate that the download of
// the update has finished.
func EventDownloadFinished(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, eventUpdateDownloadFinished, resultSuccess)
	_, err := doOmahaRequest(req)

	return err
}

// EventUpdateSucceeded posts an event to CR to indicate that the update was
// installed successfully and the new version is working fine.
func EventUpdateSucceeded(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, eventUpdateComplete, resultSuccessReboot)
	_, err := doOmahaRequest(req)

	return err
}

// EventUpdateFailed posts an event to CR to indicate that the update process
// complete but it didn't succeed.
func EventUpdateFailed(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, eventUpdateComplete, resultFailed)
	_, err := doOmahaRequest(req)

	return err
}

func buildOmahaUpdateRequest(instanceID, appID, groupID, version string) *omaha.Request {
	req := &omaha.Request{}
	app := req.AddApp(appID, version)
	app.MachineID = instanceID
	app.BootId = instanceID
	app.Track = groupID
	app.AddUpdateCheck()

	return req
}

func buildOmahaEventRequest(instanceID, appID, groupID string, eventType, eventResult int) *omaha.Request {
	req := &omaha.Request{}
	app := req.AddApp(appID, "")
	app.MachineID = instanceID
	app.BootId = instanceID
	app.Track = groupID
	event := app.AddEvent()
	event.Type = strconv.Itoa(eventType)
	event.Result = strconv.Itoa(eventResult)

	return req
}

func doOmahaRequest(req *omaha.Request) (*omaha.Response, error) {
	omahaURL := os.Getenv("CR_OMAHA_URL")
	if omahaURL == "" {
		omahaURL = defaultOmahaURL
	}

	httpClient := &http.Client{}

	payload, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Post(omahaURL, "text/xml", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	oresp := &omaha.Response{}
	if err = xml.Unmarshal(body, oresp); err != nil {
		return nil, err
	}

	return oresp, nil
}
