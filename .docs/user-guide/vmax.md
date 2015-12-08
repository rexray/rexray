#VMAX

I focus on Reliability, Availability, and Servicability for Block.

---

## Overview
The VMAX driver registers a storage driver named `vmax` with the `REX-Ray`
driver manager and is used to connect and manage VMAX block storage.

This driver will in the future be used in many scenarios including HBA and
iSCSI.  Right now, the driver is functioning in the `vmh` mode.  This means
the Volume attachment and detachment is occuring as RDM's to a VM where
`REX-Ray` is running.  This use case enables you to address volumes for
containers as 1st class VMAX volumes while being capable of taking advantage
of mobility options such as `vSphere vMotion`.

## Configuration
The following is an example configuration of the Isilon driver.

```yaml
vmax:
  smisHost: smisHost
  smisPort: smisPort
  insecure: true
  username: admin
  password: password
  sid: '000000000000'
  volumePrefix: "rexray_"
  storageGroup: storageGroup
  mode: vmh
  vmh:
    host: vcenter_or_vm_host
    username: username
    password: password
    insecure: true
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

## Extra Paramters
The following items are configurable specific to this driver.

 - `sid` is your array ID, ensure you enclose it in quotes to preserve it as a
   string.
 - `volumePrefix` is an (optional) field that limits visibility and usage of
   volumes to those only with this prefix, ie. `rexray_`.
 - `storageGroup` is the fields that determines which masking view to use
   when adding and removing volumes.

## Activating the Driver
To activate the VMAX driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `vmax` as the driver name.

## Examples
Below is a full `rexray.yml` file that works with Isilon.

```yaml
rexray:
  storageDrivers:
  - vmax
  logLevel: debug
vmax:
  smisHost: smisHost
  smisPort: smisPort
  insecure: true
  username: admin
  password: password
  sid: '000000000000'
  volumePrefix: "rexray_"
  storageGroup: storageGroup
  mode: vmh
  vmh:
    host: vcenter_or_vm_host
    username: username
    password: password
    insecure: true
```

## Instructions
For the `vmh` mode, ensure that you have pre-created the storage group as
defined by `storageGroup`.  The underlying ESX hosts in the cluster where the VM
can move should already be in a masking view that has access to the VMAX ports
and should be properly logged in.

## Caveats

- This driver currently ignores the `--volumeType` flag.
- Pre-emption is not currently supported.  A mount to an alternate VM will be
  denied.
- A max of `56` devices is supported per VM.
