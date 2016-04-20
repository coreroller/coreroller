package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/mgutz/logxi/v1"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

var (
	enableSyncer = flag.Bool("enable-syncer", true, "Enable CoreOS packages syncer")
	logger       = log.New("rollerd")
)

func main() {
	flag.Parse()

	ctl, err := newController(*enableSyncer)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer ctl.close()

	setupRoutes(ctl)
	goji.Serve()
}

func setupRoutes(ctl *controller) {
	// API router setup
	apiRouter := web.New()
	apiRouter.Use(ctl.authenticate)

	// Disable the built-in "access-log" logger
	goji.Abandon(middleware.Logger)

	// Define the API router
	goji.Handle("/api/*", apiRouter)

	// API routes

	// Users
	apiRouter.Put("/api/password", ctl.updateUserPassword)

	// Applications
	apiRouter.Post("/api/apps", ctl.addApp)
	apiRouter.Put("/api/apps/:app_id", ctl.updateApp)
	apiRouter.Delete("/api/apps/:app_id", ctl.deleteApp)
	apiRouter.Get("/api/apps/:app_id", ctl.getApp)
	apiRouter.Get("/api/apps", ctl.getApps)

	// Groups
	apiRouter.Post("/api/apps/:app_id/groups", ctl.addGroup)
	apiRouter.Put("/api/apps/:app_id/groups/:group_id", ctl.updateGroup)
	apiRouter.Delete("/api/apps/:app_id/groups/:group_id", ctl.deleteGroup)
	apiRouter.Get("/api/apps/:app_id/groups/:group_id", ctl.getGroup)
	apiRouter.Get("/api/apps/:app_id/groups", ctl.getGroups)

	// Channels
	apiRouter.Post("/api/apps/:app_id/channels", ctl.addChannel)
	apiRouter.Put("/api/apps/:app_id/channels/:channel_id", ctl.updateChannel)
	apiRouter.Delete("/api/apps/:app_id/channels/:channel_id", ctl.deleteChannel)
	apiRouter.Get("/api/apps/:app_id/channels/:channel_id", ctl.getChannel)
	apiRouter.Get("/api/apps/:app_id/channels", ctl.getChannels)

	// Packages
	apiRouter.Post("/api/apps/:app_id/packages", ctl.addPackage)
	apiRouter.Put("/api/apps/:app_id/packages/:package_id", ctl.updatePackage)
	apiRouter.Delete("/api/apps/:app_id/packages/:package_id", ctl.deletePackage)
	apiRouter.Get("/api/apps/:app_id/packages/:package_id", ctl.getPackage)
	apiRouter.Get("/api/apps/:app_id/packages", ctl.getPackages)

	// Instances
	apiRouter.Get("/api/apps/:app_id/groups/:group_id/instances/:instance_id/status_history", ctl.getInstanceStatusHistory)
	apiRouter.Get("/api/apps/:app_id/groups/:group_id/instances", ctl.getInstances)

	// Activity
	apiRouter.Get("/api/activity", ctl.getActivity)

	// Omaha server router setup
	omahaRouter := web.New()
	omahaRouter.Use(middleware.SubRouter)
	goji.Handle("/omaha/*", omahaRouter)
	goji.Handle("/v1/update/*", omahaRouter)

	// Omaha server routes
	omahaRouter.Post("/", ctl.processOmahaRequest)

	// Serve static content
	staticRouter := web.New()
	staticRouter.Use(ctl.authenticate)
	goji.Handle("/*", staticRouter)
	staticRouter.Handle("/*", http.FileServer(http.Dir("../frontend/built")))
}
