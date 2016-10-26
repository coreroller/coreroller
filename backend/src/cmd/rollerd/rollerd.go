package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/mgutz/logxi/v1"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

const (
	coreosPkgsRouterPrefix = "/coreos/"
)

var (
	enableSyncer       = flag.Bool("enable-syncer", true, "Enable CoreOS packages syncer")
	hostCoreosPackages = flag.Bool("host-coreos-packages", false, "Host CoreOS packages in CoreRoller")
	coreosPackagesPath = flag.String("coreos-packages-path", "", "Path where CoreOS packages files are stored")
	corerollerURL      = flag.String("coreroller-url", "", "CoreRoller URL (http://host:port - required when hosting CoreOS packages in CoreRoller)")
	httpLog            = flag.Bool("http-log", false, "Enable http requests logging")
	httpStaticDir      = flag.String("http-static-dir", "../frontend/built", "Path to frontend static files")
	logger             = log.New("rollerd")
)

func main() {
	flag.Parse()

	if err := checkArgs(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	conf := &controllerConfig{
		enableSyncer:       *enableSyncer,
		hostCoreosPackages: *hostCoreosPackages,
		coreosPackagesPath: *coreosPackagesPath,
		corerollerURL:      *corerollerURL,
	}
	ctl, err := newController(conf)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer ctl.close()

	setupRoutes(ctl)

	if !*httpLog {
		goji.Abandon(middleware.Logger)
	}
	goji.Serve()
}

func checkArgs() error {
	if *hostCoreosPackages {
		if *coreosPackagesPath == "" {
			return errors.New("Invalid CoreOS packages path. Please ensure you provide a valid path using -coreos-packages-path")
		}
		tmpFile, err := ioutil.TempFile(*coreosPackagesPath, "")
		if err != nil {
			return errors.New("Invalid CoreOS packages path: " + err.Error())
		}
		defer os.Remove(tmpFile.Name())

		if _, err := url.ParseRequestURI(*corerollerURL); err != nil {
			return errors.New("Invalid CoreRoller url. Please ensure the value provided using -coreroller-url is a valid url.")
		}
	}

	return nil
}

func setupRoutes(ctl *controller) {
	// API router setup
	apiRouter := web.New()
	apiRouter.Use(ctl.authenticate)
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

	// Host CoreOS packages payloads
	if *hostCoreosPackages {
		coreosPkgsRouter := web.New()
		coreosPkgsRouter.Use(middleware.SubRouter)
		goji.Handle(coreosPkgsRouterPrefix+"*", coreosPkgsRouter)
		coreosPkgsRouter.Handle("/*", http.FileServer(http.Dir(*coreosPackagesPath)))
	}

	// Serve frontend static content
	staticRouter := web.New()
	staticRouter.Use(ctl.authenticate)
	goji.Handle("/*", staticRouter)
	staticRouter.Handle("/*", http.FileServer(http.Dir(*httpStaticDir)))
}
