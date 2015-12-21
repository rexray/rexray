#Virtual Box

Just for your laptop.

---

## Overview
The VirtualBox driver registers a storage driver named `virtualbox` with the
`REX-Ray` driver manager and is used by VirtualBox VM's to to connect and
manage volumes provided by Virtual Box.

## Pre-Requisites
In order to leverage the `virtualbox` driver, REX-Ray must be located on each
VM that you wish to be able to consume external volumes.  The driver
leverages the `vboxwebserv` HTTP SOAP API which is a process that must be
started from the VirtualBox *host* (ie OS X) using `vboxwebsrv -H 0.0.0.0 -v` or
additionally with `-b` for running in the background.  This allows the VMs
running `REX-Ray` to remotely make calls to the underlying VirtualBox
application.  A test for connectivity can be done with
`telnet virtualboxip 18083` from the VM.  The `virtualboxip` is what you
would put in the `endpoint` value.

It is optional to leverage authentication.  The HTTP SOAP API can have
authentication disabled by running
`VBoxManage setproperty websrvauthlibrary null`.

Hot-Plugging is required, which limits the usefulness of this driver to `SATA`
only.  Ensure that your VM has *pre-created* this controller and it is
named `SATA`.  Otherwise the `controllerName` field must be populated
with the name of the controller you wish to use.

VirtualBox 5.0.10+ must be used.

## Configuration
The following is an example configuration of the VirtualBox driver.  
The `localMachineNameOrId` parameter is for development use where you force
REX-Ray to use a specific VM identity.  Choose a `volumePath` to store the
volume files or virtual disks.  This path should be created ahead of time.


```yaml
virtualbox:
  endpoint: http://virtualboxhost:18083
  userName: optional
  password: optional
  tls: false
  volumePath: /Users/your_user/VirtualBox Volumes
  controllerName: name
  localMachineNameOrId: forDevelopmentUse
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

## Activating the Driver
To activate the VirtualBox driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `virtualbox` as the driver name.

## Examples
Below is a working `rexray.yml` file that works with VirtualBox.

```yaml
rexray:
  storageDrivers:
  - virtualbox
virtualbox:
  endpoint: http://virtualBoxIP:18083
  volumePath: /Users/your_user/VirtualBox Volumes
```

## Caveats
- The VBoxWebSrv SOAP API changed between v4 and v5.  This functionality
  works against `4.3.28` and likely other v4 versions only.  We are
  investigating support for same features under v5.
- This driver was developed against Ubuntu 14.04.3 but should work with
  others.
- Snapshot and create volume from volume functionality is not
  available yet with this driver.
- The driver supports VirtualBox 5.0.10+
