# REX-Ray

Openly serious about storage

---
`REX-Ray` delivers persistent storage access for container runtimes, such as
Docker and Mesos, and provides an easy interface for enabling advanced storage
functionality age, virtualization and cloud platforms. For
example, here's how to list storage for a guest hosted on Amazon Web Services
(AWS) with `REX-Ray`:

```bash
$ export REXRAY_STORAGEDRIVERS=ec2
$ export AWS_ACCESSKEY=access_key
$ export AWS_SECRETKEY=secret_key
$ rexray volume get

- providername: ec2
  instanceid: i-695bb6ab
  volumeid: vol-dedbadc3
  devicename: /dev/sda1
  region: us-west-1
  status: attached
```

## Overview
`REX-Ray` is an abstraction layer between storage endpoints and container
platforms. The administration and orchestration of various storage platforms
can all be performed using the same set of commands.

### Storage Provider Support
The following storage providers and platforms are supported by `REX-Ray`.

Provider              | Storage Platform(s)
----------------------|-------------------------
[Amazon EC2](/user-guide/storage-providers/ec2/) | EBS
[Google Compute Engine](/user-guide/storage-providers/gce) | Disk
[Open Stack](/user-guide/storage-providers/openstack) | Cinder
[Rackspace](/user-guide/storage-providers/rackspace) | Cinder
EMC                   | [ScaleIO](/user-guide/storage-providers/scaleio), [XtremIO](/user-guide/storage-providers/xtremio), [VMAX](/user-guide/storage-providers/vmax), [Isilon](/user-guide/storage-providers/isilon)
[Virtual Box](/user-guide/storage-providers/virtualbox)          | Virtual Media

### Operating System Support
The following operating systems (OS) are supported by `REX-Ray`:

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
`REX-Ray` currently supports the following container platforms:

Platform            | Use
------------------|-------------------------
Docker            | [Volume Driver Plugin](/user-guide/schedulers#docker)
Mesos             | [Volume Driver Isolator module](/user-guide/schedulers#mesos)
Mesos + Docker    | [Volume Driver Plugin](/user-guide/schedulers#mesos)

## Getting Started
This section will help you get REX-Ray up and running quickly. For more advanced
configurations including
[core properties](/user-guide/config/#configuration-properties) and additional
storage providers use the `User Guide` menu in the tool-bar.

### Installing REX-Ray
The following command will download the most recent, stable build of REX-Ray
and install it to `/usr/bin/rexray` on Linux systems. REX-Ray will be
registered as either a SystemD or SystemV service depending upon the OS.

```bash
$ curl -sSL https://dl.bintray.com/emccode/rexray/install | sh -
```

### Configuring REX-Ray
Create a configuration file on the host at `/etc/rexray` in YAML format called
`config.yml` (this file can be created with `vi` or transferred over via `scp`
  or `ftp`). Here is a simple example for using Amazon EC2:
```
rexray:
  storageDrivers:
  - ec2
aws:
  accessKey: MyAccessKey
  secretKey: MySecretKey
```

From here, REX-Ray can now be used as a command line tool. View the commands
available:
```bash
$ rexray --help
```

To verify the configuration file is being accessed and the AWS Keys are being
used, list volumes that can be accessed:
```bash
$ rexray volume ls
```

If there is an error, use the `-l debug` flag and consult debugging instructions
 located under [Getting Help](/#getting-help). If nothing is returned using `ls`
, then everything is functioning as expected

### Start REX-Ray as a Service
Container platforms rely on REX-Ray to be running as a service to function
properly. For instance, Docker communicates to the REX-Ray Volume Driver through
a Unix socket file.

```bash
$ rexray start
```

### REX-Ray with Docker
Docker 1.10+ is recommended to use REX-Ray as a
[Docker Volume Driver Plugin](https://docs.docker.com/extend/plugins_volume/).

The following example uses two Amazon EC2 Virtual Machines, `EC2a` and `EC2b`,
that reside within the same Availability Zone.

From `EC2a`, create a new volume called `hellopersistence`. After the new volume
 is created, mount the volume to the host and container using the
 `--volume-driver` and `-v` flag in the `docker run` command. Create a new file
called `myfile` using `docker exec` that will be persisted throughout the
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

Congratulations, you have used `REX-Ray` to provide persistence for stateless
containers!

Examples using MongoDB, Postgres, and more with persistent storage can be found
at [Application Examples](/user-guide/application/).

## Getting Help
Having issues? No worries, let's figure it out together.

### Debug
The `debug` flag can be appended to any command in order to get verbose output:

```
$ rexray volume -l debug
```

The above command will list all of the volumes visible to `REX-Ray` with debug
logging enabled.

### GitHub and Slack
And if you need a little extra help, please don't hesitate to use
[GitHub issues](https://github.com/emccode/rexray/issues) or join the active
conversation on the
[EMC {code} Community Slack Team](http://community.emccode.com/) in
the #project-rexray channel
