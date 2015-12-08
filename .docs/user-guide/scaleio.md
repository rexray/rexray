#ScaleIO

Scale-out with simplified storage management

---

## Overview
The ScaleIO registers a storage driver named `scaleio` with the `REX-Ray`
driver manager and is used to connect and manage ScaleIO storage.

## Configuration
The following is an example configuration of the ScaleIO driver.

```yaml
scaleio:
    endpoint:             https://domain.com/scalio
    insecure:             false
    useCerts:             true
    userName:             admin
    password:             mypassword
    systemID:             0
    systemName:           sysv
    protectionDomainID:   0
    protectionDomainName: corp
    storagePoolID:        0
    storagePoolName:      gold
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

## Activating the Driver
To activate the ScaleIO driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `scaleio` as the driver name.

## Examples
Below is a full `rexray.yml` file that works with ScaleIO.

```yaml
rexray:
  storageDrivers:
  - scaleio
rackspace:
  endpoint: endpoint
  insecure: true
  userName: username
  password: password
  systemName: tenantName
  protectionDomainName: protectionDomainName
  storagePoolName: storagePoolName
```
