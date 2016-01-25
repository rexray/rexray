#ScaleIO

Scale-out with simplified storage management

---

## Overview
The ScaleIO driver registers a storage driver named `scaleio` with the `REX-Ray`
driver manager and is used to connect and manage ScaleIO storage.  The ScaleIO
`REST Gateway` is required for the driver to function.

## Configuration
The following is an example with all possible fields configured.  For a running
example see the `Examples` section.

```yaml
scaleio:
    endpoint:             https://host_ip/api
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
    thinOrThick:          ThinProvisioned
```

### Configuration Notes
- `insecure` should be set to `true` if you have not loaded the SSL
certificates on the host.  A successful wget or curl should be possible without
SSL errors to the API `endpoint` in this case.
- `useCerts` should only be set if you want to leverage the internal SSL
certificates.  This would be useful if you are deploying the REX-Ray binary
on a host that does not have any certificates installed.
- `systemID` takes priority over `systemName`.
- `protectionDomainID` takes priority over `protectionDomainName`.
- `storagePoolID` takes priority over `storagePoolName`.
- `thinkOrThick` determines whether to provision as the default
`ThinProvisioned`, or `ThickProvisioned`.

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

<br>
## Runtime Behavior
The `storageType` field that is configured per volume is considered the
ScaleIO Storage Pool.  This can be configured by default with the `storagePool`
setting.  It is important that you create unique names for your Storage Pools
on the same ScaleIO platform.  Otherwise, when specifying `storageType` it
may choose at random which `protectionDomain` the pool comes from.

The `availabilityZone` field represents the ScaleIO Protection Domain.

<br>
## Configuring the ScaleIO Gateway
- Install the `EMC-ScaleIO-gateway` package.
- Edit the `/opt/emc/scaleio/gateway/webapps/ROOT/WEB-INF/classes/gatewayUser.properties`
file and append the proper MDM IP addresses to the following `mdm.ip.addresses=`
parameter.
- Update the `gw_password` parameter and run the following command.
```bash
java -jar /opt/emc/scaleio/gateway/webapps/ROOT/resources/install-CLI.jar \
  --reset_password 'gw_password' \
  --config_file /opt/emc/scaleio/gateway/webapps/ROOT/WEB-INF/classes/gatewayUser.properties
```
- Start the gateway `service scaleio-gateway start`.

<br>
## Activating the Driver
To activate the ScaleIO driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `scaleio` as the driver name.

<br>
## Troubleshooting
Ensure that you are able to open a TCP connection to the gateway with the
address that you will be supplying below in the `gateway_ip` parameter.  For
example `telnet gateway_ip 443` should open a successful connection.  Removing
the `EMC-ScaleIO-gateway` package and reinstalling can force re-creation of
self-signed certs which may help resolve gateway problems.  Also try restarting
the gateway with `service scaleio-gateway restart`.

<br>
## Examples
Below is a full `rexray.yml` file that works with ScaleIO.

```yaml
rexray:
  storageDrivers:
  - scaleio
scaleio:
  endpoint: https://gateway_ip/api
  insecure: true
  userName: username
  password: password
  systemName: tenantName
  protectionDomainName: protectionDomainName
  storagePoolName: storagePoolName
```
