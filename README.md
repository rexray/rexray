# REX-Ray [![GoDoc](https://godoc.org/github.com/emccode/rexray?status.svg)](http://godoc.org/github.com/emccode/rexray) [![Build Status](http://travis-ci.org/emccode/rexray.svg?branch=master)](https://travis-ci.org/emccode/rexray) [![Go Report Card](http://goreportcard.com/badge/emccode/rexray)](http://goreportcard.com/report/emccode/rexray) [![Coverage Status](http://coveralls.io/repos/emccode/rexray/badge.svg?branch=master&service=github&i=3)](https://coveralls.io/github/emccode/rexray?branch=master) [![codecov.io](https://codecov.io/github/emccode/rexray/coverage.svg?branch=master)](https://codecov.io/github/emccode/rexray?branch=master) [ ![Download](http://api.bintray.com/packages/emccode/rexray/stable/images/download.svg) ](https://dl.bintray.com/emccode/rexray/stable/latest/)

REX-Ray provides a vendor agnostic storage orchestration engine.  The primary
design goal is to provide persistent storage for `Docker` containers as well as
`Mesos` frameworks and tasks.

It is additionally available as a Go package, CLI tool, and Linux service which
enables it to be used for additional use cases.

## Documentation [![Docs](https://readthedocs.org/projects/rexray/badge/?version=stable)](http://rexray.readthedocs.org/en/stable/)
You will find complete documentation for REX-Ray at [rexray.readthedocs.org](http://rexray.readthedocs.org/en/stable/), including
[licensing](http://rexray.readthedocs.org/en/stable/about/license/) and
[support](http://rexray.readthedocs.org/en/stable/#getting-help) information.
Documentation provided at RTD is based on the latest stable build. The `/.docs`
directory in this repo will refer to the latest or specific commit.

## Architecture
REX-Ray is available as a standalone process today and in the future (0.4)
additionally as a distributed model of client-server.  The `client` performs a
level abstraction of local host processes (request for volume attachment,
  discovery, format, and mounting of devices) while the `server` provides the
  necessary abstraction of the control plane for multiple storage platforms.

Irrespective of platform, REX-Ray provides common functionality for the
following.

Cloud platforms:
- AWS EC2 (EBS)
- Google Compute Engine
- OpenStack
 - Private Cloud
 - Public Cloud (RackSpace, and others)

Storage platforms:
 - EMC ScaleIO
  - XtremIO
  - VMAX
  - Isilon
 - Others
 - VirtualBox

## Operating System Support
By default we prescribe the curl-bash method of installing REX-Ray.  Other
methods are available, please consult the documentation for more information.


We explicitly support the following operating system distributions.
- Ubuntu
- Debian
- RedHat
- CentOS
- CoreOS
- OSX
- TinyLinux (boot2docker)

## Installation
The following command will install the REX-Ray client-server tool.  If using
`CentOS`, `RedHat`, `Ubuntu`, or `Debian` the necessary service manager is used
to bootstrap the process on startup.  

`curl -sSL https://dl.bintray.com/emccode/rexray/install | sh -`

## Runtime - CLI
REX-Ray can be ran as an interactive CLI to perform volume management
capabilities.

```bash
$ export REXRAY_SERVICE=ebs
$ export EBS_ACCESSKEY=access_key
$ export EBS_SECRETKEY=secret_key
$ rexray volume ls
ID            Name  Status    Size
vol-6ac6c7d6        attached  8
```

## Runtime - Service (Docker)
Additionally, it can be ran as a service to support `Docker`, `Mesos`, and other
 platforms that can communicate through `HTTP/JSON`.

```bash
$ export REXRAY_SERVICE=ebs
$ export EBS_ACCESSKEY=access_key
$ export EBS_SECRETKEY=secret_key
$ rexray service start
Starting REX-Ray...SUCCESS!

  The REX-Ray daemon is now running at PID XX. To
  shutdown the daemon execute the following command:

    sudo /usr/bin/rexray stop

$ docker run -ti --volume-driver=rexray -v test:/test busybox
$ df /test
```
