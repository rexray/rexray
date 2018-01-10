# OpenStack

Cinder

---

## Cinder
The Cinder driver registers a storage driver named `cinder` with the
`libStorage` driver manager and is used to connect and manage storage on
Cinder-compatible instances.

### Configuration
The following is an example configuration with most fields populated for
illustration. For a running example see the
[Examples](#openstack-cinder-examples) section.

```yaml
cinder:
  authURL:              https://domain.com/openstack
  userID:               0
  userName:             myusername
  password:             mypassword
  tenantID:             0
  tenantName:           customer
  domainID:             0
  domainName:           corp
  regionName:           USNW
  availabilityZoneName: Gold
  attachTimeout:        1m
  createTimeout:        10m
  deleteTimeout:        10m
```

#### Configuration Notes
- `regionName` is optional, it should be empty if you only have one region.
- `availabilityZoneName` is optional, the volume will be created in the default
availability zone if not specified.

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](../servers/libstorage.md#configuration-properties).

### Activating the Driver
To activate the Cinder driver please follow the instructions for
[activating storage drivers](../servers/libstorage.md#storage-drivers),
using `cinder` as the driver name.

<a name="openstack-cinder-examples"></a>

### Examples
Below is a full `config.yml` file that works with Cinder.

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: cinder
  server:
    services:
      cinder:
        driver: cinder
cinder:
  authUrl: https://keystoneHost:35357/v2.0/
  username: username
  password: password
  tenantName: tenantName
```
