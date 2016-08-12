# Storage Providers

Connecting storage and platforms...

---

## Overview
This page reviews the storage providers and platforms supported by `libStorage`.

### Client/Server Configuration
Regarding the examples below, please
[read the provision](./config.md#clientserver-configuration) about
client/server configurations before proceeding.

## Isilon
The Isilon driver registers a storage driver named `isilon` with the
`libStorage` driver manager and is used to connect and manage Isilon NAS
storage. The driver creates logical volumes in directories on the Isilon
cluster. Volumes are exported via NFS and restricted to a single client at a
time. Quotas can also be used to ensure that a volume directory doesn't exceed
a specified size.

### Configuration
The following is an example configuration of the Isilon driver.

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
[transformed](./config.md#configuration-properties).

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
 * `volumePath` defaults to "". This will have all new volumes created directly
   under `/ifs/volumes`.
 * `quotas` defaults to `false`. Set to `true` if you have a SmartQuotas
   license enabled.

### Activating the Driver
To activate the Isilon driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers),
using `isilon` as the driver name.

### Examples
Below is a full `config.yml` file that works with Isilon.

```yaml
libstorage:
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

## ScaleIO
The ScaleIO driver registers a storage driver named `scaleio` with the
`libStorage` driver manager and is used to connect and manage ScaleIO storage.


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
example see the `Examples` section.

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
[transformed](./config.md#configuration-properties).

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
[activating storage drivers](./config.md#storage-drivers),
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

### Examples
Below is a full `config.yml` file that works with ScaleIO.

```yaml
libstorage:
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

## VirtualBox
The VirtualBox driver registers a storage driver named `virtualbox` with the
`libStorage` driver manager and is used by VirtualBox's VMs to connect and
manage volumes provided by VirtualBox.

### Prerequisites
In order to leverage the `virtualbox` driver, the `libStorage` client or must
be located on each VM that you wish to be able to consume external volumes.
The driver leverages the `vboxwebserv` HTTP SOAP API which is a process that
must be started from the VirtualBox *host* (ie OS X) using
`vboxwebsrv -H 0.0.0.0 -v` or additionally with `-b` for running in the
background. This allows the VMs running `libStorage` to remotely make calls to
the underlying VirtualBox application. A test for connectivity can be done with
`telnet virtualboxip 18083` from the VM. The `virtualboxip` is what you
would put in the `endpoint` value.

Leveraging authentication for the VirtualBox webserver is optiona.. The HTTP
SOAP API can have authentication disabled by running
`VBoxManage setproperty websrvauthlibrary null`.

Hot-Plugging is required, which limits the usefulness of this driver to `SATA`
only.  Ensure that your VM has *pre-created* this controller and it is
named `SATA`.  Otherwise the `controllerName` field must be populated
with the name of the controller you wish to use.  The port count must be set
manually as it cannot be increased when the VMs are on.  A count of `30`
is suggested.

VirtualBox 5.0.10+ must be used.

### Configuration
The following is an example configuration of the VirtualBox driver.  
The `localMachineNameOrId` parameter is for development use where you force
`libStorage` to use a specific VM identity.  Choose a `volumePath` to store the
volume files or virtual disks.  This path should be created ahead of time.


```yaml
virtualbox:
  endpoint: http://virtualboxhost:18083
  userName: optional
  password: optional
  tls: false
  volumePath: $HOME/VirtualBox/Volumes
  controllerName: name
  localMachineNameOrId: forDevelopmentUse
```
For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config.md#configuration-properties).

### Activating the Driver
To activate the VirtualBox driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers),
using `virtualbox` as the driver name.

### Examples
Below is a working `config.yml` file that works with VirtualBox.

```yaml
libstorage:
  server:
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

### Caveats
- Snapshot and create volume from volume functionality is not available yet
  with this driver.
- The driver supports VirtualBox 5.0.10+

## AWS EFS
The AWS EFS driver registers a storage driver named `efs` with the
`libStorage` driver manager and is used to connect and manage AWS Elastic File
Systems.

### Requirements

* AWS account
* VPC - EFS can be accessed within VPC
* AWS Credentials

### Configuration
The following is an example with all possible fields configured.  For a running
example see the `Examples` section.

```yaml
efs:
  accessKey:      XXXXXXXXXX
  secretKey:      XXXXXXXXXX
  securityGroups: sg-XXXXXXX,sg-XXXXXX0,sg-XXXXXX1
  region:         us-east-1
  tag:            test
```

#### Configuration Notes
- The `accessKey` and `secretKey` configuration parameters are optional and should
be used when explicit AWS credentials configuration needs to be provided. EFS driver
uses official golang AWS SDK library and supports all other ways of providing
access credentials, like environment variables or instance profile IAM permissions.
- `region` represents AWS region where should be EFS provisioned. See official AWS
documentation for list of supported regions.
- `securityGroups` list of security groups attached to `MountPoint` instances.
If no security groups are provided the default VPC security group is used.
- `tag` is used to partition multiple services within single AWS account and is
used as prefix for EFS names in format `[tagprefix]/volumeName`.

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config.md#configuration-properties).

### Runtime Behavior

AWS EFS storage driver creates one EFS FileSystem per volume and provides root
of the filesystem as NFS mount point. Volumes aren't attached to instances
directly but rather exposed to each subnet by creating `MountPoint` in each VPC
subnet. When detaching volume from instance no action is taken as there isn't
good way to figure out if there are other instances in same subnet using
`MountPoint` that is being detached. There is no charge for `MountPoint`
so they are removed only once whole volume is deleted.

By default all EFS instances are provisioned as `generalPurpose` performance mode.
`maxIO` EFS type can be provisioned by providing `maxIO` flag as `volumetype`.

Its possible to mount same volume to multiple container on a single EC2 instance
as well as use single volume across multiple EC2 instances at the same time.

**NOTE**: Each EFS FileSystem can be accessed only from single VPC at the time.

### Activating the Driver
To activate the AWS EFS driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers),
using `efs` as the driver name.

### Troubleshooting
- Make sure that AWS credentials (user or role) has following AWS permissions on
  `libStorage` server instance that will be making calls to AWS API:
  - `elasticfilesystem:CreateFileSystem`
  - `elasticfilesystem:CreateMountTarget`
  - `ec2:DescribeSubnets`
  - `ec2:DescribeNetworkInterfaces`
  - `ec2:CreateNetworkInterface`
  - `elasticfilesystem:CreateTags`
  - `elasticfilesystem:DeleteFileSystem`
  - `elasticfilesystem:DeleteMountTarget`
  - `ec2:DeleteNetworkInterface`
  - `elasticfilesystem:DescribeFileSystems`
  - `elasticfilesystem:DescribeMountTargets`

### Examples
Below is a working `config.yml` file that works with AWS EFS.

```yaml
libstorage:
  server:
    services:
      efs:
        driver: efs
        efs:
          accessKey:      XXXXXXXXXX
          secretKey:      XXXXXXXXXX
          securityGroups: sg-XXXXXXX,sg-XXXXXX0,sg-XXXXXX1
          region:         us-east-1
          tag:            test
```