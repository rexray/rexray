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
