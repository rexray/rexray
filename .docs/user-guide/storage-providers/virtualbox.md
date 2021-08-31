# VirtualBox

Virtual storage

---


## Overview
The VirtualBox driver  registers a storage driver named `virtualbox` with the
REX-Ray driver registry and is used by VirtualBox's VMs to connect and
manage volumes provided by VirtualBox.

## Prerequisites
In order to leverage the `virtualbox` driver, the `libStorage` client or must
be located on each VM that you wish to be able to consume external volumes.
The driver leverages the `vboxwebserv` HTTP SOAP API which is a process that
must be started from the VirtualBox *host* (ie OS X) using
`vboxwebsrv -H 0.0.0.0 -v` or additionally with `-b` for running in the
background. This allows the VMs running `libStorage` to remotely make calls to
the underlying VirtualBox application. A test for connectivity can be done with
`telnet virtualboxip 18083` from the VM. The `virtualboxip` is what you
would put in the `endpoint` value.

Leveraging authentication for the VirtualBox webserver is optiona.. The HTTP
SOAP API can have authentication disabled by running
`VBoxManage setproperty websrvauthlibrary null`.

Hot-Plugging is required, which limits the usefulness of this driver to `SATA`
only.  Ensure that your VM has *pre-created* this controller and it is
named `SATA`.  Otherwise the `controllerName` field must be populated
with the name of the controller you wish to use.  The port count must be set
manually as it cannot be increased when the VMs are on.  A count of `30`
is suggested.

VirtualBox 5.0.10+ must be used.

!!! note
    For a VirtualBox VM to work successfully, the following three points are of
    the utmost importance:

    1. The SATA controller should be named `SATA`.
    2. The SATA controller's port count must allow for additional
    connections.
    3. The MAC address must not match that of any other VirtualBox VM's
    MAC addresses or the MAC address of the host.

    The REX-Ray `Vagrantfile` has
    [a section](https://github.com/AVENTER-UG/rexray/blob/master/Vagrantfile#L166-L179)
    that automatically configures these options.

## Configuration
The following is an example configuration of the VirtualBox driver.  
The `localMachineNameOrId` parameter is for development use where you force
`libStorage` to use a specific VM identity.  Choose a `volumePath` to store the
volume files or virtual disks.  This path should be created ahead of time.


```yaml
virtualbox:
  endpoint: http://virtualboxhost:18083
  userName: optional
  password: optional
  tls: false
  volumePath: $HOME/VirtualBox/Volumes
  controllerName: name
  localMachineNameOrId: forDevelopmentUse
```
For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](../servers/libstorage.md#configuration-properties).

## Activating the Driver
To activate the VirtualBox driver please follow the instructions for
[activating storage drivers](../servers/libstorage.md#storage-drivers),
using `virtualbox` as the driver name.

## Examples
Below is a working `config.yml` file that works with VirtualBox.

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: virtualbox
  server:
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

## Caveats
- Snapshot and create volume from volume functionality is not available yet
  with this driver.
- The driver supports VirtualBox 5.0.10+
