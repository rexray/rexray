#Isilon

Scale-Out NAS is my fame.

---

## Overview
The Isilon driver registers a storage driver named `isilon` with the `REX-Ray`
driver manager and is used to connect and manage Isilon NAS storage.  The driver creates logical volumes in directories on the Isilon cluster.  Volumes are exported via NFS and restricted to a single client at a time.  Quotas can also be used to ensure that a volume directory doesn't exceed a specified size.

## Configuration
The following is an example configuration of the Isilon driver.

```yaml
isilon:
  endpoint: https://endpoint:8080
  insecure: true
  username: username
  group: groupname
  password: password
  volumePath: /rexray
  nfsHost: nfsHost
  dataSubnet: subnet
  quotas: true
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

## Extra Paramters
The following items are configurable specific to this driver.

 - `volumePath` represents the location under `/ifs/volumes` to allow volumes to be created and removed.
 - `nfsHost` is the configurable host used when mounting exports
 - `dataSubnet` is the subnet the REX-Ray driver is running on

## Optional Paramters
The following items are not required, but available to this driver.

 - `insecure` defaults to `false`.
 - `group` defaults to the group of the user specified in the configuration.  Only use this option if you need volumes to be created with a different group.
 - `volumePath` defaults to "".  This will have all new volumes created directly under `/ifs/volumes`.
 - `quotas` defaults to `false`.  Set to `true` if you have a SmartQuotas license enabled.

## Activating the Driver
To activate the Isilon driver please follow the instructions for
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
  dataSubnet: subnet
  quotas: true
```

## Instructions
It is expected that the `volumePath` exists already within the Isilon system.  This example would reflect a directory create under `/ifs/volumes/rexray` for created volumes.  It is not necessary to export this volume.  The `dataSubnet` parameter is required so the Isilon driver can restrict access to attached volumes to the host that REX-Ray is running on.

If `quotas` are enabled, a SmartQuotas license must also be enabled on the Isilon cluster for the capacity size functionality of REX-Ray to work.

## Caveats

- This driver currently ignores the `--volumeType` flag.
- The account used to access the Isilon cluster must be in a role with these privileges:
	- Namespace Access (ISI_PRIV_NS_IFS_ACCESS)
	- Platform API (ISI_PRIV_LOGIN_PAPI)
	- NFS (ISI_PRIV_NFS)
	- Restore (ISI_PRIV_IFS_RESTORE)
	- Quota (ISI_PRIV_QUOTA)    (if `quotas` are enabled)

