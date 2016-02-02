# REX-Ray

Openly serious about storage

---
`REX-Ray` delivers persistent storage access for container runtimes including
those provided by Docker, Mesos and others. It is designed to enable advanced
storage functionality across common storage, virtualization and cloud platforms.

<br>
## Overview
REX-Ray is an abstraction layer between storage endpoints and container
platforms. The administration and orchestration of various storage platforms
can all be performed using the same set of commands.

### Storage Provider Support

Storage provider support for the following storage platforms:

Provider              | Storage Platform(s)
----------------------|-------------------------
Amazon EC2            | EBS
Google Compute Engine | Disk
Open Stack            | Cinder
EMC                   | ScaleIO, XtremIO, VMAX, Isilon
Virtual Box           | Virtual Media

### Operating System Support

OS support for the following operating systems:

OS       | Command Line | As Service
---------|--------------|-----------
Ubuntu   | Yes          | Yes
Debian   | Yes          | Yes
RedHat   | Yes          | Yes
CentOS   | Yes          | Yes
CoreOS   | Yes          | Yes
OS X     | Yes          | No
Windows  | No           | No

### Container Platform Support

Integration support documented with the following platforms:

Platform            | Use
------------------|-------------------------
Docker            | [Volume Driver Plugin](/user-guide/third-party/docker/)
Mesos             | [Volume Driver Isolator module](/user-guide/third-party/mesos/)
Mesos + Docker    | [Volume Driver Plugin](/user-guide/third-party/mesos/)

<br>
## Quick Start (using Docker)
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

### Example of REX-Ray with Docker
Docker 1.9.1+ is recommended to use REX-Ray as a
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
$ docker run -d --volume-driver=rexray -v hellopersistence:/mystore \
  --name temp01 busybox
$ docker exec temp01 "touch /mystore/myfile"
$ docker exec temp01 "ls /mystore"
$ docker rm -f temp01
```

From `EC2b`, create a new container that mounts the pre-existing volume and
verify `myfile` that was originally created from host `EC2a` has persisted.
```bash
$ docker run -d --volume-driver=rexray -v hellopersistence:/mystore \
  --name temp01 busybox
$ docker exec temp01 "ls /mystore"
```

Congratulations, you have used `REX-Ray` to provide persistence for stateless
containers!

Examples using MongoDB, Postgres, and more with persistent storage can be found
at [Application Examples](/user-guide/application/).

<br>
## Getting Help

### Debug
Use the `debug` flag after any command to get verbose output.
`rexray create -l debug`.

### GitHub and Slack
To get help with REX-Ray, please use
[GitHub issues](https://github.com/emccode/rexray/issues) or join the active
conversation on the
[EMC {code} Community Slack Team](http://community.emccode.com/) in
the #project-rexray channel

### License
The code and documentation are released with no warranties or SLAs and are
intended to be supported through a community driven process. Licensed under the
Apache License, Version 2.0 (the “License”)
([License](/about/license/#rex-ray-license))
