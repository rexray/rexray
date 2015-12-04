#Google Compute Engine

The Google Cloud

---

## Overview
The GCE registers a storage driver named `gce` with the `REX-Ray`
driver manager and is used to connect and manage Google Compute Engine storage.

## Pre-Requisites
In order to leverage the GCE driver, REX-Ray must be located on the
running GCE instane that you wish to receive storage.  There must also
be a `json key` file for the credentials that can be retrieved from the [API
portal](https://console.developers.google.com/apis/credentials).

## Configuration
The following is an example configuration of the GCE driver.

```yaml
gce:
  keyfile: path_to_json_key
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

## Activating the Driver
To activate the XtremIO driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `gce` as the driver name.

## Examples
Below is a full `rexray.yml` file that works with GCE.

```yaml
rexray:
  storageDrivers:
  - gce
gce:
  keyfile: /certdir/cert.json
```

## Configurable Items
The following items are configurable specific to this driver.
- [volumeTypes](https://cloud.google.com/compute/docs/reference/latest/diskTypes/list)

## Limitations
- Debian 8.2 forced mounts via pre-emption results in Input/Output error until
remounted
