#Isilon

Scale-Out NAS is my fame.

---

## Overview
The Isilon driver registers a storage driver named `isilon` with the `REX-Ray`
driver manager and is used to connect and manage Isilon NAS storage.

## Configuration
The following is an example configuration of the Isilon driver.

```yaml
isilon:
  endpoint: https://endpoint:8080
  insecure: true
  username: username
  password: password
  volumePath: /rexray
  nfsHost: nfsHost
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

## Extra Paramters
The following items are configurable specific to this driver.

 - `volumePath` represents the location under `/ifs/volumes` to allow volumes to be created and removed.
 - `nfsHost` is the configurable host used when mounting exports

## Activating the Driver
To activate the XtremIO driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `isilon` as the driver name.

## Examples
Below is a full `rexray.yml` file that works with Isilon.

```yaml
rexray:
  storageDrivers:
  - isilon
isilon:
  endpoint: https://endpoint:8080
  insecure: true
  username: username
  password: password
  volumePath: /rexray
  nfsHost: nfsHost
```

## Instructions
It is expected that the `volumePath` exists already within the Isilon system.  This would reflect a directory create under `/ifs/volumes/rexray`.  It is not necessary to export this volume.

## Caveats

- This driver currently ignores the `--size` and `--volumeType` flags.
- This driver does not support pre-emption currently.  Requests from alternate hosts can have bad results.
