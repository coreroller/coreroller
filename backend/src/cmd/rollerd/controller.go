package main

import (
	"database/sql"
	"encoding/json"
	"io"
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
	authenticator *auth.DigestAuth
	syncer        *syncer.Syncer
}

func newController() (*controller, error) {
	api, err := api.New()
	if err != nil {
		return nil, err
	}

	syncer, err := syncer.New(api)
	if err != nil {
		return nil, err
	}
	go syncer.Start()

	c := &controller{
		api:    api,
		syncer: syncer,
	}
	c.authenticator = auth.NewDigestAuthenticator("coreroller.org", c.getSecret)

	return c, nil
}

func (ctl *controller) close() {
	ctl.syncer.Stop()
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
			http.Error(w, http.StatusText(400), 400)
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
		logger.Error("updateUserPassword", "error", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	teamID, _ := c.Env["team_id"].(string)
	username, _ := c.Env["username"].(string)

	err := ctl.api.UpdateUserPassword(username, update.Password)
	switch err {
	case nil:
		http.Error(w, http.StatusText(204), 204)
	default:
		logger.Error("updateUserPassword", "error", err, "team", teamID, "username", username)
		http.Error(w, http.StatusText(400), 400)
	}
}

// ----------------------------------------------------------------------------
// API: applications CRUD
//

func (ctl *controller) addApp(c web.C, w http.ResponseWriter, r *http.Request) {
	sourceAppID := r.URL.Query().Get("clone_from")

	app := api.Application{}
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		logger.Error("addApp", "error", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}
	app.TeamID = c.Env["team_id"].(string)

	_, err := ctl.api.AddAppCloning(&app, sourceAppID)
	switch err {
	case nil:
	default:
		logger.Error("addApp", "error", err, "app", app)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	appJSON, err := ctl.api.GetAppJSON(app.ID)
	if err != nil {
		logger.Error("addApp", "error", err, "appID", app.ID)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	w.Write(appJSON)
}

func (ctl *controller) updateApp(c web.C, w http.ResponseWriter, r *http.Request) {
	app := api.Application{}
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		logger.Error("updateApp", "error", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}
	app.ID = c.URLParams["app_id"]
	app.TeamID = c.Env["team_id"].(string)

	err := ctl.api.UpdateApp(&app)
	switch err {
	case nil:
	default:
		logger.Error("updatedApp", "error", err, "app", app)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	appJSON, err := ctl.api.GetAppJSON(app.ID)
	if err != nil {
		logger.Error("updateApp", "error", err, "appID", app.ID)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	w.Write(appJSON)
}

func (ctl *controller) deleteApp(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]

	err := ctl.api.DeleteApp(appID)
	switch err {
	case nil:
		http.Error(w, http.StatusText(204), 204)
	default:
		logger.Error("deleteApp", "error", err, "appID", appID)
		http.Error(w, http.StatusText(400), 400)
	}
}

func (ctl *controller) getApp(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]

	appJSON, err := ctl.api.GetAppJSON(appID)
	switch err {
	case nil:
		w.Write(appJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getApp", "error", err, "appID", appID)
		http.Error(w, http.StatusText(400), 400)
	}
}

func (ctl *controller) getApps(c web.C, w http.ResponseWriter, r *http.Request) {
	teamID, _ := c.Env["team_id"].(string)
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	appsJSON, err := ctl.api.GetAppsJSON(teamID, page, perPage)
	switch err {
	case nil:
		w.Write(appsJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getApps", "error", err, "teamID", teamID)
		http.Error(w, http.StatusText(400), 400)
	}
}

// ----------------------------------------------------------------------------
// API: groups CRUD
//

func (ctl *controller) addGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	group := api.Group{}
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		logger.Error("addGroup", "error", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}
	group.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddGroup(&group)
	switch err {
	case nil:
	default:
		logger.Error("addGroup", "error", err, "group", group)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	groupJSON, err := ctl.api.GetGroupJSON(group.ID)
	if err != nil {
		logger.Error("addGroup", "error", err, "groupID", group.ID)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	w.Write(groupJSON)
}

func (ctl *controller) updateGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	group := api.Group{}
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		logger.Error("updateGroup", "error", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}
	group.ID = c.URLParams["group_id"]
	group.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdateGroup(&group)
	switch err {
	case nil:
	default:
		logger.Error("updateGroup", "error", err, "group", group)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	groupJSON, err := ctl.api.GetGroupJSON(group.ID)
	if err != nil {
		logger.Error("updateGroup", "error", err, "groupID", group.ID)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	w.Write(groupJSON)
}

func (ctl *controller) deleteGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	groupID := c.URLParams["group_id"]

	err := ctl.api.DeleteGroup(groupID)
	switch err {
	case nil:
		http.Error(w, http.StatusText(204), 204)
	default:
		logger.Error("deleteGroup", "error", err, "groupID", groupID)
		http.Error(w, http.StatusText(400), 400)
	}
}

func (ctl *controller) getGroup(c web.C, w http.ResponseWriter, r *http.Request) {
	groupID := c.URLParams["group_id"]

	groupJSON, err := ctl.api.GetGroupJSON(groupID)
	switch err {
	case nil:
		w.Write(groupJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getGroup", "error", err, "groupID", groupID)
		http.Error(w, http.StatusText(400), 400)
	}
}

func (ctl *controller) getGroups(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	groupsJSON, err := ctl.api.GetGroupsJSON(appID, page, perPage)
	switch err {
	case nil:
		w.Write(groupsJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getGroups", "error", err, "appID", appID)
		http.Error(w, http.StatusText(400), 400)
	}
}

// ----------------------------------------------------------------------------
// API: channels CRUD
//

func (ctl *controller) addChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channel := api.Channel{}
	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		logger.Error("addChannel", "error", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}
	channel.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddChannel(&channel)
	switch err {
	case nil:
	default:
		logger.Error("addChannel", "error", err, "channel", channel)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	channelJSON, err := ctl.api.GetChannelJSON(channel.ID)
	if err != nil {
		logger.Error("addChannel", "error", err, "channelID", channel.ID)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	w.Write(channelJSON)
}

func (ctl *controller) updateChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channel := api.Channel{}
	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		logger.Error("updateChannel", "error", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}
	channel.ID = c.URLParams["channel_id"]
	channel.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdateChannel(&channel)
	switch err {
	case nil:
	default:
		logger.Error("updateChannel", "error", err, "channel", channel)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	channelJSON, err := ctl.api.GetChannelJSON(channel.ID)
	if err != nil {
		logger.Error("updateChannel", "error", err, "channelID", channel.ID)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	w.Write(channelJSON)
}

func (ctl *controller) deleteChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channelID := c.URLParams["channel_id"]

	err := ctl.api.DeleteChannel(channelID)
	switch err {
	case nil:
		http.Error(w, http.StatusText(204), 204)
	default:
		logger.Error("deleteChannel", "error", err, "channelID", channelID)
		http.Error(w, http.StatusText(400), 400)
	}
}

func (ctl *controller) getChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	channelID := c.URLParams["channel_id"]

	channelJSON, err := ctl.api.GetChannelJSON(channelID)
	switch err {
	case nil:
		w.Write(channelJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getChannel", "error", err, "channelID", channelID)
		http.Error(w, http.StatusText(400), 400)
	}
}

func (ctl *controller) getChannels(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	channelsJSON, err := ctl.api.GetChannelsJSON(appID, page, perPage)
	switch err {
	case nil:
		w.Write(channelsJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getChannels", "error", err, "appID", appID)
		http.Error(w, http.StatusText(400), 400)
	}
}

// ----------------------------------------------------------------------------
// API: packages CRUD
//

func (ctl *controller) addPackage(c web.C, w http.ResponseWriter, r *http.Request) {
	pkg := api.Package{}
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		logger.Error("addPackage", "error", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}
	pkg.ApplicationID = c.URLParams["app_id"]

	_, err := ctl.api.AddPackage(&pkg)
	switch err {
	case nil:
	default:
		logger.Error("addPackage", "error", err, "package", pkg)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	pkgJSON, err := ctl.api.GetPackageJSON(pkg.ID)
	if err != nil {
		logger.Error("addPackage", "error", err, "packageID", pkg.ID)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	w.Write(pkgJSON)
}

func (ctl *controller) updatePackage(c web.C, w http.ResponseWriter, r *http.Request) {
	pkg := api.Package{}
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		logger.Error("updatePackage", "error", err)
		http.Error(w, http.StatusText(400), 400)
		return
	}
	pkg.ID = c.URLParams["package_id"]
	pkg.ApplicationID = c.URLParams["app_id"]

	err := ctl.api.UpdatePackage(&pkg)
	switch err {
	case nil:
	default:
		logger.Error("updatePackage", "error", err, "package", pkg)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	pkgJSON, err := ctl.api.GetPackageJSON(pkg.ID)
	if err != nil {
		logger.Error("updatePackage", "error", err, "packageID", pkg.ID)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	w.Write(pkgJSON)
}

func (ctl *controller) deletePackage(c web.C, w http.ResponseWriter, r *http.Request) {
	packageID := c.URLParams["package_id"]

	err := ctl.api.DeletePackage(packageID)
	switch err {
	case nil:
		http.Error(w, http.StatusText(204), 204)
	default:
		logger.Error("deletePackage", "error", err, "packageID", packageID)
		http.Error(w, http.StatusText(400), 400)
	}
}

func (ctl *controller) getPackage(c web.C, w http.ResponseWriter, r *http.Request) {
	packageID := c.URLParams["package_id"]

	pkgJSON, err := ctl.api.GetPackageJSON(packageID)
	switch err {
	case nil:
		w.Write(pkgJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getPackage", "error", err, "packageID", packageID)
		http.Error(w, http.StatusText(400), 400)
	}
}

func (ctl *controller) getPackages(c web.C, w http.ResponseWriter, r *http.Request) {
	appID := c.URLParams["app_id"]
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	perPage, _ := strconv.ParseUint(r.URL.Query().Get("perpage"), 10, 64)

	pkgsJSON, err := ctl.api.GetPackagesJSON(appID, page, perPage)
	switch err {
	case nil:
		w.Write(pkgsJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getPackages", "error", err, "appID", appID)
		http.Error(w, http.StatusText(400), 400)
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

	instanceStatusHistoryJSON, err := ctl.api.GetInstanceStatusHistoryJSON(instanceID, appID, groupID, limit)
	switch err {
	case nil:
		w.Write(instanceStatusHistoryJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getInstanceStatusHistory", "error", err, "appID", appID, "groupID", groupID, "instanceID", instanceID, "limit", limit)
		http.Error(w, http.StatusText(400), 400)
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

	instancesJSON, err := ctl.api.GetInstancesJSON(p)
	switch err {
	case nil:
		w.Write(instancesJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getInstances", "error", err, "appID", appID, "groupID", groupID, "p", p)
		http.Error(w, http.StatusText(400), 400)
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

	activityEntriesJSON, err := ctl.api.GetActivityJSON(teamID, p)
	switch err {
	case nil:
		w.Write(activityEntriesJSON)
	case sql.ErrNoRows:
		w.Write([]byte(`[]`))
	default:
		logger.Error("getActivity", "error", err, "teamID", teamID, "p", p)
		http.Error(w, http.StatusText(400), 400)
	}
}

// ----------------------------------------------------------------------------
// OMAHA server
//

func (ctl *controller) processOmahaRequest(c web.C, w http.ResponseWriter, r *http.Request) {
	pipeReader, pipeWriter := io.Pipe()
	go omaha.HandleRequest(ctl.api, r.Body, pipeWriter, getRequestIP(r))
	io.Copy(w, pipeReader)
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
