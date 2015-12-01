# REX-Ray

Openly serious about storage

---

`REX-Ray` provides visibility and management of external/underlying storage
via guest storage introspection. Available as a Go package, CLI tool, and Linux
service, and with built-in third-party support for tools such as `Docker`,
`REX-Ray` is easily integrated into any workflow. For example, here's how to
list storage for a guest hosted on Amazon Web Services (AWS) with `REX-Ray`:

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
Today `REX-Ray` supports the following storage providers:

* [Amazon Elastic Computer Cloud (EC2)](https://aws.amazon.com/ec2/)
* [Openstack on Rackspace](http://www.rackspace.com/en-us/cloud/openstack)
* [Rackspace](http://www.rackspace.com)
* [ScaleIO](http://www.emc.com/storage/scaleio/index.htm)
* [XtremIO](http://xtremio.com/) (with Multipath & Device Mapper support)

`REX-Ray` also supports integration with the following platforms:

* [Docker](https://docs.docker.com/extend/plugins_volume/)

### Operating System Support
`REX-Ray` currently supports the following operating systems:

OS      | Command Line | As Service
--------|--------------|-----------
Linux   | Yes          | Yes
OS X    | Yes          | No
Windows | No           | No

## Getting Started
This section will help you get started using `REX-Ray`.

### Installing REX-Ray
The following command will download the most recent, stable build of `REX-Ray`
and install it to `/usr/bin/rexray.` On Linux systems `REX-Ray` will also be
registered as either a SystemD or SystemV service.

```bash
curl -sSL https://dl.bintray.com/emccode/rexray/install | sh -
```

`REX-Ray` can also be installed from
[a pre-built binary](./user-guide/installation.md#install-a-pre-built-binary), an RPM or DEB
package, or by
[building it from source](./user-guide/installation.md#build-and-install-from-source).

Once installed, `REX-Ray` can be used by simply typing `rexray` on the command
line, but in order for `REX-Ray` to do much more than print out help text,
configuration is necessary:

### Configuring REX-Ray
The first step to getting started is [configuring `REX-Ray`](/user-guide/config/)!

### Configuring Storage Providers
* [Amazon Elastic Computer Cloud (EC2)](/user-guide/ec2/)
* [Rackspace](/user-guide/rackspace/)
* [ScaleIO](/user-guide/scaleio/)
* [OpenStack](/user-guide/openstack/)
* [XtremIO](/user-guide/xtremio/)

### Configuring External Integration
* [Docker](/user-guide/docker/)
* [Mesos](/user-guide/mesos/)

## Getting Help
To get help with REX-Ray, please use the
[discussion group](https://groups.google.com/forum/#!forum/emccode-users),
[GitHub issues](https://github.com/emccode/rexray/issues), or tagging questions
with **EMC** at [StackOverflow](https://stackoverflow.com).

The code and documentation are released with no warranties or SLAs and are
intended to be supported through a community driven process.
