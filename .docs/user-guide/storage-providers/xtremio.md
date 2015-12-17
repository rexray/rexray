#XtremIO

Not just a flash in the pan

---

## Overview
The XtremIO registers a storage driver named `xtremio` with the `REX-Ray`
driver manager and is used to connect and manage XtremIO storage.

## Configuration
The following is an example configuration of the XtremIO driver.

```yaml
xtremio:
    endpoint:         https://domain.com/xtremio
    userName:         admin
    password:         mypassword
    insecure:         false
    deviceMapper:     false
    multipath:        true
    remoteManagement: false
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

## Activating the Driver
To activate the XtremIO driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `xtremio` as the driver name.

## Examples
Below is a full `rexray.yml` file that works with XtremIO.

```yaml
rexray:
  storageDrivers:
  - xtremio
xtremio:
  endpoint: endpoint
  insecure: true
  username: admin
  password: password
  multipath: true
```

## Pre-Requisites
### Install
The driver currently is built for iSCSI operations with XtremIO.  It is expected
that connectivity between the host and XIO has been established.  The following
packages can be used for this.  `open-scsi` provides the iSCSI connectivity.
`multipath-tools` enables multi-path behavior and relates to the `multipath`
flag if installed.

- `apt-get install open-iscsi`
- `apt-get install multipath-tools`
- `iscsiadm -m discovery -t st -p 192.168.1.61`
- `iscsiadm -m node -l`

### XIO
Once a login has occured, then you should be able to create a initiator
group for this iSCSI IQN.  You can leverage default naming for the initiator
and group.
