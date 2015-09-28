![Coreroller](https://github.com/coreroller/coreroller/raw/master/docs/screenshots/coreroller.png)

[![Build Status](https://travis-ci.org/coreroller/coreroller.svg)](https://travis-ci.org/coreroller/coreroller)

## Overview

**CoreRoller** is a set of tools to control and monitor the rollout of your updates. It's aimed to be an open source alternative to CoreOS [CoreUpdate](https://coreos.com/products/coreupdate/).

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
- Manage updates for your own applications
- Define your own groups and channels, even for the CoreOS application (pre-installed)
- Define rollout policies per group, controlling how updates should be applied for a set of instances
- Pause/resume updates at any time at the group level
- Statistics about versions installed in your instances, updates progress status, etc
- Activity stream to get notified about important events or errors
- Manage updates for your CoreOS clusters out of the box
- HTTP Restul and Golang APIs
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

This container runs `PostgreSQL` (the datastore used by CoreRoller) and the `CoreRoller server` (aka rollerd). Please **note** that this setup is just for demoing purposes. If you plan to do a more serious deployment in the *backend/docker* and *backend/systemd* folders you will find some sample Dockerfiles and systemd units that should point you in the right direction.

### CoreOS updates

Once you have CoreRoller up, it's time to give it some work to do. You may be interested in managing the CoreOS updates in your cluster with it. To do it, you basically need to update the server/endpoint they use to pull updates from (by default the public CoreOS servers). The process is slightly different if you want to do it in existing machines or in new ones.

###### New machines

In new machines, you can set up the updates server in the cloud config. Here is a small example of how to do it:

	coreos:
		update:
			group: stable
			server: https://your.coreroller.host/v1/update/

In addition to the default `stable`, `beta` and `alpha` groups, you can also create and use **custom groups** to control the updates. In that case, you can use the group id you will find next to the group name in the dashboard.

	coreos:
		update:
			group: 1a809ab1-c01c-4a6b-8ac8-6b17cb9bae08
			server: https://your.coreroller.host/v1/update/

###### Existing machines

To update the update server in existing instances please edit `/etc/coreos/update.conf` and update the `SERVER` value (and optionally `GROUP`):

	SERVER=https://your.coreroller.host/v1/update/
	
To apply these changes run:

	sudo systemctl restart update-engine
	
In may take a few minutes to see an update request coming through. If you want to see it sooner, you can force it running this command:

	update_engine_client -update

**Note:** the CoreUpdate docs do a great job explaining in detail how this process works and most of the information it contains is valid for CoreRoller as well, so please have a look at them [here](https://coreos.com/products/coreupdate/docs/latest/configure-machines.html) if you want more information.

### Managing updates for your own applications

In addition to manage the updates for CoreOS, you can use CoreRoller for your own applications as well. 

In the `updaters/lib` directory there are some sample helpers that can be useful to create your own updaters that talk to CoreRoller or even embed them into your own applications. 

In the `updaters/examples` you'll find a sample minimal application built using [grace](https://github.com/facebookgo/grace) that is able to update itself using CoreRoller in a graceful way.

Some more documents, examples and updaters are on their way :)

## Contributing

CoreRoller is an Open Source project and we welcome contributions.

Please have a look at the README.md files inside the backend and frontend directories. They introduce the architecture of CoreRoller and provide some details on how to setup the development environment to do some work on the backend or frontend respectively.

## License

CoreRoller is released under the terms of the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0).

CoreRoller was created by Cintia Sánchez García, Mathieu Lohier and Sergio Castaño Arteaga.