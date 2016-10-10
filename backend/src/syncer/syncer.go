package syncer

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"api"

	"github.com/aquam8/go-omaha/omaha"
	"github.com/mgutz/logxi/v1"
	"github.com/satori/go.uuid"
	"gopkg.in/mgutz/dat.v1"
)

const (
	coreosUpdatesURL = "https://public.update.core-os.net/v1/update/"
	coreosAppID      = "{e96281a6-d1af-4bde-9a0a-97b76e56dc57}"
	checkFrequency   = 1 * time.Hour
)

var (
	logger = log.New("syncer")

	// ErrInvalidAPIInstance error indicates that no valid api instance was
	// provided to the syncer constructor.
	ErrInvalidAPIInstance = errors.New("sync: invalid api instance")
)

// Syncer represents a process in charge of checking for updates in the
// different official CoreOS channels and updating the CoreOS application in
// CoreRoller as needed (creating new packages and updating channels to point
// to them).
type Syncer struct {
	api         *api.API
	stopCh      chan struct{}
	machinesIDs map[string]string
	bootIDs     map[string]string
	versions    map[string]string
	channelsIDs map[string]string
	httpClient  *http.Client
}

// New creates a new Syncer instance.
func New(api *api.API) (*Syncer, error) {
	if api == nil {
		return nil, ErrInvalidAPIInstance
	}

	s := &Syncer{
		api:         api,
		stopCh:      make(chan struct{}),
		machinesIDs: make(map[string]string, 3),
		bootIDs:     make(map[string]string, 3),
		channelsIDs: make(map[string]string, 3),
		versions:    make(map[string]string, 3),
		httpClient:  &http.Client{},
	}

	if err := s.initialize(); err != nil {
		return nil, err
	}

	return s, nil
}

// Start makes the syncer start working. It will check for updates every
// checkFrequency until it's asked to stop.
func (s *Syncer) Start() {
	logger.Debug("syncer ready!")
	checkCh := time.Tick(checkFrequency)
	startup_checkCh := time.After(1 * time.Minute)

L:
	for {
		select {
		case <-startup_checkCh:
			_ = s.checkForUpdates()
		case <-checkCh:
			_ = s.checkForUpdates()
		case <-s.stopCh:
			break L
		}
	}

	s.api.Close()
}

// Stop stops the polling for updates.
func (s *Syncer) Stop() {
	logger.Debug("stopping syncer..")
	s.stopCh <- struct{}{}
}

// initialize does some initial setup to prepare the syncer, checking in
// CoreRoller the last versions we know about for the different channels in the
// CoreOS application and keeping track of some ids.
func (s *Syncer) initialize() error {
	coreosApp, err := s.api.GetApp(coreosAppID)
	if err != nil {
		return err
	}

	for _, c := range coreosApp.Channels {
		if c.Name == "stable" || c.Name == "beta" || c.Name == "alpha" {
			s.machinesIDs[c.Name] = "{" + uuid.NewV4().String() + "}"
			s.bootIDs[c.Name] = "{" + uuid.NewV4().String() + "}"
			s.channelsIDs[c.Name] = c.ID

			if c.Package != nil {
				s.versions[c.Name] = c.Package.Version
			} else {
				s.versions[c.Name] = "766.0.0"
			}
		}
	}

	return nil
}

// checkForUpdates polls the public CoreOS servers looking for updates in the
// official channels (stable, beta, alpha) sending Omaha requests. When an
// update is received we'll process it, creating packages and updating channels
// in CoreRoller as needed.
func (s *Syncer) checkForUpdates() error {
	for channel, currentVersion := range s.versions {
		logger.Debug("checking for updates", "channel", channel, "currentVersion", currentVersion)

		update, err := s.doOmahaRequest(channel, currentVersion)
		if err != nil {
			return err
		}
		if update.Status == "ok" {
			logger.Debug("checkForUpdates, got an update", "channel", channel, "currentVersion", currentVersion, "availableVersion", update.Manifest.Version)
			if err := s.processUpdate(channel, update); err != nil {
				return err
			}
			s.versions[channel] = update.Manifest.Version
			s.bootIDs[channel] = "{" + uuid.NewV4().String() + "}"
		} else {
			logger.Debug("checkForUpdates, no update available", "channel", channel, "currentVersion", currentVersion, "updateStatus", update.Status)
		}

		select {
		case <-time.After(1 * time.Minute):
		case <-s.stopCh:
			break
		}
	}

	return nil
}

