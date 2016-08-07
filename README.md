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

## Features

- Dashboard to control and monitor your applications updates
- Manage updates for your own applications, not just for CoreOS
- Define your own groups and channels, even for the CoreOS application (pre-installed)
- Define roll-out policies per group, controlling how updates should be applied to a set of instances
- Pause/resume updates at any time at the group level
- Statistics about versions installed in your instances, updates progress status, etc
- Activity stream to get notified about important events or errors
- Manage updates for your CoreOS clusters out of the box
- HTTP Restful and Golang APIs
- Based on the [Omaha](https://code.google.com/p/omaha/wiki/ServerProtocol) protocol developed by Google
- Backend server built using Golang, dashboard built using React

More exciting features coming soon, stay tuned!

## Status

CoreRoller is currently in *beta*.

Users are encouraged to play and experiment with it, but please be advised that something may not work as expected or the API may change as the project is currently under active development.

Please report your experience and/or any bug you find as [issues](https://github.com/coreroller/coreroller/issues) on this repository.

## Getting started

The best way to give it a try is to launch a Docker container using the public images hosted in the Docker Hub. Feel free to use your favourite tool or environment to spin it up (Kitematic, etc) or just do it in the command line running this command:

	docker run -d -p 8000:8000 coreroller/demo

Once the container is up, just point your browser to:

	http://localhost:8000/
	
and you should be ready to go. Default username/password is `admin/admin`.

This demo container runs `PostgreSQL` (the datastore used by CoreRoller) and the `CoreRoller server` (aka rollerd).

In addition to this [coreroller/demo](https://hub.docker.com/r/coreroller/demo) image, there are some other images available in the docker hub that may be helpful to you in you plan to do a more serious deployment.

- **[coreroller/postgres](https://hub.docker.com/r/coreroller/postgres)**: this image runs PostgreSQL but it contains the required extensions already installed and the databases used by CoreRoller created. Do not forget to setup properly the volumes in the container to avoid any data loss.

- **[coreroller/rollerd](https://hub.docker.com/r/coreroller/rollerd)**: this image runs the backend server, a dependency free Golang binary that will power the dashboard and serve all Omaha updates and events requests.

You don't have to build these images yourself since they all have been made available in Docker Hub, and will be rebuilt automatically.

If you'd like to build one yourself - to try something for example, just do the following (let's say for rollerd):

	cd coreroller/rollerd
	docker build -t coreroller/rollerd .

You will find the Dockerfiles used to build these images in `backend/docker`. Additionally, in the `backend/systemd` directory there are some systemd unit files that might be handy in case you want to deploy CoreRoller in your CoreOS cluster using `fleet`. You can also use the sample kubernetes configuration files in the `backend/kubernetes` folder to deploy CoreRoller using `kubernetes` (`kubectl create -f backend/kubernetes`). These units and config files are just samples, feel free to adjust them to suit your specific needs.

## Managing CoreOS updates

Once you have CoreRoller up, it's time to give it some work to do. You may be interested in managing the CoreOS updates in your cluster with it. To do it, you basically need to update the server/endpoint the updater uses to pull updates from (by default the public CoreOS servers). The process is slightly different if you want to do it in existing machines or in new ones.

### New machines

In new machines, you can set up the updates server in the cloud config. Here is a small example of how to do it:

	coreos:
		update:
			group: stable
			server: https://your.coreroller.host/v1/update/

In addition to the default `stable`, `beta` and `alpha` groups, you can also create and use **custom groups** to control the updates. In that case, you can use the group id you will find next to the group name in the dashboard instead of one of the official channels.

### Existing machines

To update the update server in existing instances please edit `/etc/coreos/update.conf` and update the `SERVER` value (and optionally `GROUP` if needed):

	SERVER=https://your.coreroller.host/v1/update/
	
To apply these changes run:

	sudo systemctl restart update-engine
	
In may take a few minutes to see an update request coming through. If you want to see it sooner, you can force it running this command:

	update_engine_client -update

**Note:** the CoreUpdate docs do a great job explaining in detail how this process works and most of the information it contains is valid for CoreRoller as well, so please have a look at them [here](https://coreos.com/products/coreupdate/docs/latest/configure-machines.html) if you want more information.

## Managing updates for your own applications

In addition to manage the updates for CoreOS, you can use CoreRoller for your own applications as well. It's really easy to send updates and events requests to the Omaha server that CoreRoller provides.

In the `updaters/lib` directory there are some sample helpers that can be useful to create your own updaters that talk to CoreRoller or even embed them into your own applications. 

In the `updaters/examples` you'll find a sample minimal application built using [grace](https://github.com/facebookgo/grace) that is able to update itself using CoreRoller in a graceful way.

Some more documents, examples and updaters are on their way :)

## Contributing

CoreRoller is an Open Source project and we welcome contributions. 

Below you will find some introductory notes that should point you in the right direction to start playing with the CoreRoller source code.

### Backend

The CoreRoller backend (aka rollerd) is a Golang application. Builds and vendored dependencies are managed using [gb](http://getgb.io), so you just need a working Golang environment and gb installed to start working with it (there is **no** need to do a `go get` to fetch the dependencies as they are already contained in `backend/vendor`).

The backend source code is located inside `backend/src` and is structured as follows:

- **Package `api`**: provides functionality to do CRUD operations on all elements found in CoreRoller (applications, groups, channels, packages, etc), abstracting the rest of the components from the underlying datastore (PostgreSQL). It also controls the groups' roll-out policy logic and the instances/events registration.

- **Package `omaha`**: provides functionality to validate, handle, process and reply to Omaha updates and events requests received from the Omaha clients. It relies on the `api` package to get update packages, store events, or register instances when needed.

- **Package `syncer`**: provides some functionality to synchronize packages available in the official CoreOS channels, storing the references to them in your CoreRoller datastore. It's basically in charge of keeping up to date your the CoreOS application in your CoreRoller installation.

- **Cmd `rollerd`**: is the main backend process, exposing the functionality described above in the different packages through its http server. It provides several http endpoints used to drive most of the functionality of the dashboard as well as handling the Omaha updates and events requests received from your servers and applications.

- **Cmd `initdb`**: is just a helper to reset your database, and causing the migration to be re-run. `rollerd` will apply all database migrations automatically, so this process should only be used to wipe out all your data and start from a clean state.

Please make sure that your code is formatted using `gofmt` and makes [gometalinter](https://github.com/alecthomas/gometalinter) happy :) 

#### Backend datastore (PostgreSQL)

CoreRoller uses `PostgreSQL` as datastore, so you'll also need it if you are planning to do some work on CoreRoller. You can install it locally or use the docker image available in the docker hub (coreroller/postgres). 

If you plan to install it yourself locally, please be aware that the [semver](https://github.com/theory/pg-semver/)[1] extension is required and it's not installed by default with PostgreSQL. When installing it manually instead of using the docker image, you'll also need to run the following commands to created the necessary databases and extensions:

	psql -U postgres -c "create database coreroller;"
	psql -U postgres -c "alter database coreroller set timezone = 'UTC';"
	psql -U postgres -d coreroller -c "create extension \"uuid-ossp\";"
	psql -U postgres -d coreroller -c "create extension semver;"

To run the tests you will also need to setup the coreroller\_tests database:
	
	psql -U postgres -c "create database coreroller_tests;"
	psql -U postgres -c "alter database coreroller set timezone = 'UTC';"
	psql -U postgres -d coreroller_tests -c "create extension \"uuid-ossp\";"
	psql -U postgres -d coreroller_tests -c "create extension semver;"

**[1] UPDATE:** *as of 7 Aug 2016 the SEMVER data type that the pg-semver Postgresql extension provides is no longer used (`db/migrations/0005_remove_pgsemver.sql` migration takes care of altering the affected tables). Please note that the extension is still required to be available and installed as the first migration script `0001_initial.sql` will create some tables with fields that use the semver data type (that will be altered afterwards by the next migration scripts). This situation is just temporary to not affect existing deployments and allow a clean update using automatic database migrations, but will be improved in the future to facilitate using other Postgresql installations where the extension may not be available (such as AWS RDS) now that there is no real need for the extension.*

### Frontend

The frontend side of CoreRoller (dashboard) is a javascript web application built using `react/flux`. It's written using `ECMAScript6`, that is transformed to ES5 by Babel when the application is built. 

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

## License

CoreRoller is released under the terms of the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0).

CoreRoller was created by Cintia Sánchez García, Mathieu Lohier and Sergio Castaño Arteaga.
