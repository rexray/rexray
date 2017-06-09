# REX-Ray [![GoDoc](https://godoc.org/github.com/codedellemc/rexray?status.svg)](http://godoc.org/github.com/codedellemc/rexray) [![Build Status](http://travis-ci.org/codedellemc/rexray.svg?branch=master)](https://travis-ci.org/codedellemc/rexray) [![Go Report Card](http://goreportcard.com/badge/codedellemc/rexray)](http://goreportcard.com/report/codedellemc/rexray) [![codecov.io](https://codecov.io/github/codedellemc/rexray/coverage.svg?branch=master)](https://codecov.io/github/codedellemc/rexray?branch=master) [ ![Download](http://api.bintray.com/packages/emccode/rexray/stable/images/download.svg) ](https://dl.bintray.com/emccode/rexray/stable/latest/)

REX-Ray provides a vendor agnostic storage orchestration engine.  The primary
design goal is to provide persistent storage for `Docker`, `Kubernetes`, and `Mesos`.

It is additionally available as a Go package, CLI tool, and Linux service which
enables it to be used for additional use cases.

## Documentation [![Docs](https://readthedocs.org/projects/rexray/badge/?version=stable)](http://rexray.readthedocs.org/en/stable/)
You will find complete documentation for REX-Ray at [rexray.readthedocs.org](http://rexray.readthedocs.org/en/stable/), including
[licensing](http://rexray.readthedocs.org/en/stable/about/license/) and
[support](http://rexray.readthedocs.org/en/stable/#getting-help) information.
Documentation provided at RTD is based on the latest stable build.  The `/.docs`
directory in this repo will refer to the latest or specific commit.

## Architecture
REX-Ray is available as a standalone process today and as a distributed
model of client-server.  The `client` performs a level abstraction of local
host processes (request for volume attachment, discovery, format, and mounting
of devices) while the `server` provides the necessary abstraction of the
control plane for multiple storage platforms/

### Storage Provider Support
The following storage providers and platforms are supported by REX-Ray.

Provider              | Storage Platform(s)
----------------------|--------------------
Amazon EC2 | [EBS](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#aws-ebs), [EFS](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#aws-efs), [S3FS](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#aws-s3fs)
Ceph | [RBD](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#ceph-rbd)
Dell EMC | [ScaleIO](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#dell-emc-scaleio), [Isilon](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#dell-emc-isilon)
DigitalOcean | [Block Storage](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#do-block-storage)
FittedCloud | [EBS Optimizer](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/#ebs-optimizer)
Google | [GCE Persistent Disk](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#gce-persistent-disk)
Microsoft | [Azure Unmanaged Disk](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#azure-ud)
OpenStack | [Cinder](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#cinder)
VirtualBox | [Virtual Media](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#virtualbox)


### Operating System Support
The following operating systems are supported by REX-Ray:

Operating System | Command Line | Service
-----------------|--------------|-----------
CentOS 7         | Yes          | Yes
CoreOS           | Yes          | Yes
RHEL 7           | Yes          | Yes
Ubuntu 14+       | Yes          | Yes

## Installation
The following command will install the REX-Ray client-server tool.  If using
`CentOS`, `Debian`, `RHEL`, or `Ubuntu` the necessary service manager is used
to bootstrap the process on startup

`curl -sSL https://dl.bintray.com/emccode/rexray/install | sh -`

## Runtime - CLI
REX-Ray can be run as an interactive CLI to perform volume management
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
Additionally, it can be run as a service to support `Docker`, `Mesos`, and other
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
$ df -h /test
```

## Runtime - Docker Plugin
Starting with Docker 1.13, Docker now supports a new plugin architecture in
which plugins can be installed as containers.

```bash
$ docker plugin install rexray/ebs EBS_ACCESSKEY=access_key EBS_SECRETKEY=secret_key
Plugin "rexray/ebs:latest" is requesting the following privileges:
 - network: [host]
 - mount: [/dev]
 - allow-all-devices: [true]
 - capabilities: [CAP_SYS_ADMIN]
Do you grant the above permissions? [y/N] y
latest: Pulling from rexray/ebs
2ef3a0b3d192: Download complete
Digest: sha256:86a3bf7fdab857c955d7ef3fb94c01e350e34ba0f7fd3d0bd485e45f1592e1c2
Status: Downloaded newer image for rexray/ebs:latest
Installed plugin rexray/ebs:latest

$ docker plugin ls
ID                  NAME                   DESCRIPTION              ENABLED
450420731dc3        rexray/ebs:latest      REX-Ray for Amazon EBS   true

$ docker run -ti --volume-driver=rexray/ebs -v test:/test busybox
$ df -h /test
```
