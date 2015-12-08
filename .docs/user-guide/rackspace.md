#Rackspace

They manage your services, we manage your storage

---

## Overview
The Rackspace driver registers a storage driver named `rackspace` with the
`REX-Ray` driver manager and is used to connect and manage storage on Rackspace
instances.

## Configuration
The following is an example configuration of the Rackspace driver.

```yaml
rackspace:
    authURL:    https://domain.com/rackspace
    userID:     0
    userName:   admin
    password:   mypassword
    tenantID:   0
    tenantName: customer
    domainID:   0
    domainName: corp
```

## Activating the Driver
To activate the Rackspace driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `rackspace` as the driver name.

## Examples
Below is a full `rexray.yml` file that works with Rackspace.

```yaml
rexray:
  storageDrivers:
  - rackspace
rackspace:
  authUrl: https://keystoneHost:35357/v2.0/
  username: username
  password: password
  tenantName: tenantName
  regionName: regionName
```
