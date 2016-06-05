package main

import (
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"api"
	"omaha"
	"syncer"

	"github.com/abbot/go-http-auth"
	"github.com/zenazn/goji/web"
)

type controller struct {
	api           *api.API
	omahaHandler  *omaha.Handler
	authenticator *auth.DigestAuth
	syncer        *syncer.Syncer
}

func newController(enableSyncer bool) (*controller, error) {
	api, err := api.New()
	if err != nil {
		return nil, err
	}

	c := &controller{
		api:          api,
		omahaHandler: omaha.NewHandler(api),
	}
	c.authenticator = auth.NewDigestAuthenticator("coreroller.org", c.getSecret)

	if enableSyncer {
		syncer, err := syncer.New(api)
		if err != nil {
			return nil, err
		}
		c.syncer = syncer
		go syncer.Start()
	}

	return c, nil
}

func (ctl *controller) close() {
	if ctl.syncer != nil {
		ctl.syncer.Stop()
	}
	ctl.api.Close()
}

// ----------------------------------------------------------------------------
// Authentication
//

// getSecret returns the hashed password for the provided user.
func (ctl *controller) getSecret(username, realm string) string {
	user, err := ctl.api.GetUser(username)
	if err != nil {
		return ""
	}
	return user.Secret
}

// authenticate is a middleware handler in charge of authenticating requests.
func (ctl *controller) authenticate(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		username := r.Header.Get("X-Authenticated-Username")

		user, err := ctl.api.GetUser(username)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		c.Env["username"] = username
		c.Env["team_id"] = user.TeamID

		w.Header()["X-Authenticated-Username"] = []string{username}
		h.ServeHTTP(w, r)
	}
	return auth.JustCheck(ctl.authenticator, fn)
}

// ----------------------------------------------------------------------------
// API: users
//

