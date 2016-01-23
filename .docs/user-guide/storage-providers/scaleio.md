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
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

## Configuring the Gateway

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

## Activating the Driver
To activate the ScaleIO driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `scaleio` as the driver name.

## Troubleshooting
Ensure that you are able to open a TCP connection to the gateway with the
address that you will be supplying below in the `gateway_ip` parameter.  For
example `telnet gateway_ip 443` should open a successful connection.  Removing
the `EMC-ScaleIO-gateway` package and reinstalling can force re-creation of
self-signed certs which may help resolve gateway problems.  Also try restarting
the gateway with `service scaleio-gateway restart`.

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
