package helpers

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/coreos/go-omaha/omaha"
)

type Update struct {
	Version  string
	URL      string
	Filename string
	Hash     string
}

const (
	CoreRollerOmahaURL = "http://localhost:8000/omaha/"

	// Event types
	EventUpdateComplete         = 3
	EventUpdateDownloadStarted  = 13
	EventUpdateDownloadFinished = 14
	EventUpdateInstalled        = 800

	// Event results
	ResultFailed        = 0
	ResultSuccess       = 1
	ResultSuccessReboot = 2
)

var (
	ErrInvalidOmahaResponse = errors.New("invalid omaha response")
	ErrNoUpdate             = errors.New("no update available")
)

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

func EventDownloadStarted(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, EventUpdateDownloadStarted, ResultSuccess)
	_, err := doOmahaRequest(req)

	return err
}

func EventDownloadFinished(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, EventUpdateDownloadFinished, ResultSuccess)
	_, err := doOmahaRequest(req)

	return err
}

func EventUpdateSucceeded(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, EventUpdateComplete, ResultSuccessReboot)
	_, err := doOmahaRequest(req)

	return err
}

func EventUpdateFailed(instanceID, appID, groupID string) error {
	req := buildOmahaEventRequest(instanceID, appID, groupID, EventUpdateComplete, ResultFailed)
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
	app.BootId = instanceID
	app.Track = groupID
	event := app.AddEvent()
	event.Type = strconv.Itoa(eventType)
	event.Result = strconv.Itoa(eventResult)

	return req
}

func doOmahaRequest(req *omaha.Request) (*omaha.Response, error) {
	httpClient := &http.Client{}

	payload, err := xml.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Post(CoreRollerOmahaURL, "text/xml", bytes.NewReader(payload))
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
