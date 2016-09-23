# REX-Ray

Openly serious about storage

---
REX-Ray delivers persistent storage access for container runtimes, such as
Docker and Mesos, and provides an easy interface for enabling advanced storage
functionality across common storage, virtualization and cloud platforms. For
example, here's how to list the volumes for a VirtualBox VM running Linux with
REX-Ray:

```sh
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
REX-Ray is a storage orchestration tool that provides a set of common commands
for managing multiple storage platforms. Built on top of the
[libStorage](http://libstorage.readthedocs.io/en/stable) framework, REX-Ray
enables persistent storage for container runtimes such as Docker and
Mesos.

!!! note "note"

    The initial REX-Ray 0.4.x release omits support for several,
    previously verified storage platforms. These providers will be
    reintroduced incrementally, beginning with 0.4.1. If an absent driver
    prevents the use of REX-Ray, please continue to use 0.3.3 until such time
    the storage platform is introduced in REX-Ray 0.4.x. Instructions on how
    to [install](./user-guide/installation.md#rex-ray-033) and
    [configure](http://rexray.readthedocs.io/en/v0.3.3) REX-Ray 0.3.3 are both
    available.

### Storage Provider Support
The following storage providers and platforms are supported by REX-Ray.

Provider              | Storage Platform(s)
----------------------|--------------------
EMC | [ScaleIO](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#scaleio), [Isilon](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#isilon)
[Oracle VirtualBox](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#virtualbox) | Virtual Media
Amazon EC2 | [EBS](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#aws-ebs), [EFS](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#aws-efs)

Support for the following storage providers will be reintroduced in upcoming
releases:

Provider              | Storage Platform(s)
----------------------|--------------------
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

```sh
$ curl -sSL https://dl.bintray.com/emccode/rexray/install | sh
```

Refer to the User Guide's
[installation topic](./user-guide/installation/#installation) for instructions
on building REX-Ray from source, installing specific versions, and more.


### Configuring REX-Ray
REX-Ray requires a configuration file for storing details used to communicate
with storage providers. This can include authentication credentials and
driver-specific configuration options. Refer to the libStorage Storage Providers
[documentation](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/)
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

```sh
$ rexray --help
```

To verify the configuration file, use REX-Ray to list the volumes:

```sh
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

```sh
$ rexray service start
```

### Hello REX-Ray
In the grand tradition of technical documentation, the first true end-to-end
example of REX-Ray is called `Hello REX-Ray`. It showcases a two-node
deployment with the first node configured as a REX-Ray/libStorage server and
the second node as merely a client. Both nodes have Docker (1.11+) installed
and configured to leverage REX-Ray for persistent storage.

The below example does have a few requirements:

 * VirtualBox 5.0+
 * Vagrant 1.8+
 * Ruby 2.0+

#### Start REX-Ray Vagrant Environment
Before bringing the Vagrant environment online, please ensure it is
accomplished in a clean directory:

```sh
$ cd $(mktemp -d)
```

Inside the newly created, temporary directory, download the REX-Ray
[Vagrantfile](https://github.com/emccode/rexray/blob/master/Vagrantfile):

```sh
$ curl -fsSLO https://raw.githubusercontent.com/emccode/rexray/master/Vagrantfile
```

Now it is time to bring the REX-Ray environment online:

!!! note "note"

    The next step could potentially open up the system on which the command
    is executed to security vulnerabilities. The Vagrantfile brings the
    VirtualBox web service online if it is not already running. However,
    in the name of simplicity the Vagrantfile also disables the web server's
    authentication module. Please do not disable authentication for the
    VirtualBox web server if this example is being executed on an open network
    or without some type of firewall in place.

```sh
$ vagrant up
```

The above command should result in output similar to
[this Gist](https://gist.github.com/akutz/13fc3b2237ea2c295a25c2e367e6bd8f).

Once the command has been completed successfully there will be two VMs online
named `node0` and `node1`. Both nodes are running Docker and REX-Ray with
`node0` configured to act as a libStorage server.

Now that the environment is online it is time to showcase Docker leveraging
REX-Ray to create persistent storage as well as illustrating REX-Ray's
distributed deployment capabilities.

#### Node 0
First, SSH into `node0`

```sh
$ vagrant ssh node0
```

From `node0` use Docker with REX-Ray to create a new volume named
`hellopersistence`:

```sh
vagrant@node0:~$ docker volume create --driver rexray --opt size=1 \
                 --name hellopersistence
```

After the volume is created, mount it to the host and container using the
`--volume-driver` and `-v` flag in the `docker run` command:

```sh
vagrant@node0:~$ docker run -tid --volume-driver=rexray \
                 -v hellopersistence:/mystore \
                 --name temp01 busybox
```

Create a new file named `myfile` on the file system backed by the persistent
volume using `docker exec`:

```sh
vagrant@node0:~$ docker exec temp01 touch /mystore/myfile
```

Verify the file was successfully created by listing the contents of the
persistent volume:

```sh
vagrant@node0:~$ docker exec temp01 ls /mystore
```

Remove the container that was used to write the data to the persistent volume:

```sh
vagrant@node0:~$ docker rm -f temp01
```

Finally, exit the SSH session to `node0`:

```sh
vagrant@node0:~$ exit
```

#### Node 1
It's time to connect to `node1` and use the volume `hellopersistence` that was
created in the previous section from `node0`.

!!! note "note"

    While `node1` runs both the Docker and REX-Ray services like `node0`, the
    REX-Ray service on `node1` in no way understands or is configured for the
    VirtualBox storage driver. All interactions with the VirtualBox web service
    occurs via `node0`'s libStorage server with which `node1` communicates.

Use the vagrant command to SSH into `node1`:

```sh
$ vagrant ssh node1
```

Next, create a new container that mounts the existing volume,
`hellopersistence`:

```sh
vagrant@node1:~$ docker run -tid --volume-driver=rexray \
                 -v hellopersistence:/mystore \
                 --name temp01 busybox
```

The next command validates the file `myfile` created from `node0` in the
previous section has persisted inside the volume across machines:

```sh
vagrant@node1:~$ docker exec temp01 ls /mystore
```

Finally, exit the SSH session to `node1`:

```sh
vagrant@node1:~$ exit
```

#### Cleaning Up
Be sure to kill the VirtualBox web server with a quick `killall vboxwebsrv` and
to tear down the Vagrant environment with `vagrant destroy`. Omitting these
commands will leave the web service and REX-Ray Vagrant nodes online and
consume additional system resources.

#### Congratulations
REX-Ray has been used to provide persistence for stateless containers! Examples
using MongoDB, Postgres, and more with persistent storage can be found at
[Application Examples](./user-guide/applications.md).

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
