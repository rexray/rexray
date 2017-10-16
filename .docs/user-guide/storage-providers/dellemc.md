# Dell EMC

Isilon, ScaleIO

---

<a name="dell-emc-isilon"></a>

## Isilon
The Isilon driver registers a storage driver named `isilon` with the
libStorage service registry and is used to connect and manage Isilon NAS
storage. The driver creates logical volumes in directories on the Isilon
cluster. Volumes are exported via NFS and restricted to a single client at a
time. Quotas can also be used to ensure that a volume directory doesn't exceed
a specified size.

### Configuration
The following is an example configuration of the Isilon driver. For a running
example see the [Examples](./dellemc.md#dell-emc-isilon-examples)
section.

```yaml
isilon:
  endpoint: https://endpoint:8080
  insecure: true
  username: username
  group: groupname
  password: password
  volumePath: /libstorage
  nfsHost: nfsHost
  dataSubnet: subnet
  quotas: true
```

For information on the equivalent environment variable and CLI flag names
please see the section on how configuration properties are
[transformed](../servers/libstorage.md#configuration-properties).

### Extra Parameters
The following items are configurable specific to this driver.

 * `volumePath` represents the location under `/ifs/volumes` to allow volumes to
   be created and removed.
 * `nfsHost` is the configurable NFS server hostname or IP (often a
   SmartConnect name) used when mounting exports
 * `dataSubnet` is the subnet the REX-Ray driver is running on. This is used
   for the NFS export host ACLs.

### Optional Parameters
The following items are not required, but available to this driver.

 * `insecure` defaults to `false`.
 * `group` defaults to the group of the user specified in the configuration.
   Only use this option if you need volumes to be created with a different
   group.
 * `volumePath` defaults to "". This will have all new volumes created directly
   under `/ifs/volumes`.
 * `quotas` defaults to `false`. Set to `true` if you have a SmartQuotas
   license enabled.

### Activating the Driver
To activate the Isilon driver please follow the instructions for
[activating storage drivers](../servers/libstorage.md#storage-drivers),
using `isilon` as the driver name.

<a name="dell-emc-isilon-examples"></a>

### Examples
Below is a full `config.yml` file that works with Isilon.

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: isilon
  server:
    services:
      isilon:
        driver: isilon
        isilon:
          endpoint: https://endpoint:8080
          insecure: true
          username: username
          password: password
          volumePath: /libstorage
          nfsHost: nfsHost
          dataSubnet: subnet
          quotas: true
```

### Instructions
It is expected that the `volumePath` exists already within the Isilon system.
This example would reflect a directory create under `/ifs/volumes/libstorage`
for created volumes. It is not necessary to export this volume. The `dataSubnet`
parameter is required so the Isilon driver can restrict access to attached
volumes to the host that REX-Ray is running on.

If `quotas` are enabled, a SmartQuotas license must also be enabled on the
Isilon cluster for the capacity size functionality of `libStorage` to work.

A SnapshotIQ license must be enabled on the Isilon cluster for the snapshot
functionality of `libStorage` to work.

### Caveats
The Isilon driver is not without its caveats:

 * The account used to access the Isilon cluster must be in a role with the
  following privileges:
    * Namespace Access (ISI_PRIV_NS_IFS_ACCESS)
    * Platform API (ISI_PRIV_LOGIN_PAPI)
    * NFS (ISI_PRIV_NFS)
    * Restore (ISI_PRIV_IFS_RESTORE)
    * Quota (ISI_PRIV_QUOTA)          (if `quotas` are enabled)
    * Snapshot (ISI_PRIV_SNAPSHOT)    (if snapshots are used)

<a class="headerlink hiddenanchor" name="dell-emc-scaleio"></a>

you can set the RBAC rights on the Isilon console:

```bash
 create RBAC group
isi auth roles create --name libstorage_roles
 asign privileges to role
isi auth roles modify libstorage_roles --add-priv  ISI_PRIV_NS_IFS_ACCESS
isi auth roles modify libstorage_roles --add-priv  ISI_PRIV_LOGIN_PAPI   
isi auth roles modify libstorage_roles --add-priv  ISI_PRIV_NFS       
isi auth roles modify libstorage_roles --add-priv  ISI_PRIV_IFS_RESTORE
isi auth roles modify libstorage_roles --add-priv  ISI_PRIV_QUOTA      
isi auth roles modify libstorage_roles  --add-priv  ISI_PRIV_SNAPSHOT
 add user to RBAC group
isi auth roles modify libstorage_roles --add-user libstorage
```

<a name="dell-emc-scaleio"></a>

## ScaleIO
The ScaleIO driver registers a storage driver named `scaleio` with the
libStorage service registry and is used to connect and manage ScaleIO storage.


### Requirements
 - The ScaleIO `REST Gateway` is required for the driver to function.
 - The `libStorage` client or application that embeds the `libStorage` client
   must reside on a host that has the SDC client installed. The command
   `/opt/emc/scaleio/sdc/bin/drv_cfg --query_guid` should be executable and
   should return the local SDC GUID.
 - The [official](http://www.oracle.com/technetwork/java/javase/downloads/index.html)
   Oracle Java Runtime Environment (JRE) is required. During testing, use of the
   Open Java Development Kit (JDK) resulted in unexpected errors.

### Configuration
The following is an example with all possible fields configured.  For a running
example see the [Examples](./dellemc.md#dell-emc-scaleio-examples)
section.

```yaml
scaleio:
  endpoint:             https://host_ip/api
  apiVersion:           "2.0"
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

#### Configuration Notes
- The `apiVersion` can optionally be set here to force certain API behavior.
The default is to retrieve the endpoint API, and pass this version during calls.
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
[transformed](../servers/libstorage.md#configuration-properties).

### Runtime Behavior
The `storageType` field that is configured per volume is considered the
ScaleIO Storage Pool.  This can be configured by default with the `storagePool`
setting.  It is important that you create unique names for your Storage Pools
on the same ScaleIO platform.  Otherwise, when specifying `storageType` it
may choose at random which `protectionDomain` the pool comes from.

The `availabilityZone` field represents the ScaleIO Protection Domain.

### Configuring the Gateway
- Install the `EMC-ScaleIO-gateway` package.
- Edit the
`/opt/emc/scaleio/gateway/webapps/ROOT/WEB-INF/classes/gatewayUser.properties`
file and append the proper MDM IP addresses to the following `mdm.ip.addresses=`
parameter.
- By default the password is the same as your administrative MDM password.
- Start the gateway `service scaleio-gateway start`.
 - With 1.32 we have noticed a restart of the gateway may be necessary as well
after an initial install with `service scaleio-gateway restart`.

### Activating the Driver
To activate the ScaleIO driver please follow the instructions for
[activating storage drivers](../servers/libstorage.md#storage-drivers),
using `scaleio` as the driver name.

### Troubleshooting
- Verify your parameters for `system`, `protectionDomain`, and
`storagePool` are correct.
- Verify that have the ScaleIO SDC service installed with
`rpm -qa EMC-ScaleIO-sdc`
- Verify that the following command returns the local SDC GUID
`/opt/emc/scaleio/sdc/bin/drv_cfg --query_guid`.
- Ensure that you are able to open a TCP connection to the gateway with the
address that you will be supplying below in the `gateway_ip` parameter.  For
example `telnet gateway_ip 443` should open a successful connection.  Removing
the `EMC-ScaleIO-gateway` package and reinstalling can force re-creation of
self-signed certs which may help resolve gateway problems.  Also try restarting
the gateway with `service scaleio-gateway restart`.
- Ensure that you have the correct authentication credentials for the gateway.
This can be done with a curl login. You should receive an authentication
token in return.
`curl --insecure --user admin:XScaleio123 https://gw_ip:443/api/login`
- Please review the gateway log at
`/opt/emc/scaleio/gateway/logs/catalina.out` for errors.

<a name="dell-emc-scaleio-examples"></a>

### Examples
Below is a full `config.yml` file that works with ScaleIO.

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: scaleio
  server:
    services:
      scaleio:
        driver: scaleio
        scaleio:
          endpoint: https://gateway_ip/api
          insecure: true
          userName: username
          password: password
          systemName: tenantName
          protectionDomainName: protectionDomainName
          storagePoolName: storagePoolName
```