func (ctl *controller) updateUserPassword(c web.C, w http.ResponseWriter, r *http.Request) {
	var update struct {
		Password string
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		logger.Error("updateUserPassword", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	teamID, _ := c.Env["team_id"].(string)
	username, _ := c.Env["username"].(string)

	err := ctl.api.UpdateUserPassword(username, update.Password)
	switch err {
	case nil:
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
	default:
		logger.Error("updateUserPassword", "error", err.Error(), "team", teamID, "username", username)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: applications CRUD
//

func (ctl *controller) addApp(c web.C, w http.ResponseWriter, r *http.Request) {
	sourceAppID := r.URL.Query().Get("clone_from")

	app := &api.Application{}
	if err := json.NewDecoder(r.Body).Decode(app); err != nil {
		logger.Error("addApp - decoding payload", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	app.TeamID = c.Env["team_id"].(string)

	_, err := ctl.api.AddAppCloning(app, sourceAppID)
	if err != nil {
		logger.Error("addApp - cloning app", "error", err.Error(), "app", app, "sourceAppID", sourceAppID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	app, err = ctl.api.GetApp(app.ID)
	if err != nil {
		logger.Error("addApp - getting added app", "error", err.Error(), "appID", app.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(app); err != nil {
		logger.Error("addApp - encoding app", "error", err.Error(), "app", app)
	}
}

func (ctl *controller) updateApp(c web.C, w http.ResponseWriter, r *http.Request) {
	app := &api.Application{}
	if err := json.NewDecoder(r.Body).Decode(app); err != nil {
		logger.Error("updateApp - decoding payload", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	app.ID = c.URLParams["app_id"]
	app.TeamID = c.Env["team_id"].(string)

	err := ctl.api.UpdateApp(app)
	if err != nil {
		logger.Error("updatedApp - updating app", "error", err.Error(), "app", app)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	app, err = ctl.api.GetApp(app.ID)
	if err != nil {
		logger.Error("updateApp - getting updated app", "error", err.Error(), "appID", app.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(app); err != nil {
		logger.Error("updateApp - encoding app", "error", err.Error(), "appID", app.ID)
	}
}

func (ctl *controller) deleteApp(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]

	err := ctl.api.DeleteApp(appID)
	switch err {
	case nil:
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
	default:
		logger.Error("deleteApp", "error", err.Error(), "appID", appID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (ctl *controller) getApp(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]

	app, err := ctl.api.GetApp(appID)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(app); err != nil {
			logger.Error("getApp - encoding app", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getApp - getting app", "error", err.Error(), "appID", appID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (ctl *controller) getApps(c web.C, w http.ResponseWriter, r *http.Request) {
	teamID, _ := c.Env["team_id"].(string)
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	apps, err := ctl.api.GetApps(teamID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(apps); err != nil {
			logger.Error("getApps - encoding apps", "error", err.Error(), "teamID", teamID)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getApps - getting apps", "error", err.Error(), "teamID", teamID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: groups CRUD
//

func (ctl *controller) addGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	group := &api.Group{}
	if err := json.NewDecoder(r.Body).Decode(group); err != nil {
		logger.Error("addGroup - decoding payload", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	group.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddGroup(group)
	if err != nil {
		logger.Error("addGroup - adding group", "error", err.Error(), "group", group)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	group, err = ctl.api.GetGroup(group.ID)
	if err != nil {
		logger.Error("addGroup - getting added group", "error", err.Error(), "groupID", group.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(group); err != nil {
		logger.Error("addGroup - encoding group", "error", err.Error(), "group", group)
	}
}

func (ctl *controller) updateGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	group := &api.Group{}
	if err := json.NewDecoder(r.Body).Decode(group); err != nil {
		logger.Error("updateGroup - decoding payload", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	group.ID = c.URLParams["group_id"]
	group.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdateGroup(group)
	if err != nil {
		logger.Error("updateGroup - updating group", "error", err.Error(), "group", group)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	group, err = ctl.api.GetGroup(group.ID)
	if err != nil {
		logger.Error("updateGroup - fetching updated group", "error", err.Error(), "groupID", group.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(group); err != nil {
		logger.Error("updateGroup - encoding group", "error", err.Error(), "group", group)
	}
}

func (ctl *controller) deleteGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	groupID := c.URLParams["group_id"]

	err := ctl.api.DeleteGroup(groupID)
	switch err {
	case nil:
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
	default:
		logger.Error("deleteGroup", "error", err.Error(), "groupID", groupID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (ctl *controller) getGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	groupID := c.URLParams["group_id"]

	group, err := ctl.api.GetGroup(groupID)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(group); err != nil {
			logger.Error("getGroup - encoding group", "error", err.Error(), "group", group)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getGroup - getting group", "error", err.Error(), "groupID", groupID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (ctl *controller) getGroups(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	groups, err := ctl.api.GetGroups(appID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(groups); err != nil {
			logger.Error("getGroups - encoding groups", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getGroups - getting groups", "error", err.Error(), "appID", appID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: channels CRUD
//

func (ctl *controller) addChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channel := &api.Channel{}
	if err := json.NewDecoder(r.Body).Decode(channel); err != nil {
		logger.Error("addChannel", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	channel.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddChannel(channel)
	if err != nil {
		logger.Error("addChannel", "error", err.Error(), "channel", channel)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	channel, err = ctl.api.GetChannel(channel.ID)
	if err != nil {
		logger.Error("addChannel", "error", err.Error(), "channelID", channel.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(channel); err != nil {
		logger.Error("addChannel - encoding channel", "error", err.Error(), "channelID", channel.ID)
	}
}

func (ctl *controller) updateChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channel := &api.Channel{}
	if err := json.NewDecoder(r.Body).Decode(channel); err != nil {
		logger.Error("updateChannel - decoding payload", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	channel.ID = c.URLParams["channel_id"]
	channel.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdateChannel(channel)
	if err != nil {
		logger.Error("updateChannel - updating channel", "error", err.Error(), "channel", channel)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	channel, err = ctl.api.GetChannel(channel.ID)
	if err != nil {
		logger.Error("updateChannel - getting channel updated", "error", err.Error(), "channelID", channel.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(channel); err != nil {
		logger.Error("updateChannel - encoding channel", "error", err.Error(), "channelID", channel.ID)
	}
}

func (ctl *controller) deleteChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channelID := c.URLParams["channel_id"]

	err := ctl.api.DeleteChannel(channelID)
	switch err {
	case nil:
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
	default:
		logger.Error("deleteChannel", "error", err.Error(), "channelID", channelID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (ctl *controller) getChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channelID := c.URLParams["channel_id"]

	channel, err := ctl.api.GetChannel(channelID)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(channel); err != nil {
			logger.Error("getChannel - encoding channel", "error", err.Error(), "channelID", channelID)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getChannel - getting updated channel", "error", err.Error(), "channelID", channelID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (ctl *controller) getChannels(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	channels, err := ctl.api.GetChannels(appID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(channels); err != nil {
			logger.Error("getChannels - encoding channel", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getChannels - getting channels", "error", err.Error(), "appID", appID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: packages CRUD
//

func (ctl *controller) addPackage(c web.C, w http.ResponseWriter, r *http.Request) {
	pkg := &api.Package{}
	if err := json.NewDecoder(r.Body).Decode(pkg); err != nil {
		logger.Error("addPackage - decoding payload", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	pkg.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddPackage(pkg)
	if err != nil {
		logger.Error("addPackage - adding package", "error", err.Error(), "package", pkg)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	pkg, err = ctl.api.GetPackage(pkg.ID)
	if err != nil {
		logger.Error("addPackage - getting added package", "error", err.Error(), "packageID", pkg.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(pkg); err != nil {
		logger.Error("addPackage - encoding package", "error", err.Error(), "packageID", pkg.ID)
	}
}

func (ctl *controller) updatePackage(c web.C, w http.ResponseWriter, r *http.Request) {
	pkg := &api.Package{}
	if err := json.NewDecoder(r.Body).Decode(pkg); err != nil {
		logger.Error("updatePackage - decoding payload", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	pkg.ID = c.URLParams["package_id"]
	pkg.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdatePackage(pkg)
	if err != nil {
		logger.Error("updatePackage - updating package", "error", err.Error(), "package", pkg)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	pkg, err = ctl.api.GetPackage(pkg.ID)
	if err != nil {
		logger.Error("addPackage - getting updated package", "error", err.Error(), "packageID", pkg.ID)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(pkg); err != nil {
		logger.Error("updatePackage - encoding package", "error", err.Error(), "packageID", pkg.ID)
	}
}

func (ctl *controller) deletePackage(c web.C, w http.ResponseWriter, r *http.Request) {
	packageID := c.URLParams["package_id"]

	err := ctl.api.DeletePackage(packageID)
	switch err {
	case nil:
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
	default:
		logger.Error("deletePackage", "error", err.Error(), "packageID", packageID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (ctl *controller) getPackage(c web.C, w http.ResponseWriter, r *http.Request) {
	packageID := c.URLParams["package_id"]

	pkg, err := ctl.api.GetPackage(packageID)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(pkg); err != nil {
			logger.Error("getPackage - encoding package", "error", err.Error(), "packageID", packageID)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getPackage - getting package", "error", err.Error(), "packageID", packageID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (ctl *controller) getPackages(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	pkgs, err := ctl.api.GetPackages(appID, page, perPage)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(pkgs); err != nil {
			logger.Error("getPackages - encoding packages", "error", err.Error(), "appID", appID)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getPackages - getting packages", "error", err.Error(), "appID", appID)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: instances
//

func (ctl *controller) getInstanceStatusHistory(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	groupID := c.URLParams["group_id"]
	instanceID := c.URLParams["instance_id"]
	limit, _ := strconv.ParseUint(r.URL.Query().Get("limit"), 10, 64)

	instanceStatusHistory, err := ctl.api.GetInstanceStatusHistory(instanceID, appID, groupID, limit)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(instanceStatusHistory); err != nil {
			logger.Error("getInstanceStatusHistory - encoding status history", "error", err.Error(), "appID", appID, "groupID", groupID, "instanceID", instanceID, "limit", limit)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getInstanceStatusHistory - getting status history", "error", err.Error(), "appID", appID, "groupID", groupID, "instanceID", instanceID, "limit", limit)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (ctl *controller) getInstances(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	groupID := c.URLParams["group_id"]

	p := api.InstancesQueryParams{
		ApplicationID: appID,
		GroupID:       groupID,
		Version:       r.URL.Query().Get("version"),
	}
	p.Status, _ = strconv.Atoi(r.URL.Query().Get("status"))
	p.Page, _ = strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	p.PerPage, _ = strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	instances, err := ctl.api.GetInstances(p)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(instances); err != nil {
			logger.Error("getInstances - encoding instances", "error", err.Error(), "params", p)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getInstances - getting instances", "error", err.Error(), "params", p)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// API: activity
//

func (ctl *controller) getActivity(c web.C, w http.ResponseWriter, r *http.Request) {
	teamID, _ := c.Env["team_id"].(string)

	p := api.ActivityQueryParams{
		AppID:      r.URL.Query().Get("app"),
		GroupID:    r.URL.Query().Get("group"),
		ChannelID:  r.URL.Query().Get("channel"),
		InstanceID: r.URL.Query().Get("instance"),
		Version:    r.URL.Query().Get("version"),
	}
	p.Severity, _ = strconv.Atoi(r.URL.Query().Get("severity"))
	p.Start, _ = time.Parse(time.RFC3339, r.URL.Query().Get("start"))
	p.End, _ = time.Parse(time.RFC3339, r.URL.Query().Get("end"))
	p.Page, _ = strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	p.PerPage, _ = strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	activityEntries, err := ctl.api.GetActivity(teamID, p)
	switch err {
	case nil:
		if err := json.NewEncoder(w).Encode(activityEntries); err != nil {
			logger.Error("getActivity - encoding activity entries", "error", err.Error(), "params", p)
		}
	case sql.ErrNoRows:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		logger.Error("getActivity", "error", err, "teamID", teamID, "params", p)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

// ----------------------------------------------------------------------------
// OMAHA server
//

func (ctl *controller) processOmahaRequest(c web.C, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctl.omahaHandler.Handle(r.Body, w, getRequestIP(r))
}

// ----------------------------------------------------------------------------
// Helpers
//

func getRequestIP(r *http.Request) string {
	if ip := r.Header.Get("X-FORWARDED-FOR"); ip != "" {
		return ip
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
