#XtremIO

Not just a flash in the pan

---

## Overview
The XtremIO registers a storage driver named `xtremio` with the `REX-Ray`
driver manager and is used to connect and manage XtremIO storage.

## Configuration Options
The following are the configuration options for the `xtremio` storage driver.

 EnvVar | YAML | CLI  
--------|------|------
`GOXTREMIO_ENDPOINT` | `xtremIoEndpoint` | `--xtremIoEndpoint`
`GOXTREMIO_USERNAME` | `xtremIoUserName` | `--xtremIoUserName`
`GOXTREMIO_PASSWORD` | `xtremIoPassword` | `--xtremIoPassword`
`GOXTREMIO_INSECURE` | `xtremIoInsecure` | `--xtremIoInsecure`
`GOXTREMIO_DM` | `xtremIoDeviceMapper` | `--xtremIoDeviceMapper`
`GOXTREMIO_MULTIPATH` | `xtremIoMultipath` | `--xtremIoMultipath`
`GOXTREMIO_REMOTEMANAGEMENT` | `xtremIoRemoteManagement` | `--xtremIoRemoteManagement`

## Activating the Driver
To activate the XtremIO driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `xtremio` as the driver name.

## Examples
