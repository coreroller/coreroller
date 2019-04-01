![Coreroller](https://github.com/coreroller/coreroller/raw/master/docs/screenshots/coreroller.png)

[![Travis Widget]][Travis] [![GoReportCard Widget]][GoReportCard]

[Travis]: https://travis-ci.org/coreroller/coreroller
[Travis Widget]: https://travis-ci.org/coreroller/coreroller.svg?branch=master
[GoReportCard]: https://goreportcard.com/report/coreroller/coreroller
[GoReportCard Widget]: https://goreportcard.com/badge/coreroller/coreroller

## Overview

**CoreRoller** is a set of tools to control and monitor the rollout of your updates. It's aimed to be an open source alternative to CoreOS CoreUpdate.

Screenshots:

<table>
    <tr>
        <td width="33%"><img src="https://github.com/coreroller/coreroller/raw/master/docs/screenshots/screenshot1.png"></td>
        <td width="33%"><img src="https://github.com/coreroller/coreroller/raw/master/docs/screenshots/screenshot2.png"></td>
        <td width="33%"><img src="https://github.com/coreroller/coreroller/raw/master/docs/screenshots/screenshot3.png"></td>
    </tr>
    <tr>
        <td><img src="https://github.com/coreroller/coreroller/raw/master/docs/screenshots/screenshot4.png"></td>
        <td><img src="https://github.com/coreroller/coreroller/raw/master/docs/screenshots/screenshot5.png"></td>
        <td><img src="https://github.com/coreroller/coreroller/raw/master/docs/screenshots/screenshot6.png"></td>
    </tr>
</table>

## News

*May 8, 2017*

CoreRoller 1.0 released! The project will use semantic versioning from now on and docker images automatically built will be tagged accordingly as well. If you were already using CoreRoller before this release please check out the [release notes](https://github.com/coreroller/coreroller#release-notes) below.

## Features

- Dashboard to control and monitor your applications updates
- Ready to serve updates for to CoreOS clusters out of the box
- Manage updates for your own applications as well, not just CoreOS
- Define your own groups and channels, even for the CoreOS application (pre-installed)
- Define roll-out policies per group, controlling how updates should be applied to a set of instances
- Pause/resume updates at any time at the group level
- Statistics about versions installed in your instances, updates progress status, etc
- Activity stream in UI to get notified about important events or errors
- Post HipChat notifications about important events
- Based on the [Omaha](https://code.google.com/p/omaha/wiki/ServerProtocol) protocol developed by Google

## Status

CoreRoller is *stable*. It has been used in production enviroments for over a year without any major issues. 

Please report any bug you find as [issues](https://github.com/coreroller/coreroller/issues) on this repository.

## Getting started

The best way to give it a try is to launch a Docker container using the public images hosted in Docker Hub: 

	docker run -d -p 8000:8000 coreroller/demo

Once the container is up, just point your browser to:

	http://localhost:8000/
	
and you should be ready to go. Default username/password is `admin/admin`.

This demo container runs `PostgreSQL` (the datastore used by CoreRoller) and the `CoreRoller server` (aka rollerd).

In addition to this [coreroller/demo](https://hub.docker.com/r/coreroller/demo) image, there are some other images available in the docker hub that may be helpful to you as a starting point when preparing your own custom images.

- **[coreroller/rollerd](https://hub.docker.com/r/coreroller/rollerd)**: this image runs the backend server, a dependency free Golang binary that will power the dashboard and serve all Omaha updates and events requests.

- **[coreroller/postgres](https://hub.docker.com/r/coreroller/postgres)**: this image runs PostgreSQL and creates the database used by CoreRoller. Do not forget to setup properly the volumes in the container to avoid any data loss.

These images are rebuilt and tagged automatically after every commit. Dockerfiles used to build them can be found in the `backend/docker` directory.

Additionally, in the `backend/systemd` directory there are some systemd unit files that might be handy in case you want to deploy CoreRoller in your CoreOS cluster using `fleet`. You can also use the sample kubernetes configuration files in the `backend/kubernetes` folder to deploy CoreRoller using `kubernetes` (`kubectl create -f backend/kubernetes`). These units and config files are just samples, feel free to adjust them to suit your specific needs.

## Managing CoreOS updates

Once you have CoreRoller up, it's time to give it some work to do. You may be interested in managing the CoreOS updates in your cluster with it.

The process is slightly different if you want to do it in existing machines or in new ones. In both cases it's very simple as it only requires updating the server url the CoreOS updater uses to pull updates from. By default your CoreOS instances use the public CoreOS servers to get updates, so you'll have to point them to your CoreRoller deployment instead.

### New machines

In new machines, you can set up the updates server in the cloud config. Here is a small example of how to do it:

	coreos:
		update:
			group: stable
			server: http://your.coreroller.host:port/v1/update/

In addition to the default `stable`, `beta` and `alpha` groups, you can also create and use **custom groups** for greater control over the updates. In that case, you **must** use the group id (not the name) you will find next to the group name in the dashboard.

	coreos:
		update:
			group: ab51a000-02dc-4fc7-a6b0-c42881c89856
			server: http://your.coreroller.host:port/v1/update/

**Note**: The sample CoreRoller containers provided use the `port 8000` by default (**plain HTTP, no SSL**). Please adjust the update url setup in your servers to match your CoreRoller deployment.

### Existing machines

To update the update server in existing instances please edit `/etc/coreos/update.conf` and update the `SERVER` value (and optionally `GROUP` if needed):

	SERVER=https://your.coreroller.host/v1/update/

When using custom groups instead of the official ones (stable, beta, alpha) the group id **must** be used, not the group name:

    GROUP=ab51a000-02dc-4fc7-a6b0-c42881c89856
	
To apply these changes run:

	sudo systemctl restart update-engine
	
In may take a few minutes to see an update request coming through. If you want to see it sooner, you can force it running this command:

	update_engine_client -update

**Note:** the CoreUpdate docs do a great job explaining in detail how this process works and most of the information it contains applies to CoreRoller as well, so please have a look at them [here](https://coreos.com/products/coreupdate/docs/latest/configure-machines.html) for more information.

### CoreOS packages in CoreRoller

Out of the box CoreRoller polls periodically the public CoreOS update servers to create packages and update channels in your CoreRoller deployment as they become publicly available. So if rollerd has access to the Internet you'll see eventually new packages (pointing to the official image files) added to the CoreOS application in CoreRoller. This functionality can be disabled if needed (i.e. you want to deploy your custom built images, etc) using the rollerd flag `-enable-syncer=false`.

By default, CoreRoller only stores metadata about the official CoreOS packages available, not the packages payload. This means that the updates CoreRoller serves to your instances contain instructions to download the packages payload from the public CoreOS update servers directly, so your servers need access to the Internet to download them.

In some cases, you may prefer to host the CoreOS packages payload as well in CoreRoller. When CoreRoller is instructed to behave this way, in addition to get the packages metadata, it will also download the package payload itself so that it can serve it to your instances when serving updates.

Enabling CoreOS packages payload hosting requires passing some parameters to rollerd:

    rollerd -host-coreos-packages=true -coreos-packages-path=/PATH/TO/STORE/PACKAGES -coreroller-url=http://your.coreroller.host:port

## Managing updates for your own applications

In addition to manage updates for CoreOS, you can use CoreRoller for your own applications as well. It's really easy to send updates and events requests to the Omaha server that CoreRoller provides.

In the `updaters/lib` directory there are some sample helpers that can be useful to create your own updaters that talk to CoreRoller or even embed them into your own applications. 

In the `updaters/examples` you'll find a sample minimal application built using [grace](https://github.com/facebookgo/grace) that is able to update itself using CoreRoller in a graceful way.

## Other features

### HipChat notifications

CoreRoller supports posting notifications to HipChat when certain events occur. This way you'll be notified when a channel points to a new package, a rollout starts, fails or succeeds, etc.

Enabling HipChat notifications requires setting a couple of environment variables for `rollerd`:

    CR_HIPCHAT_ROOM=ROOM_ID
    CR_HIPCHAT_TOKEN=TOKEN

## Contributing

CoreRoller is an Open Source project and we welcome contributions. Before submitting any code for new features or major changes, please open an [issue](https://github.com/coreroller/coreroller/issues) and discuss first.

Below you will find some introductory notes that should point you in the right direction to start playing with the CoreRoller source code.

### Setup
See: [docs/local-setup.md](./docs/local-setup.md)

### Backend

The CoreRoller backend (aka rollerd) is a Golang application. Builds and vendored dependencies are managed using [gb](http://getgb.io), so you just need a working Golang environment and gb installed to start working with it (there is **no** need to do a `go get` to fetch the dependencies as they are already contained in `backend/vendor`).

The backend source code is located inside `backend/src` and is structured as follows:

- **Package `api`**: provides functionality to do CRUD operations on all elements found in CoreRoller (applications, groups, channels, packages, etc), abstracting the rest of the components from the underlying datastore (PostgreSQL). It also controls the groups' roll-out policy logic and the instances/events registration.

- **Package `omaha`**: provides functionality to validate, handle, process and reply to Omaha updates and events requests received from the Omaha clients. It relies on the `api` package to get update packages, store events, or register instances when needed.

- **Package `syncer`**: provides some functionality to synchronize packages available in the official CoreOS channels, storing the references to them in your CoreRoller datastore and even downloading packages payloads when configured to do so. It's basically in charge of keeping up to date your the CoreOS application in your CoreRoller installation.

- **Cmd `rollerd`**: is the main backend process, exposing the functionality described above in the different packages through its http server. It provides several http endpoints used to drive most of the functionality of the dashboard as well as handling the Omaha updates and events requests received from your servers and applications.

- **Cmd `initdb`**: is just a helper to reset your database, and causing the migrations to be re-run. `rollerd` will apply all database migrations automatically, so this process should only be used to wipe out all your data and start from a clean state (you should probably never need it).

Please make sure that your code is formatted using `gofmt` and makes [gometalinter](https://github.com/alecthomas/gometalinter) happy :) 

#### Backend datastore (PostgreSQL)

CoreRoller uses `PostgreSQL` as datastore, so you will need it installed if you are planning to do some work on CoreRoller.

You can install it locally or use the docker image available in the docker hub (coreroller/postgres). If you don't use the docker image provided, you'll need to run the following commands to create the necessary databases:

	psql -U postgres -c "create database coreroller;"

To run the tests you will also need to setup the coreroller\_tests database:
	
	psql -U postgres -c "create database coreroller_tests;"

### Frontend

The frontend side of CoreRoller (dashboard) is a javascript web application built using `react/flux`.

To do some development in the frontend it's highly recommended to setup the backend service as well, as that will allow you to fully interact with the real API. Also, the backend server is ready to serve the static assets you'll build out of the box, so you can point your browser to `http://localhost:8000/` and interact with your development environment while you work on it.

To build the webapp and download its dependencies you'll need Node.js installed. Building the webapp is a simple and straightforward process:

Fetch the project dependencies

	npm install (inside frontend)
	
Build it

	npm run build
	
*That's it!* The build process will generate a built **main.js** containing the web application and a built **styles.css** file containing the styles generated from the less templates. Both built files can be found inside `frontend/built` in the `js` and `css` directories respectively, along with some other files that will allow you to serve the webapp straight away from rollerd.

While working on specific parts of the webapp, you may find handy running one of the watchers that will build the js or css files (or both) as soon as one relevant source file is modified.

	npm run watch
	npm run watch-js
	npm run watch-css
	
Unlike the build-* commands, the watchers do NOT minify the built files, so they'll be considerably bigger than the final ones.

Enjoy!

## Release notes

### 1.0

The main change that this first release introduces is dropping the PostgreSQL `semver` extension completely. A few months ago CoreRoller stopped using it, but it was still part of the database migrations. This was a [problem](https://github.com/coreroller/coreroller/issues/42) as even though we were not using the extension anymore, it had to be present in your PostgreSQL server so that all migrations could be processed correctly. In scenarios where installing the extension wasn't possible (i.e. AWS RDS), this was a blocker. So we decided to drop it completely and merge all database migrations into a single one (applying some other minor database schema changes as well).

If you were using CoreRoller before this release we recommend you to follow the procedure below to upgrade to the stable 1.0 version:

(please adjust database host and user as needed)

    # Step 1: Backup existing CoreRoller data and move it to safe location
    pg_dump -h COREROLLER_DB -U postgres -d coreroller --data-only --exclude-table=database_migrations > coreroller_backup.sql

    # Step 2: Start a new PostgreSQL server (free of any coreroller tables and data)
    # (If you are setting up your PostgreSQL instance manually instead of using the provided docker image don't forget creating the database)
	psql -h NEW_COREROLLER_DB -U postgres -c "create database coreroller;"

    # Step 3: Start a new rollerd 1.0 instance connecting to the new PostgreSQL server. The new migration file will create the schema and add the sample data.

    # Step 4: Delete the sample data from new PostgreSQL instance
    psql -h NEW_COREROLLER_DB -U postgres -d coreroller -c "delete from team; delete from event_type; delete from instance;"

    # Step 5: Restore data previously backed up into new PostgreSQL instance
    psql -h NEW_COREROLLER_DB -U postgres coreroller < coreroller_backup.sql

Future database schema changes will be applied automatically by `rollerd` using migrations as usual.

## License

CoreRoller is released under the terms of the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0).

CoreRoller was created by Cintia Sánchez García, Mathieu Lohier and Sergio Castaño Arteaga.
