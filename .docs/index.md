# REX-Ray

Openly serious about storage

---
REX-Ray delivers persistent storage access for container runtimes, such as
Docker and Mesos, and provides an easy interface for enabling advanced storage
functionality across common storage, virtualization and cloud platforms. For
example, here's how to list the volumes for a VirtualBox VM running Linux with
REX-Ray:

```bash
$ rexray volume --service virtualbox
- attachments:
  - instanceID:
      id: e71578b0-1bfb-4fa5-bcd5-4ae982fd4a9b
      driver: virtualbox
    status: /Users/akutz/VirtualBox/libStorage/libStorage.vmdk
    volumeID: 1b819454-a280-4cff-aff5-141f4e8fd154
  name: libStorage.vmdk
  size: 64
  status: /Users/akutz/VirtualBox/libStorage/libStorage.vmdk
  id: 1b819454-a280-4cff-aff5-141f4e8fd154
  type: ""
```

## Overview
REX-Ray is an abstraction layer between storage endpoints and container
platforms. The administration and orchestration of various storage platforms
can all be performed using the same set of commands.

!!! note "note"

    The initial REX-Ray 0.4.x release omits support for several,
    previously verified storage platforms. These providers will be
    reintroduced incrementally, beginning with 0.4.1. If an absent driver
    prevents the use of REX-Ray, please continue to use 0.3.3 until such time
    the storage platform is introduced in REX-Ray 0.4.x. Instructions on how
    to install REX-Ray 0.3.3 may be found
    [here](./user-guide/installation.md#rex-ray-033).

### Storage Provider Support
The following storage providers and platforms are supported by REX-Ray.

Provider              | Storage Platform(s)
----------------------|--------------------
EMC | [ScaleIO](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#scaleio), [Isilon](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#isilon)
[Oracle VirtualBox](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#virtualbox) | Virtual Media

Support for the following storage providers will be reintroduced in upcoming
releases:

Provider              | Storage Platform(s)
----------------------|--------------------
[Amazon EC2](./user-guide/storage-providers.md#coming-soon) | EBS
[Google Compute Engine](./user-guide/storage-providers.md#coming-soon) | Disk
[Open Stack](./user-guide/storage-providers.md#coming-soon) | Cinder
[Rackspace](./user-guide/storage-providers.md#coming-soon) | Cinder
EMC | [XtremIO](./user-guide/storage-providers.md#coming-soon), [VMAX](./user-guide/storage-providers.md#coming-soon)

### Operating System Support
The following operating systems (OS) are supported by REX-Ray:

OS             | Command Line | Service
---------------|--------------|-----------
Ubuntu 12+     | Yes          | Yes
Debian 6+      | Yes          | Yes
RedHat         | Yes          | Yes
CentOS 6+      | Yes          | Yes
CoreOS         | Yes          | Yes
TinyLinux (boot2docker)| Yes          | Yes
OS X Yosemite+ | Yes          | No
Windows        | No           | No

### Container Platform Support
REX-Ray currently supports the following container platforms:

Platform            | Use
------------------|-------------------------
Docker            | [Volume Driver Plugin](./user-guide/schedulers.md#docker)
Mesos             | [Volume Driver Isolator module](./user-guide/schedulers.md#mesos)
Mesos + Docker    | [Volume Driver Plugin](./user-guide/schedulers.md#mesos)

## Getting Started
This section will help in getting REX-Ray up and running quickly.

### Installing REX-Ray
The following command will download the most recent, stable build of REX-Ray
and install it to `/usr/bin/rexray` on Linux systems. REX-Ray will be
registered as either a SystemD or SystemV service depending upon the OS.

```bash
$ curl -sSL https://dl.bintray.com/emccode/rexray/install | sh
```

Refer to the User Guide's
[installation topic](./user-guide/installation/#installation) for instructions
on building REX-Ray from source, installing specific versions, and more.


### Configuring REX-Ray
REX-Ray requires a configuration file for storing details used to communicate
with storage providers. This can include authentication credentials and
driver-specific configuration options. Refer to the
[libStorage Storage Providers documentation](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/)
for sample configurations of all supported storage platforms. Additionally, look
at [core properties](./user-guide/config.md#configuration-properties) &
[logging](./user-guide/config.md#logging-configuration) for advanced
configurations.

Create a configuration file on the host at `/etc/rexray/config.yml`. Here is a
simple example for using Oracle VirtualBox:

```yaml
libstorage:
  service: virtualbox
```

Refer to the
[VirtualBox documentation](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/#virtualbox)
for additional VirtualBox configuration options.

From here, REX-Ray can now be used as a command line tool. View the commands
available:

```bash
$ rexray --help
```

To verify the configuration file, use REX-Ray to list the volumes:

```bash
$ rexray volume
- attachments:
  - instanceID:
      id: e71578b0-1bfb-4fa5-bcd5-4ae982fd4a9b
      driver: virtualbox
    status: /Users/akutz/VirtualBox/libStorage/libStorage.vmdk
    volumeID: 1b819454-a280-4cff-aff5-141f4e8fd154
  name: libStorage.vmdk
  size: 64
  status: /Users/akutz/VirtualBox/libStorage/libStorage.vmdk
  id: 1b819454-a280-4cff-aff5-141f4e8fd154
  type: ""
```

If there is an error, use the `-l debug` flag and consult
[debugging instructions](/#getting-help).

### Start REX-Ray as a Service
Container platforms rely on REX-Ray to be running as a service to function
properly. For instance, Docker communicates to the REX-Ray Volume Driver via
a UNIX socket file.

```bash
$ rexray service start
```

### REX-Ray with Docker
Docker 1.10+ is recommended to use REX-Ray as a
[Docker Volume Driver Plugin](https://docs.docker.com/extend/plugins_volume/).

The following example uses two Amazon EC2 Virtual Machines, `EC2a` and `EC2b`,
that reside within the same Availability Zone.

From `EC2a`, create a new volume called `hellopersistence`. After the new
volume is created, mount the volume to the host and container using
the `--volume-driver` and `-v` flag in the `docker run` command. Create a new
file called `myfile` using `docker exec` that will be persisted throughout the
example. Lastly, stop and remove the container so it no longer exists:

```bash
$ docker volume create --driver rexray --opt size=10 --name hellopersistence
$ docker run -tid --volume-driver=rexray -v hellopersistence:/mystore \
  --name temp01 busybox
$ docker exec temp01 touch /mystore/myfile
$ docker exec temp01 ls /mystore
$ docker rm -f temp01
```

From `EC2b`, create a new container that mounts the pre-existing volume and
verify `myfile` that was originally created from host `EC2a` has persisted.
```bash
$ docker run -tid --volume-driver=rexray -v hellopersistence:/mystore \
  --name temp01 busybox
$ docker exec temp01 ls /mystore
```

Congratulations, REX-Ray has been used to provide persistence for stateless
containers!

Examples using MongoDB, Postgres, and more with persistent storage can be found
at [Application Examples](./user-guide/application.md).

## Getting Help
Having issues? No worries, let's figure it out together.

### Debug
The `-l debug` flag can be appended to any command in order to get verbose
output. The following command will list all of the volumes visible to REX-Ray
with debug logging enabled:

```
$ rexray volume -l debug
```

For an example of the full output from the above command, please refer to this
[Gist](https://gist.github.com/akutz/df2afe2dc43f75b67b8977f398095ed7).

### GitHub and Slack
If a little extra help is needed, please don't hesitate to use
[GitHub issues](https://github.com/emccode/rexray/issues) or join the active
conversation on the
[EMC {code} Community Slack Team](http://community.emccode.com/) in
the #project-rexray channel