// doOmahaRequest sends an Omaha request checking if there is an update for a
// specific CoreOS channel, returning the update check to the caller.
func (s *Syncer) doOmahaRequest(channel, currentVersion string) (*omaha.UpdateCheck, error) {
	req := omaha.NewRequest("Chateau", "CoreOS", currentVersion+"_x86_64", "")
	req.Version = "CoreOSUpdateEngine-0.1.0.0"
	req.UpdaterVersion = "CoreOSUpdateEngine-0.1.0.0"
	req.InstallSource = "scheduler"
	req.IsMachine = "1"
	app := req.AddApp(coreosAppID, currentVersion)
	app.AddUpdateCheck()
	app.MachineID = s.machinesIDs[channel]
	app.BootId = s.bootIDs[channel]
	app.Track = channel

	payload, err := xml.Marshal(req)
	if err != nil {
		logger.Error("checkForUpdates, marshalling request xml", "error", err)
		return nil, err
	}
	logger.Debug("doOmahaRequest", "request", string(payload))

	resp, err := s.httpClient.Post(coreosUpdatesURL, "text/xml", bytes.NewReader(payload))
	if err != nil {
		logger.Error("checkForUpdates, posting omaha response", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("checkForUpdates, reading omaha response", "error", err)
		return nil, err

	}
	logger.Debug("doOmahaRequest", "response", string(body))

	oresp := &omaha.Response{}
	err = xml.Unmarshal(body, oresp)
	if err != nil {
		logger.Error("checkForUpdates, unmarshalling omaha response", "error", err)
		return nil, err
	}

	return oresp.Apps[0].UpdateCheck, nil
}

// processUpdate is in charge of creating packages in the CoreOS application in
// CoreRoller and updating the appropriate channel to point to the new channel.
func (s *Syncer) processUpdate(channelName string, update *omaha.UpdateCheck) error {
	// Create new package and action for CoreOS application in CoreRoller if
	// needed (package may already exist and we just need to update the channel
	// reference to it
	pkg, err := s.api.GetPackageByVersion(coreosAppID, update.Manifest.Version)
	if err != nil {
		pkg = &api.Package{
			Type:          api.PkgTypeCoreos,
			URL:           update.Urls.Urls[0].CodeBase,
			Version:       update.Manifest.Version,
			Filename:      dat.NullStringFrom(update.Manifest.Packages.Packages[0].Name),
			Size:          dat.NullStringFrom(update.Manifest.Packages.Packages[0].Size),
			Hash:          dat.NullStringFrom(update.Manifest.Packages.Packages[0].Hash),
			ApplicationID: coreosAppID,
		}
		if _, err = s.api.AddPackage(pkg); err != nil {
			logger.Error("processUpdate, adding package", "error", err, "channelName", channelName)
			return err
		}

		coreosAction := &api.CoreosAction{
			Event:                 update.Manifest.Actions.Actions[0].Event,
			ChromeOSVersion:       update.Manifest.Actions.Actions[0].ChromeOSVersion,
			Sha256:                update.Manifest.Actions.Actions[0].Sha256,
			NeedsAdmin:            update.Manifest.Actions.Actions[0].NeedsAdmin,
			IsDelta:               update.Manifest.Actions.Actions[0].IsDelta,
			DisablePayloadBackoff: update.Manifest.Actions.Actions[0].DisablePayloadBackoff,
			MetadataSignatureRsa:  update.Manifest.Actions.Actions[0].MetadataSignatureRsa,
			MetadataSize:          update.Manifest.Actions.Actions[0].MetadataSize,
			Deadline:              update.Manifest.Actions.Actions[0].Deadline,
			PackageID:             pkg.ID,
		}
		if _, err = s.api.AddCoreosAction(coreosAction); err != nil {
			logger.Error("processUpdate, adding coreos action", "error", err, "channelName", channelName)
			return err
		}
	}

	// Update channel to point to the package with the new version
	channel, err := s.api.GetChannel(s.channelsIDs[channelName])
	if err != nil {
		logger.Error("processUpdate, getting channel to update", "error", err, "channelName", channelName)
		return err
	}
	channel.PackageID = dat.NullStringFrom(pkg.ID)
	if err = s.api.UpdateChannel(channel); err != nil {
		logger.Error("processUpdate, updating channel", "error", err, "channelName", channelName)
		return err
	}

	return nil
}
