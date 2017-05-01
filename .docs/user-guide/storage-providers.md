# Storage Providers

Connecting storage and platforms...

---

## Overview
This page reviews the storage providers and platforms supported by `libStorage`.

### Client/Server Configuration
Regarding the examples below, please
[read the provision](./config.md#clientserver-configuration) about
client/server configurations before proceeding.

## Amazon
libStorage includes support for multiple Amazon Web Services (AWS) storage
services.

<a class="headerlink hiddenanchor" name="aws-ebs"></a>

### Elastic Block Storage
The AWS EBS driver registers a storage driver named `ebs` with the
libStorage service registry and is used to connect and manage AWS Elastic Block
Storage volumes for EC2 instances.

!!! note
    For backwards compatibility, the driver also registers a storage driver
    named `ec2`. The use of `ec2` in config files is deprecated but functional.

!!! note
    The EBS driver does not yet support snapshots or tags, as previously
    supported in Rex-Ray v0.3.3.

The EBS driver is made possible by the
[official Amazon Go AWS SDK](https://github.com/aws/aws-sdk-go.git).

#### Requirements

* AWS account
* VPC - EBS can be accessed within VPC
* AWS Credentials

#### Configuration
The following is an example with all possible fields configured.  For a running
example see the [Examples](./storage-providers.md#aws-ebs-examples) section.

```yaml
ebs:
  accessKey:      XXXXXXXXXX
  secretKey:      XXXXXXXXXX
  region:         us-east-1
  maxRetries:     10
  kmsKeyID:       arn:aws:kms:us-east-1:012345678910:key/abcd1234-a123-456a-a12b-a123b4cd56ef
  statusMaxAttempts:  10
  statusInitialDelay: 100ms
  statusTimeout:      2m
```

##### Configuration Notes
- The `accessKey` and `secretKey` configuration parameters are optional and
should be used when explicit AWS credentials configuration needs to be provided.
EBS driver uses official golang AWS SDK library and supports all other ways of
providing access credentials, like environment variables or instance profile IAM
 permissions.
- `region` represents AWS region where EBS volumes should be provisioned.
See official AWS documentation for list of supported regions.
<!-- - `tag` is used to partition multiple services within single AWS account
and is used as prefix for EBS names in format `[tagprefix]/volumeName`. -->
- `maxRetries` is the number of retries that will be made for failed operations
  by the AWS SDK.
- If the `kmsKeyID` field is specified it will be used as the encryption key for
all volumes that are created with a truthy encryption request field.
- `statusMaxAttempts` is the number of times the status of a volume will be
  queried before giving up when waiting on a status change
- `statusInitialDelay` specifies a time duration used to wait when polling
  volume status. This duration is used in exponential backoff, such that the
  first wait will be for this duration, the second for 2x, the third for 4x,
  etc. The units of the duration must be given (e.g. "100ms" or "1s").
- `statusTimeout` is a maximum length of time that polling for volume status can
  occur. This serves as a backstop against a stuck request of malfunctioning API
  that never returns.


For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config.md#configuration-properties).

<!--### Volume tagging (optional)
By default, EBS driver has access to all volumes and snapshots defined in your
AWS account. Volume tagging gives you the ability to only include management of
volumes that have a specific `ec2 tag`. There is an optional `tag` key in
`ebs` configuration section to limit the access to and tagging that happens
for any new volumes or snapshots created. The objects will have a `ec2 tag`
called `libstroageSet` with a value defined by the configurable `tag`.

For example, if you had a set of hosts you can configure `libstorage` to tag
them with `prod`, `testing` or `development` each with its own set of volumes
and snapshots.

Volumes and snapshots that are accessed directly from `volumeID` can still be
controlled regardless of the `tag`. -->

#### Activating the Driver
To activate the AWS EBS driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers),
using `ebs` as the driver name.

#### Troubleshooting
- Make sure that AWS credentials (user or role) has following AWS permissions on
  `libStorage` server instance that will be making calls to AWS API:
    - `ec2:AttachVolume`,
    - `ec2:CreateVolume`,
    - `ec2:CreateSnapshot`,
    - `ec2:CreateTags`,
    - `ec2:DeleteVolume`,
    - `ec2:DeleteSnapshot`,
    - `ec2:DescribeAvailabilityZones`,
    - `ec2:DescribeInstances`,
    - `ec2:DescribeVolumes`,
    - `ec2:DescribeVolumeAttribute`,
    - `ec2:DescribeVolumeStatus`,
    - `ec2:DescribeSnapshots`,
    - `ec2:CopySnapshot`,
    - `ec2:DescribeSnapshotAttribute`,
    - `ec2:DetachVolume`,
    - `ec2:ModifySnapshotAttribute`,
    - `ec2:ModifyVolumeAttribute`,
    - `ec2:DescribeTags`

<a class="headerlink hiddenanchor" name="aws-ebs-examples"></a>

#### Examples
Below is a working `config.yml` file that works with AWS EBS.

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: ebs
  server:
    services:
      ebs:
        driver: ebs
        ebs:
          accessKey:      XXXXXXXXXX
          secretKey:      XXXXXXXXXX
          region:         us-east-1
```

<a class="headerlink hiddenanchor" name="aws-efs"></a>

### Elastic File System
The AWS EFS driver registers a storage driver named `efs` with the
libStorage service registry and is used to connect and manage AWS Elastic File
Systems.

#### Requirements

* AWS account
* VPC - EFS can be accessed within VPC
* AWS Credentials

#### Configuration
The following is an example with all possible fields configured.  For a running
example see the [Examples](./storage-providers.md#aws-efs-examples) section.

```yaml
efs:
  accessKey:      XXXXXXXXXX
  secretKey:      XXXXXXXXXX
  securityGroups:
  - sg-XXXXXXX
  - sg-XXXXXX0
  - sg-XXXXXX1
  region:              us-east-1
  tag:                 test
  disableSessionCache: false
  statusMaxAttempts:  6
  statusInitialDelay: 1s
  statusTimeout:      2m
```

##### Configuration Notes
- The `accessKey` and `secretKey` configuration parameters are optional and
should be used when explicit AWS credentials configuration needs to be provided.
EFS driver uses official golang AWS SDK library and supports all other ways of
providing access credentials, like environment variables or instance profile IAM
 permissions.
- `region` represents AWS region where EFS should be provisioned. See official
AWS documentation for list of supported regions.
- `securityGroups` list of security groups attached to `MountPoint` instances.
If no security groups are provided the default VPC security group is used.
- `tag` is used to partition multiple services within single AWS account and is
used as prefix for EFS names in format `[tagprefix]/volumeName`.
- `disableSessionCache` is a flag that can be used to disable the session cache.
If the session cache is disabled then a new AWS connection is established with
every API call.
- `statusMaxAttempts` is the number of times the status of a volume will be
  queried before giving up when waiting on a status change
- `statusInitialDelay` specifies a time duration used to wait when polling
  volume status. This duration is used in exponential backoff, such that the
  first wait will be for this duration, the second for 2x, the third for 4x,
  etc. The units of the duration must be given (e.g. "100ms" or "1s").
- `statusTimeout` is a maximum length of time that polling for volume status can
  occur. This serves as a backstop against a stuck request of malfunctioning API
  that never returns.


For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config.md#configuration-properties).

#### Runtime Behavior

AWS EFS storage driver creates one EFS FileSystem per volume and provides root
of the filesystem as NFS mount point. Volumes aren't attached to instances
directly but rather exposed to each subnet by creating `MountPoint` in each VPC
subnet. When detaching volume from instance no action is taken as there isn't
good way to figure out if there are other instances in same subnet using
`MountPoint` that is being detached. There is no charge for `MountPoint`
so they are removed only once whole volume is deleted.

By default all EFS instances are provisioned as `generalPurpose` performance
mode. `maxIO` EFS type can be provisioned by providing `maxIO` flag as
`volumetype`.

Its possible to mount same volume to multiple container on a single EC2 instance
as well as use single volume across multiple EC2 instances at the same time.

!!! note
    Each EFS FileSystem can be accessed only from single VPC at the time.

#### Activating the Driver
To activate the AWS EFS driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers),
using `efs` as the driver name.

#### Troubleshooting
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

<a class="headerlink hiddenanchor" name="aws-efs-examples"></a>

#### Examples
Below is a working `config.yml` file that works with AWS EFS.

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: efs
  server:
    services:
      efs:
        driver: efs
        efs:
          accessKey:      XXXXXXXXXX
          secretKey:      XXXXXXXXXX
          securityGroups:
          - sg-XXXXXXX
          - sg-XXXXXX0
          - sg-XXXXXX1
          region:         us-east-1
          tag:            test
```

<a class="headerlink hiddenanchor" name="aws-s3fs"></a>

### Simple Storage Service
The AWS S3FS driver registers a storage driver named `s3fs` with the
libStorage service registry and provides the ability to mount Amazon Simple
Storage Service (S3) buckets as filesystems using the
[`s3fs`](https://github.com/s3fs-fuse/s3fs-fuse) FUSE command.

Unlike the other AWS-related drivers, the S3FS storage driver does not need
to deployed or used by an EC2 instance. Any client can take advantage of
Amazon's S3 buckets.

#### Requirements
* AWS account
* The [`s3fs`](https://github.com/s3fs-fuse/s3fs-fuse) FUSE command must be
present on client nodes.

#### Configuration
The following is an example with all possible fields configured.  For a running
example see the [Examples](./storage-providers.md#aws-s3fs-examples) section.

#### Server-Side Configuration
```yaml
s3fs:
  region:           us-east-1
  accessKey:        XXXXXXXXXX
  secretKey:        XXXXXXXXXX
  disablePathStyle: false
```

* The `accessKey` and `secretKey` configuration parameters are optional and
should be used when explicit AWS credentials configuration needs to be provided.
The S3FS driver uses the official Golang AWS SDK library and supports all other
ways of  providing access credentials, like environment variables or instance
profile IAM permissions.
* `region` represents AWS region where S3FS buckets should be provisioned.
Please see the official AWS documentation for list of
[supported regions](http://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region).
* The `disablePathStyle` property disables the use of the path style for
bucket endpoints. The path style is more stable with regards to regions
than bucket URI FQDNs, but the path style is also less performant.

#### Client-Side Configuration
```yaml
s3fs:
  cmd:            s3fs
  options:
  - XXXX
  - XXXX
  accessKey:      XXXXXXXXXX
  secretKey:      XXXXXXXXXX
```

* The `cmd` property defaults simply to `s3fs` with the assumption that the
`s3fs` binary will be in the path. This value can also be the absolute path
to the `s3fs` binary.
* `options` is a list of options to pass to the `s3fs` command. Please see the
[official](https://github.com/s3fs-fuse/s3fs-fuse/wiki/Fuse-Over-Amazon)
documentation for a full list of CLI options. The `-o` prefix should not be
provided in the configuration file.
* The credential properties can be defined on the client via the configuration
file and will be supplied to the `s3fs` process via environment variables.
However, the `s3fs` command will also look in all the
[usual places](https://github.com/s3fs-fuse/s3fs-fuse/wiki/Fuse-Over-Amazon)
for the credentials if they're not in this file.

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./config.md#configuration-properties).

#### Runtime Behavior
The AWS S3FS storage driver can create new buckets as well as remove existing
ones. Buckets are mounted to clients as filesystems using the
[`s3fs`](https://github.com/s3fs-fuse/s3fs-fuse) FUSE command. For clients
to correctly mount and unmount S3 buckets the `s3fs` command should be in
the path of the executor or configured via the `s3fs.cmd` property in the
client-side REX-Ray configuration file.

The client must also have access to the AWS credentials used for mounting and
unmounting S3 buckets. These credentials can be stored in the client-side
REX-Ray configuration file or via
[any means avaialble](https://github.com/s3fs-fuse/s3fs-fuse/wiki/Fuse-Over-Amazon)
to the `s3fs` command.


#### Activating the Driver
To activate the AWS S3FS driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers),
using `s3fs` as the driver name.

<a class="headerlink hiddenanchor" name="aws-s3fs-examples"></a>

#### Examples
Below is a working `config.yml` file that works with AWS S3FS.

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: s3fs
  server:
    services:
      s3fs:
        driver: s3fs
        s3fs:
          accessKey:      XXXXXXXXXX
          secretKey:      XXXXXXXXXX
```

## Ceph
libStorage includes support for the following Ceph storage technologies.

<a class="headerlink hiddenanchor" name="ceph-rbd"></a>

### RADOS Block Device
The Ceph RBD driver registers a driver named `rbd` with the `libStorage` driver
manager and is used to connect and mount RADOS Block Devices from a Ceph
cluster.

#### Requirements

* The `ceph` and `rbd` binary executables must be installed on the host
* The `rbd` kernel module must be installed
* A `ceph.conf` file must be present in its default location
  (`/etc/ceph/ceph.conf`)
* The ceph `admin` key must be present in `/etc/ceph/`

#### Configuration
The following is an example with all possible fields configured. For a running
example see the [Examples](./storage-providers.md#ceph-rbd-examples) section.

```yaml
rbd:
  defaultPool: rbd
```

##### Configuration Notes

* The `defaultPool` parameter is optional, and defaults to "rbd". When set, all
  volume requests that do not reference a specific pool will use the
  `defaultPool` value as the destination storage pool.

#### Runtime behavior

The Ceph RBD driver only works when the client and server are on the same node.
There is no way for a centralized `libStorage` server to attach volumes to
clients, therefore the `libStorage` server must be running on each node that
wishes to mount RBD volumes.

The RBD driver uses the format of `<pool>.<name>` for the volume ID. This allows
for the use of multiple pools by the driver. During a volume create, if the
volume ID is given as `<pool>.<name>`, a volume named *name* will be created in
the *pool* storage pool. If no pool is referenced, the `defaultPool` will be
used.

Both *pool* and *name* may only contain alphanumeric characters, underscores,
and dashes.

When querying volumes, the driver will return all RBDs present in all pools in
the cluster, prefixing each volume with the appropriate `<pool>.` value.

All RBD creates are done using the default 4MB object size, and using the
"layering" feature bit to ensure greatest compatibility with the kernel clients.

#### Activating the Driver
To activate the Ceph RBD driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers), using `rbd` as the
driver name.

#### Troubleshooting

* Make sure that `ceph` and `rbd` commands work without extra parameters for
  ID, key, and monitors. All configuration must come from `ceph.conf`.
* Check status of the ceph cluster with `ceph -s` command.

<a class="headerlink hiddenanchor" name="ceph-rbd-examples"></a>

#### Examples

Below is a full `config.yml` that works with RBD

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: rbd
  server:
    services:
      rbd:
        driver: rbd
        rbd:
          defaultPool: rbd
```

#### Caveats
* Snapshot and copy functionality is not yet implemented
* libStorage Server must be running on each host to mount/attach RBD volumes
* There is not yet options for using non-admin cephx keys or changing RBD create
  features
* Volume pre-emption is not supported. Ceph does not provide a method to
  forcefully detach a volume from a remote host -- only a host can attach and
  detach volumes from itself.
* RBD advisory locks are not yet in use. A volume is returned as "unavailable"
  if it has a watcher other than the requesting client. Until advisory locks are
  in place, it may be possible for a client to attach a volume that is already
  attached to another node. Mounting and writing to such a volume could lead to
  data corruption.

## Dell EMC
libStorage includes support for several Dell EMC storage platforms.

<a class="headerlink hiddenanchor" name="dell-emc-isilon"></a>

### Isilon
The Isilon driver registers a storage driver named `isilon` with the
libStorage service registry and is used to connect and manage Isilon NAS
storage. The driver creates logical volumes in directories on the Isilon
cluster. Volumes are exported via NFS and restricted to a single client at a
time. Quotas can also be used to ensure that a volume directory doesn't exceed
a specified size.

#### Configuration
The following is an example configuration of the Isilon driver. For a running
example see the [Examples](./storage-providers.md#dell-emc-isilon-examples)
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
[transformed](./config.md#configuration-properties).

#### Extra Parameters
The following items are configurable specific to this driver.

 * `volumePath` represents the location under `/ifs/volumes` to allow volumes to
   be created and removed.
 * `nfsHost` is the configurable NFS server hostname or IP (often a
   SmartConnect name) used when mounting exports
 * `dataSubnet` is the subnet the REX-Ray driver is running on. This is used
   for the NFS export host ACLs.

#### Optional Parameters
The following items are not required, but available to this driver.

 * `insecure` defaults to `false`.
 * `group` defaults to the group of the user specified in the configuration.
   Only use this option if you need volumes to be created with a different
   group.
 * `volumePath` defaults to "". This will have all new volumes created directly
   under `/ifs/volumes`.
 * `quotas` defaults to `false`. Set to `true` if you have a SmartQuotas
   license enabled.

#### Activating the Driver
To activate the Isilon driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers),
using `isilon` as the driver name.

<a class="headerlink hiddenanchor" name="dell-emc-isilon-examples"></a>

#### Examples
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

#### Instructions
It is expected that the `volumePath` exists already within the Isilon system.
This example would reflect a directory create under `/ifs/volumes/libstorage`
for created volumes. It is not necessary to export this volume. The `dataSubnet`
parameter is required so the Isilon driver can restrict access to attached
volumes to the host that REX-Ray is running on.

If `quotas` are enabled, a SmartQuotas license must also be enabled on the
Isilon cluster for the capacity size functionality of `libStorage` to work.

A SnapshotIQ license must be enabled on the Isilon cluster for the snapshot
functionality of `libStorage` to work.

#### Caveats
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

### ScaleIO
The ScaleIO driver registers a storage driver named `scaleio` with the
libStorage service registry and is used to connect and manage ScaleIO storage.


#### Requirements
 - The ScaleIO `REST Gateway` is required for the driver to function.
 - The `libStorage` client or application that embeds the `libStorage` client
   must reside on a host that has the SDC client installed. The command
   `/opt/emc/scaleio/sdc/bin/drv_cfg --query_guid` should be executable and
   should return the local SDC GUID.
 - The [official](http://www.oracle.com/technetwork/java/javase/downloads/index.html)
   Oracle Java Runtime Environment (JRE) is required. During testing, use of the
   Open Java Development Kit (JDK) resulted in unexpected errors.

#### Configuration
The following is an example with all possible fields configured.  For a running
example see the [Examples](./storage-providers.md#dell-emc-scaleio-examples)
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

##### Configuration Notes
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

#### Runtime Behavior
The `storageType` field that is configured per volume is considered the
ScaleIO Storage Pool.  This can be configured by default with the `storagePool`
setting.  It is important that you create unique names for your Storage Pools
on the same ScaleIO platform.  Otherwise, when specifying `storageType` it
may choose at random which `protectionDomain` the pool comes from.

The `availabilityZone` field represents the ScaleIO Protection Domain.

#### Configuring the Gateway
- Install the `EMC-ScaleIO-gateway` package.
- Edit the
`/opt/emc/scaleio/gateway/webapps/ROOT/WEB-INF/classes/gatewayUser.properties`
file and append the proper MDM IP addresses to the following `mdm.ip.addresses=`
parameter.
- By default the password is the same as your administrative MDM password.
- Start the gateway `service scaleio-gateway start`.
 - With 1.32 we have noticed a restart of the gateway may be necessary as well
after an initial install with `service scaleio-gateway restart`.

#### Activating the Driver
To activate the ScaleIO driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers),
using `scaleio` as the driver name.

#### Troubleshooting
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

<a class="headerlink hiddenanchor" name="dell-emc-scaleio-examples"></a>

#### Examples
Below is a full `config.yml` file that works with ScaleIO.

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: scaelio
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

## DigitalOcean
Thanks to the efforts of our tremendous community, libStorage also has built-in
support for DigitalOcean!

<a class="headerlink hiddenanchor" name="digitalocean-block-storage"></a>
<a class="headerlink hiddenanchor" name="dobs"></a>

### DO Block Storage
The DigitalOcean Block Storage (DOBS) driver registers a driver named `dobs`
with the libStorage service registry and is used to attach and mount
DigitalOcean block storage devices to DigitalOcean instances.

#### Requirements
The DigitalOcean block storage driver has the following requirements:

* Valid DigitalOcean account
* Valid DigitalOcean [access token](https://goo.gl/iKoAec)

#### Configuration
The following is an example with all possible fields configured. For a running
example see the [Examples](./storage-providers.md#dobs-examples) section.

```yaml
dobs:
  token:  123456
  region: nyc1
  statusMaxAttempts: 10
  statusInitialDelay: 100ms
  statusTimeout: 2m
```

##### Configuration notes
- The `token` contains your DigitalOcean [access token](https://goo.gl/iKoAec)
- `region` specifies the DigitalOcean region where volumes should be created
- `statusMaxAttempts` is the number of times the status of a volume will be
  queried before giving up when waiting on a status change
- `statusInitialDelay` specifies a time duration used to wait when polling
  volume status. This duration is used in exponential backoff, such that the
  first wait will be for this duration, the second for 2x, the third for 4x,
  etc. The units of the duration must be given (e.g. "100ms" or "1s").
- `statusTimeout` is a maximum length of time that polling for volume status can
  occur. This serves as a backstop against a stuck request of malfunctioning API
  that never returns.

!!! note
    The DigitalOcean service currently only supports block storage volumes in
    specific regions. Make sure to use a [suuported region](https://www.digitalocean.com/community/tutorials/how-to-use-block-storage-on-digitalocean#what-is-digitalocean-block-storage).

    The standard environment variable for the DigitalOcean access token is
    `DIGITALOCEAN_ACCESS_TOKEN`. However, the environment variable mapped to
    this driver's `dobs.token` property is `DOBS_TOKEN`. This choice was made
    to ensure that the driver must be explicitly configured for access instead
    of detecting a default token that may not be intended for the driver.

<a class="headerlink hiddenanchor" name="dobs-examples"></a>

#### Examples
Below is a full `config.yml` that works with DOBS

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: dobs
  server:
    services:
      dobs:
        driver: dobs
        dobs:
          token: 123456
          region: nyc1
```

## FittedCloud
Another example of the great community shared by the libStorage project, the
talented people at FittedCloud have provided a driver for their EBS optimizer.

<a class="headerlink hiddenanchor" name="fittedcloud-ebs"></a>
<a class="headerlink hiddenanchor" name="fittedcloud-ebs-optimizer"></a>

### EBS Optimizer
The FittedCloud EBS Optimizer driver registers a storage driver named
`fittedcloud` with the libStorage service registry and provides the ability to
connect and manage thin-provisioned EBS volumes for EC2 instances.

!!! note
    This version of the FittedCloud driver only supports configurations where
    client and server are on the same host.  The libStorage server must be
    running on each node along side with the FittedCloud Agent.

!!! note
    This version of the FittedCloud driver does not support co-existing with the
    ebs driver on the same host. As a result it also doesn't support optimizing
    existing EBS volumes. See the [Examples](#fittedcloud-examples) section
    below for a running example.

!!! note
    The FittedCloud driver does not yet support snapshots or tags.

#### Requirements
This driver has the following requirements:

* AWS account
* VPC - EBS can be accessed within VPC
* AWS Credentials
* FittedCloud Agent software

<a class="headerlink hiddenanchor" name="fittedcloud-getting-started"></a>

#### Getting Started
Before starting, please make sure to register as a user by visiting the
FittedCloud [customer website](https://customer.fittedcloud.com/register).
Once an account is activated it will be assigned a user ID, which can be found
on the Settings page after logging into the web site.

The following commands will download and install the latest FittedCloud Agent
software. The flags `-o S -m` enable new thin volumes to be created via the
docker command instead of optimizing existing EBS volumes.
Please replace the `<User ID>` with a FittedCloud user ID.

```sh
$ curl -skSL 'https://customer.fittedcloud.com/downloadsoftware?ver=latest' \
  -o fcagent.run
$ sudo bash ./fcagent.run -- -o S -m -d <User ID>
```

Please refer to FittedCloud
[website](https://customer.fittedcloud.com/download) for more details.

<a class="headerlink hiddenanchor" name="fittedcloud-config"></a>

#### Configuration
The following is an example with all possible fields configured.  For a running
example see the [Examples](./storage-providers.md#fittedcloud-examples) section.

```yaml
ebs:
  accessKey:      XXXXXXXXXX
  secretKey:      XXXXXXXXXX
  kmsKeyID:       abcd1234-a123-456a-a12b-a123b4cd56ef
  statusMaxAttempts:  10
  statusInitialDelay: 100ms
  statusTimeout:      2m
```

##### Configuration Notes
- FittedCloud driver shares the ebs driver's configuration
parameters.
- The `accessKey` and `secretKey` configuration parameters are optional and
should be used when explicit AWS credentials configuration needs to be provided.
FittedCloud driver uses official golang AWS SDK library and supports all other
ways of providing access credentials, like environment variables or instance
profile IAM permissions.
- If the `kmsKeyID` field is specified it will be used as the encryption key for
all volumes that are created with a truthy encryption request field.
- `statusMaxAttempts` is the number of times the status of a volume will be
  queried before giving up when waiting on a status change
- `statusInitialDelay` specifies a time duration used to wait when polling
  volume status. This duration is used in exponential backoff, such that the
  first wait will be for this duration, the second for 2x, the third for 4x,
  etc. The units of the duration must be given (e.g. "100ms" or "1s").
- `statusTimeout` is a maximum length of time that polling for volume status can
  occur. This serves as a backstop against a stuck request of malfunctioning API
  that never returns.

<a class="headerlink hiddenanchor" name="fittedcloud-examples"></a>

#### Examples
The following example illustrates how to configured the FittedCloud driver:

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: fittedcloud
  server:
    services:
      fittedcloud:
        driver: fittedcloud
ebs:
  accessKey:  XXXXXXXXXX
  secretKey:  XXXXXXXXXX
```

Additional information on configuring the FittedCloud driver may be found
at [this](https://goo.gl/I6mf20) location.

## Google
libStorage ships with support for Google Compute Engine (GCE) as well.

<a class="headerlink hiddenanchor" name="gce-persistent-disk"></a>

### GCE Persistent Disk
The Google Compute Engine Persistent Disk (GCEPD) driver registers a driver
named `gcepd` with the libStorage service registry and is used to connect and
mount Google Compute Engine (GCE) persistent disks with GCE machine instances.

#### Requirements
* GCE account
* The libStorage server must be running on a GCE instance created with a Service
  Account with appropriate permissions, or a Service Account credentials file
  in JSON format must be supplied. If not using the Compute Engine default
  Service Account with the Cloud Platform/"all cloud APIs" scope, create a new
  Service Account via the [IAM Portal](https://console.cloud.google.com/iam-admin/serviceaccounts).
  This Service Account requires the `Compute Engine/Instance Admin`,
  `Compute Engine/Storage Admin`, and `Project/Service Account Actor` roles.
  Then create/download a new private key in JSON format. see
  [creating a service account](https://developers.google.com/identity/protocols/OAuth2ServiceAccount#creatinganaccount)
  for details. The libStorage service must be restarted in order for permissions
  changes on a service account to take effect.

#### Configuration
The following is an example with all possible fields configured. For a running
example see the [Examples](./storage-providers.md#gce-persistent-disk-examples)
section.

```yaml
gcepd:
  keyfile: /etc/gcekey.json
  zone: us-west1-b
  defaultDiskType: pd-ssd
  tag: rexray
  statusMaxAttempts:  10
  statusInitialDelay: 100ms
  statusTimeout:      2m
```

##### Configuration Notes
* The `keyfile` parameter is optional. It specifies a path on disk to a file
  containing the JSON-encoded Service Account credentials. This file can be
  downloaded from the GCE web portal. If `keyfile` is specified, the GCE
  instance's service account is not considered, and is not necessary. If
  `keyfile` is *not* specified, the application will try to lookup
  [application default credentials](https://developers.google.com/identity/protocols/application-default-credentials).
  This has the effect of looking for credentials in the priority described
  [here](https://godoc.org/golang.org/x/oauth2/google#FindDefaultCredentials).
* The `zone` parameter is optional, and configures the driver to *only* allow
  access to the given zone. Creating and listing disks from other zones will be
  denied. If a zone is not specified, the zone from the client Instance ID will
  be used when creating new disks.
* The `defaultDiskType` parameter is optional and specifies what type of disk
  to create, either `pd-standard` or `pd-ssd`. When not specified, the default
  is `pd-ssd`.
* The `tag` parameter is optional, and causes the driver to create or return
  disks that have a matching tag. The tag is implemented by using the GCE
  label functionality available in the beta API. The value of the `tag`
  parameter is used as the value for a label with the key `libstoragetag`.
  Use of this parameter is encouraged, as the driver will only return volumes
  that have been created by the driver, which is most useful to eliminate
  listing the boot disks of every GCE disk in your project/zone. If you wsih to
  "expose" previously created disks to the `GCEPD` driver, you can edit the
  labels on the existing disk to have a key of `libstoragetag` and a value
  matching that given in `tag`.
* `statusMaxAttempts` is the number of times the status of a volume will be
  queried before giving up when waiting on a status change
* `statusInitialDelay` specifies a time duration used to wait when polling
  volume status. This duration is used in exponential backoff, such that the
  first wait will be for this duration, the second for 2x, the third for 4x,
  etc. The units of the duration must be given (e.g. "100ms" or "1s").
* `statusTimeout` is a maximum length of time that polling for volume status can
  occur. This serves as a backstop against a stuck request of malfunctioning API
  that never returns.

#### Runtime behavior
* The GCEPD driver enforces the GCE requirements for disk sizing and naming.
  Disks must be created with a minimum size of 10GB. Disk names must adhere to
  the regular expression of `[a-z]([-a-z0-9]*[a-z0-9])?`, which means the first
  character must be a lowercase letter, and all following characters must be a
  dash, lowercase letter, or digit, except the last character, which cannot be a
  dash.
* If the `zone` parameter is not specified in the driver configuration, and a
  request is received to list all volumes that does not specify a zone in the
  InstanceID header, volumes from all zones will be returned.
* By default, all disks will be created with type `pd-ssd`, which creates an SSD
  based disk. If you wish to create disks that are not SSD-based, change the
  default via the driver config, or the type can be changed at creation time by
  using the `Type` field of the create request.

#### Activating the Driver
To activate the GCEPD driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers), using `gcepd` as the
driver name.

#### Troubleshooting
* Make sure that the JSON credentials file as specified in the `keyfile`
  configuration parameter is present and accessible, or that you are running in
  a GCE instance created with a Service Account attached. Whether using a
  `keyfile` or the Service Account associated with the GCE instance, the Service
  Account must have the appropriate permissions as described in
  `Configuration Notes`

<a class="headerlink hiddenanchor" name="gce-persistent-disk-examples"></a>

#### Examples
Below is a full `config.yml` that works with GCE

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: gcepd
  server:
    services:
      gcepd:
        driver: gcepd
        gcepd:
          keyfile: /etc/gcekey.json
          tag: rexray
```

#### Caveats
* Snapshot and copy functionality is not yet implemented
* Most GCE instances can have up to 64 TB of total persistent disk space
  attached. Shared-core machine types or custom machine types with less than
  3.75 GB of memory are limited to 3 TB of total persistent disk space. Total
  persistent disk space for an instance includes the size of the root persistent
  disk. You can attach up to 16 independent persistent disks to most instances,
  but instances with shared-core machine types or custom machine types with less
  than 3.75 GB of memory are limited to a maximum of 4 persistent disks,
  including the root persistent disk. See
  [GCE Disks](https://cloud.google.com/compute/docs/disks/) docs for more
  details.
* If running libStorage server in a mode where volume mounts will not be
  performed on the same host where libStorage server is running, it should be
  possible to use a Service Account without the `Service Account Actor` role,
  but this has not been tested. Note that if persistent disk mounts are to be
  performed on *any* GCE instances that have a Service Account associated with
  the, the `Service Account Actor` role is required.

## Microsoft
Microsoft Azure support is included with libStorage as well.

<a class="headerlink hiddenanchor" name="azure-ud"></a>

### Azure Unmanaged Disk
The Microsoft Azure Unmanaged Disk (Azure UD) driver registers a driver
named `azureud` with the libStorage service registry and is used to connect and
mount Azure unmanaged disks from Azure page blob storage with Azure virtual
machines.

#### Requirements
* An Azure account
* An Azure subscription
* An Azure storage account
* An Azure resource group
* Any virtual machine where disks are going to be attached must have the
  `lsscsi` utility installed. You can install this with `yum install lsscsi` on
  Red Hat based distributions, or with `apt-get install lsscsi` on Debian based
  distributions.

#### Configuration
The following is an example with all possible fields configured. For a running
example see the [Examples](./storage-providers.md#azure-ud-examples) section.

```yaml
azureud:
  subscriptionID: abcdef01-2345-6789-abcd-ef0123456789
  resourceGroup: testgroup
  tenantID: usernamehotmail.onmicrosoft.com
  storageAccount: username
  storageAccessKey: XXXXXXXX
  clientID: 123def01-2345-6789-abcd-ef0123456789
  clientSecret: XXXXXXXX
  certPath:
  container: vhds
  useHTTPS: true
```

##### Configuration Notes
* `subscriptionID` is required, and is the UUID of your Azure subscription
* `resourceGroup` is required, and is the name of the resource group for your
  VMs and storage.
* `tenantID` is required, and is either the domain or UUID for your active
  directory account within Azure.
* `storageAccount` is required, and is the name of the storage account where
  your disks will be created.
* `storageAccessKey` is required, and is a valid access key associated with the
  `storageAccount`.
* `clientID` is required, and is the UUID of your client, which was created as
  an App Registration within your Azure active directory account.
* `clientSecret` is required if `certPath` is not provided instead. It is a
  valid access key associated with `clientID`, and is managed as part of the App
  Registration.
* `certPath` is an alternative to `clientSecret`, contains the location of a
  PKCS encoded RSA private key associated with `clientID`.
* `container` is optional, and specifies the name of an existing container
  within `storageAccount`. This container must already exist and is not created
  automatically.
* `useHTTPS` is optional, and is a boolean value on whether to use HTTPS when
  communicating with the Azure storage endpoint.

#### Runtime Behavior
* The `container` config option defaults to `vhds`, and this container is
  present by default in Azure. Changing this option is only necessary if you
  want to use a different container within your storage account.
* Volume Attach/Detach operations in Azure take a long time, sometimes greater
  than 60 seconds, which is libStorage's default task timeout. When the timeout
  is hit, libStorage returns information to the caller about a queued task, and
  that the task is still running. This may cause issues for upstream callers.
  It is *highly* recommended to adjust this default timeout to 120 seconds by
  setting the `libstorage.server.tasks.exeTimeout` property. This is done in
  the `Examples` section below.

#### Activating the Driver
To activate the Azure UD driver please follow the instructions for
[activating storage drivers](./config.md#storage-drivers), using `azureud` as
the driver name.

#### Troubleshooting
* After creating your app registration, you must go into the
  `Required Permissions` tab and grant access to "Windows Azure Service
  Management API". Choose the delegated permission for accessing as organization
  users.
* You must also grant your app registration access to your subscription, by
  going to Subscriptions->Your `subscriptionID`->Access Control (IAM). From
  there, add your app registration as a user, which you will have to search for
  by name. Grant the role of "Owner".
* You should carefully check that your VM is compatible with the storage account
  you want to use. For example, if you need Azure Premium storage your machine
  should be of a compatible size (e.g. DS_V2, FS). For more details see the
  available VM [sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/virtual-machines-windows-sizes).  

<a class="headerlink hiddenanchor" name="azure-ud-examples"></a>

#### Examples
Below is a full `config.yml` that works with Azure UD

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: azureud
  server:
    tasks:
      exeTimeout: 120s
    services:
      azure:
        driver: azureud
        azureud:
          subscriptionID: abcdef01-2345-6789-abcd-ef0123456789
          resourceGroup: testgroup
          tenantID: usernamehotmail.onmicrosoft.com
          storageAccount: username
          storageAccessKey: XXXXXXXX
          clientID: 123def01-2345-6789-abcd-ef0123456789
          clientSecret: XXXXXXXX
```

#### Caveats
* Snapshot and Copy functionality is not yet implemented
* The number of disks you can attach to a Virtual Machine depends on its type.
* Good resources for reading about disks in Azure are
  [here](https://docs.microsoft.com/en-us/azure/storage/storage-standard-storage)
  and [here](https://docs.microsoft.com/en-us/azure/storage/storage-about-disks-and-vhds-linux).

## VirtualBox
The VirtualBox driver registers a storage driver named `virtualbox` with the
libStorage service registry and is used by VirtualBox's VMs to connect and
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

!!! note
    For a VirtualBox VM to work successfully, the following three points are of
    the utmost importance:

    1. The SATA controller should be named `SATA`.
    2. The SATA controller's port count must allow for additional
    connections.
    3. The MAC address must not match that of any other VirtualBox VM's
    MAC addresses or the MAC address of the host.

    The REX-Ray `Vagrantfile` has
    [a section](https://github.com/codedellemc/rexray/blob/master/Vagrantfile#L166-L179)
    that automatically configures these options.

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
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: virtualbox
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
