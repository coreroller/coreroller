package helpers

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

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
		return nil, errors.New(string(app.Status))
	}

	status := app.UpdateCheck.Status
	switch status {
	case "ok":
		update := &Update{
			Version:  app.UpdateCheck.Manifest.Version,
			URL:      app.UpdateCheck.URLs[0].CodeBase,
			Filename: app.UpdateCheck.Manifest.Packages[0].Name,
			Hash:     app.UpdateCheck.Manifest.Packages[0].SHA1,
		}
		return update, nil
	case "noupdate":
		return nil, ErrNoUpdate
	default:
		return nil, errors.New(string(status))
	}
}

// EventDownloadStarted posts an event to CR to indicate that the download of
// the update has started.
func EventDownloadStarted(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, omaha.EventTypeUpdateDownloadStarted, omaha.EventResultSuccess)
	_, err := doOmahaRequest(req)

	return err
}

// EventDownloadFinished posts an event to CR to indicate that the download of
// the update has finished.
func EventDownloadFinished(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, omaha.EventTypeUpdateDownloadFinished, omaha.EventResultSuccess)
	_, err := doOmahaRequest(req)

	return err
}

// EventUpdateSucceeded posts an event to CR to indicate that the update was
// installed successfully and the new version is working fine.
func EventUpdateSucceeded(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, omaha.EventTypeUpdateComplete, omaha.EventResultSuccessReboot)
	_, err := doOmahaRequest(req)

	return err
}

// EventUpdateFailed posts an event to CR to indicate that the update process
// complete but it didn't succeed.
func EventUpdateFailed(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, omaha.EventTypeUpdateComplete, omaha.EventResultError)
	_, err := doOmahaRequest(req)

	return err
}

func buildOmahaUpdateRequest(instanceID, appID, groupID, version string) *omaha.Request {
	req := &omaha.Request{}
	app := req.AddApp(appID, version)
	app.MachineID = instanceID
	app.BootID = instanceID
	app.Track = groupID
	app.AddUpdateCheck()

	return req
}

func buildOmahaEventRequest(instanceID, appID, groupID string, eventType omaha.EventType, eventResult omaha.EventResult) *omaha.Request {
	req := &omaha.Request{}
	app := req.AddApp(appID, "")
	app.MachineID = instanceID
	app.BootID = instanceID
	app.Track = groupID
	event := app.AddEvent()
	event.Type = eventType
	event.Result = eventResult

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
