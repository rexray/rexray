# Storage Providers

Connecting storage and platforms...

---

## Overview
This page reviews the storage providers and platforms supported by `REX-Ray`.

## Amazon EC2
The Amazon EC2 driver registers a storage driver named `ec2` with the `REX-Ray`
driver manager and is used to connect and manage storage on EC2 instances. The
EC2 driver is made possible by the
[goamz project](https://github.com/mitchellh/goamz).

### Configuration
The AWS EC2 driver can be configured several ways. In descending order of precedence:

1. REX-Ray config
```yaml
aws:
    accessKey: MyAccessKey
    secretKey: MySecretKey
    region:    USNW
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

1. Shared credentials file: `~/.aws/credentials`, or `C:\Users\USERNAME\.aws\credentials` on Windows
1. Environment variables: `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`
1. IAM Instance Profile

The last three configuration methods are documented [here](https://docs.aws.amazon.com/amazonswf/latest/awsrbflowguide/set-up-creds.html).

### Activating the Driver
To activate the EC2 driver please follow the instructions for
[activating storage drivers](/user-guide/config#storage-drivers),
using `ec2` as the driver name.

### Examples
Below is a full `config.yml` file that works with OpenStack.

```yaml
rexray:
  storageDrivers:
  - ec2
aws:
    accessKey: MyAccessKey
    secretKey: MySecretKey
```

## Google Compute Engine
The Google Compute Engine (GCE) registers a storage driver named `gce` with the
`REX-Ray` driver manager and is used to connect and manage GCE storage.

### Prerequisites
In order to leverage the GCE driver, REX-Ray must be located on the
running GCE instance that you wish to receive storage.  There must also
be a `json key` file for the credentials that can be retrieved from the [API
portal](https://console.developers.google.com/apis/credentials).

### Configuration
The following is an example configuration of the GCE driver.

```yaml
gce:
  keyfile: path_to_json_key
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

### Activating the Driver
To activate the GCE driver please follow the instructions for
[activating storage drivers](/user-guide/config#storage-drivers),
using `gce` as the driver name.

### Examples
Below is a full `config.yml` file that works with GCE.

```yaml
rexray:
  storageDrivers:
  - gce
gce:
  keyfile: /certdir/cert.json
```

### Configurable Items
The following items are configurable specific to this driver.
- [volumeTypes](https://cloud.google.com/compute/docs/reference/latest/diskTypes/list)

## Isilon
The Isilon driver registers a storage driver named `isilon` with the `REX-Ray`
driver manager and is used to connect and manage Isilon NAS storage.  The
driver creates logical volumes in directories on the Isilon cluster.  Volumes
are exported via NFS and restricted to a single client at a time.  Quotas can
also be used to ensure that a volume directory doesn't exceed a specified size.

### Configuration
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

### Extra Parameters
The following items are configurable specific to this driver.

 * `volumePath` represents the location under `/ifs/volumes` to allow volumes to
   be created and removed.
 * `nfsHost` is the configurable host used when mounting exports
 * `dataSubnet` is the subnet the REX-Ray driver is running on

### Optional Parameters
The following items are not required, but available to this driver.

 * `insecure` defaults to `false`.
 * `group` defaults to the group of the user specified in the configuration.  
   Only use this option if you need volumes to be created with a different
   group.
 * `volumePath` defaults to "".  This will have all new volumes created directly
   under `/ifs/volumes`.
 * `quotas` defaults to `false`.  Set to `true` if you have a SmartQuotas
   license enabled.

### Activating the Driver
To activate the Isilon driver please follow the instructions for
[activating storage drivers](/user-guide/config#storage-drivers),
using `isilon` as the driver name.

### Examples
Below is a full `config.yml` file that works with Isilon.

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

### Instructions
It is expected that the `volumePath` exists already within the Isilon system.
This example would reflect a directory create under `/ifs/volumes/rexray` for
created volumes.  It is not necessary to export this volume.  The `dataSubnet`
parameter is required so the Isilon driver can restrict access to attached
volumes to the host that REX-Ray is running on.

If `quotas` are enabled, a SmartQuotas license must also be enabled on the
Isilon cluster for the capacity size functionality of REX-Ray to work.

A SnapshotIQ license must be enabled on the Isilon cluster for the snapshot
functionality of `REX-Ray` to work.

### Caveats
The Isilon driver is not without its caveats:

 * The `--volumeType` flag is ignored
 * The account used to access the Isilon cluster must be in a role with the
  following privileges:
    * Namespace Access (ISI_PRIV_NS_IFS_ACCESS)
    * Platform API (ISI_PRIV_LOGIN_PAPI)
    * NFS (ISI_PRIV_NFS)
    * Restore (ISI_PRIV_IFS_RESTORE)
    * Quota (ISI_PRIV_QUOTA)          (if `quotas` are enabled)
    * Snapshot (ISI_PRIV_SNAPSHOT)    (if snapshots are used)

## OpenStack
The OpenStack driver registers a storage driver named `openstack` with the
`REX-Ray` driver manager and is used to connect and manage storage on OpenStack
instances.

### Configuration
The following is an example configuration of the OpenStack driver.

```yaml
openstack:
    authURL:              https://domain.com/openstack
    userID:               0
    userName:             admin
    password:             mypassword
    tenantID:             0
    tenantName:           customer
    domainID:             0
    domainName:           corp
    regionName:           USNW
    availabilityZoneName: Gold
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

### Activating the Driver
To activate the OpenStack driver please follow the instructions for
[activating storage drivers](/user-guide/config#storage-drivers),
using `openstack` as the driver name.

### Examples
Below is a full `config.yml` file that works with OpenStack.

```yaml
rexray:
  storageDrivers:
  - openstack
openstack:
  authUrl: https://keystoneHost:35357/v2.0/
  username: username
  password: password
  tenantName: tenantName
  regionName: regionName
```

## Rackspace
The Rackspace driver registers a storage driver named `rackspace` with the
`REX-Ray` driver manager and is used to connect and manage storage on Rackspace
instances.

### Configuration
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

### Activating the Driver
To activate the Rackspace driver please follow the instructions for
[activating storage drivers](/user-guide/config#storage-drivers),
using `rackspace` as the driver name.

### Examples
Below is a full `config.yml` file that works with Rackspace.

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

## ScaleIO
The ScaleIO driver registers a storage driver named `scaleio` with the `REX-Ray`
driver manager and is used to connect and manage ScaleIO storage.  The ScaleIO
`REST Gateway` is required for the driver to function.

### Configuration
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

#### Configuration Notes
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

### Runtime Behavior
The `storageType` field that is configured per volume is considered the
ScaleIO Storage Pool.  This can be configured by default with the `storagePool`
setting.  It is important that you create unique names for your Storage Pools
on the same ScaleIO platform.  Otherwise, when specifying `storageType` it
may choose at random which `protectionDomain` the pool comes from.

The `availabilityZone` field represents the ScaleIO Protection Domain.

### Configuring the Gateway
- Install the `EMC-ScaleIO-gateway` package.
- Edit the `/opt/emc/scaleio/gateway/webapps/ROOT/WEB-INF/classes/gatewayUser.properties`
file and append the proper MDM IP addresses to the following `mdm.ip.addresses=`
parameter.
- By default the password is the same as your administrative MDM password.
- Start the gateway `service scaleio-gateway start`.
 - With 1.32 we have noticed a restart of the gateway may be necessary as well
after an initial install with `service scaleio-gateway restart`. 

### Activating the Driver
To activate the ScaleIO driver please follow the instructions for
[activating storage drivers](/user-guide/config#storage-drivers),
using `scaleio` as the driver name.

### Troubleshooting
Ensure that you are able to open a TCP connection to the gateway with the
address that you will be supplying below in the `gateway_ip` parameter.  For
example `telnet gateway_ip 443` should open a successful connection.  Removing
the `EMC-ScaleIO-gateway` package and reinstalling can force re-creation of
self-signed certs which may help resolve gateway problems.  Also try restarting
the gateway with `service scaleio-gateway restart`.

### Examples
Below is a full `config.yml` file that works with ScaleIO.

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

## VirtualBox
The VirtualBox driver registers a storage driver named `virtualbox` with the
`REX-Ray` driver manager and is used by VirtualBox VM's to to connect and
manage volumes provided by Virtual Box.

### Prerequisites
In order to leverage the `virtualbox` driver, REX-Ray must be located on each
VM that you wish to be able to consume external volumes.  The driver
leverages the `vboxwebserv` HTTP SOAP API which is a process that must be
started from the VirtualBox *host* (ie OS X) using `vboxwebsrv -H 0.0.0.0 -v` or
additionally with `-b` for running in the background.  This allows the VMs
running `REX-Ray` to remotely make calls to the underlying VirtualBox
application.  A test for connectivity can be done with
`telnet virtualboxip 18083` from the VM.  The `virtualboxip` is what you
would put in the `endpoint` value.

It is optional to leverage authentication.  The HTTP SOAP API can have
authentication disabled by running
`VBoxManage setproperty websrvauthlibrary null`.

Hot-Plugging is required, which limits the usefulness of this driver to `SATA`
only.  Ensure that your VM has *pre-created* this controller and it is
named `SATA`.  Otherwise the `controllerName` field must be populated
with the name of the controller you wish to use.  The port count must be set
manually as it cannot be increased when the VMs are on.  A count of `30`
is sugggested.

VirtualBox 5.0.10+ must be used.

### Configuration
The following is an example configuration of the VirtualBox driver.  
The `localMachineNameOrId` parameter is for development use where you force
REX-Ray to use a specific VM identity.  Choose a `volumePath` to store the
volume files or virtual disks.  This path should be created ahead of time.


```yaml
virtualbox:
  endpoint: http://virtualboxhost:18083
  userName: optional
  password: optional
  tls: false
  volumePath: /Users/your_user/VirtualBox Volumes
  controllerName: name
  localMachineNameOrId: forDevelopmentUse
```

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config/#all-other-properties).

### Activating the Driver
To activate the VirtualBox driver please follow the instructions for
[activating storage drivers](/user-guide/config#storage-drivers),
using `virtualbox` as the driver name.

### Examples
Below is a working `config.yml` file that works with VirtualBox.

```yaml
rexray:
  storageDrivers:
  - virtualbox
virtualbox:
  endpoint: http://virtualBoxIP:18083
  volumePath: /Users/your_user/VirtualBox Volumes
```

### Caveats
- Snapshot and create volume from volume functionality is not
  available yet with this driver.
- The driver supports VirtualBox 5.0.10+

## VMAX
The VMAX driver registers a storage driver named `vmax` with the `REX-Ray`
driver manager and is used to connect and manage VMAX block storage.

This driver will in the future be used in many scenarios including HBA and
iSCSI.  Right now, the driver is functioning in the `vmh` mode.  This means
the Volume attachment and detachment is occurring as RDM's to a VM where
`REX-Ray` is running.  This use case enables you to address volumes for
containers as 1st class VMAX volumes while being capable of taking advantage
of mobility options such as `vSphere vMotion`.

### Configuration
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

### Extra Parameters
The following items are configurable specific to this driver.

 - `sid` is your array ID, ensure you enclose it in quotes to preserve it as a
   string.
 - `volumePrefix` is an (optional) field that limits visibility and usage of
   volumes to those only with this prefix, ex. `rexray_`.
 - `storageGroup` is the fields that determines which masking view to use
   when adding and removing volumes.

### Activating the Driver
To activate the VMAX driver please follow the instructions for
[activating storage drivers](/user-guide/config#storage-drivers),
using `vmax` as the driver name.

### Examples
Below is a full `config.yml` file that works with Isilon.

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

### Instructions
For the `vmh` mode, ensure that you have pre-created the storage group as
defined by `storageGroup`.  The underlying ESX hosts in the cluster where the VM
can move should already be in a masking view that has access to the VMAX ports
and should be properly logged in.

### Caveats
- This driver currently ignores the `--volumeType` flag.
- Pre-emption is not currently supported.  A mount to an alternate VM will be
  denied.
- A max of `56` devices is supported per VM.

## XtremIO
The XtremIO registers a storage driver named `xtremio` with the `REX-Ray`
driver manager and is used to connect and manage XtremIO storage.

### Configuration
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

### Activating the Driver
To activate the XtremIO driver please follow the instructions for
[activating storage drivers](/user-guide/config#storage-drivers),
using `xtremio` as the driver name.

### Examples
Below is a full `config.yml` file that works with XtremIO.

```yaml
rexray:
  storageDrivers:
  - xtremio
xtremio:
  endpoint: endpoint
  insecure: true
  username: admin
  password: password
  multipath: true
```

### Prerequisites
To use the XtremIO driver, an iSCSI connection between the host and XtremIO
must already be established.

#### Install
The driver currently is built for iSCSI operations with XtremIO.  It is expected
that connectivity between the host and XIO has been established.  The following
packages can be used for this.  `open-scsi` provides the iSCSI connectivity.
`multipath-tools` enables multi-path behavior and relates to the `multipath`
flag if installed.

- `apt-get install open-iscsi`
- `apt-get install multipath-tools`
- `iscsiadm -m discovery -t st -p 192.168.1.61`
- `iscsiadm -m node -l`

#### Initiator Group
Once a login has occurred, then you should be able to create a initiator
group for this iSCSI IQN.  You can leverage default naming for the initiator
and group.
