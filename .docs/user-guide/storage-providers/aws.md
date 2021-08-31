# Amazon Web Services

EBS, EFS, S3FS

---

<a name="aws-ebs"></a>

## Elastic Block Storage
The AWS EBS driver registers a storage driver named `ebs` with the
libStorage service registry and is used to connect and manage AWS Elastic Block
Storage volumes for EC2 instances.

!!! note
    For backwards compatibility, the driver also registers a storage driver
    named `ec2`. The use of `ec2` in config files is deprecated but functional.
    The `ec2` driver **will be removed in 0.7.0**, at which point all instances
    of `ec2` in config files must use `ebs` instead.

!!! note
    The EBS driver does not yet support snapshots or tags, as previously
    supported in Rex-Ray v0.3.3.
    
!!! note
    Due to issues with device naming, it is currently not possible to run the rexray/ebs 
    plugin on 5th generation (C5 and M5) instances in AWS. See [here](https://github.com/AVENTER-UG/rexray/issues/1104)
    for more information.

The EBS driver is made possible by the
[official Amazon Go AWS SDK](https://github.com/aws/aws-sdk-go.git).

### Requirements

* AWS account
* VPC - EBS can be accessed within VPC
* AWS Credentials

### Configuration
The following is an example with all possible fields configured.  For a running
example see the [Examples](./aws.md#aws-ebs-examples) section.

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

#### Configuration Notes
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
- `useLargeDeviceRange` specifies if REX-Ray should use largest available
  device range `/dev/xvd[b-c][a-z]` for EBS volumes. By default this
  parameter is `false`, so AWS recommended device range `/dev/xvd[f-p]` is used.
  If this parameter is defined it must be defined both server and client-side.
  See
  [AWS documentation on device naming](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html)
  for more information.
- `nvmeBinPath` specifies the path to the `nvme` tool that's used with
  instances with nvme storage.

For information on the equivalent environment variable and CLI flag names
please see the section on how non top-level configuration properties are
[transformed](./../servers/libstorage.md#configuration-properties).

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

### Activating the Driver
To activate the AWS EBS driver please follow the instructions for
[activating storage drivers](./../servers/libstorage.md#storage-drivers),
using `ebs` as the driver name.

### Troubleshooting
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

<a name="aws-ebs-examples"></a>

### Examples
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

### NVMe Support
Support for NVMe requires a `udev` rule to alias the NVMe device to the path
REX-Ray expects as a mount point. A similar udev rule is built into the Amazon 
Linux AMI already, and trivial to add to other linux distributions.

The following is an example of the `udev` rule that must be in place:

```shell
# /etc/udev/rules.d/999-aws-ebs-nvme.rules
# ebs nvme devices
KERNEL=="nvme[0-9]*n[0-9]*", ENV{DEVTYPE}=="disk", ATTRS{model}=="Amazon Elastic Block Store", PROGRAM="/usr/local/bin/ebs-nvme-mapping /dev/%k", SYMLINK+="%c"
```

This script is a helper for creating the required device aliases required by
REX-Ray to support NVMe:

```shell
#!/bin/bash
#/usr/local/bin/ebs-nvme-mapping
vol=$(/usr/sbin/nvme id-ctrl --raw-binary "${1}" | \
      cut -c3073-3104 | tr -s ' ' | sed 's/ $//g')
vol=${vol#/dev/}
[ -n "${vol}" ] && echo "${vol/xvd/sd} ${vol/sd/xvd}"
```

<a name="aws-efs"></a>

## Elastic File System
The AWS EFS driver registers a storage driver named `efs` with the
libStorage service registry and is used to connect and manage AWS Elastic File
Systems.

### Requirements

* AWS account
* VPC - EFS can be accessed within VPC
* AWS Credentials

### Configuration
The following is an example with all possible fields configured.  For a running
example see the [Examples](./aws.md#aws-efs-examples) section.

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

#### Configuration Notes
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
[transformed](./../servers/libstorage.md#configuration-properties).

### Runtime Behavior

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

### Activating the Driver
To activate the AWS EFS driver please follow the instructions for
[activating storage drivers](./../servers/libstorage.md#storage-drivers),
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

<a name="aws-efs-examples"></a>

### Examples
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

<a name="aws-s3fs"></a>

## Simple Storage Service
The AWS S3FS driver registers a storage driver named `s3fs` with the
libStorage service registry and provides the ability to mount Amazon Simple
Storage Service (S3) buckets as filesystems using the
[`s3fs`](https://github.com/s3fs-fuse/s3fs-fuse) FUSE command.

Unlike the other AWS-related drivers, the S3FS storage driver does not need
to deployed or used by an EC2 instance. Any client can take advantage of
Amazon's S3 buckets.

### Requirements
* AWS account
* The [`s3fs`](https://github.com/s3fs-fuse/s3fs-fuse) FUSE command must be
present on client nodes.

### Configuration
The following is an example with all possible fields configured.  For a running
example see the [Examples](./aws.md#aws-s3fs-examples) section.

### Server-Side Configuration
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

### Client-Side Configuration
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
[transformed](./../servers/libstorage.md#configuration-properties).

### Runtime Behavior
The AWS S3FS storage driver can create new buckets as well as remove existing
ones. Buckets are mounted to clients as filesystems using the
[`s3fs`](https://github.com/s3fs-fuse/s3fs-fuse) FUSE command. For clients
to correctly mount and unmount S3 buckets the `s3fs` command should be in
the path of the executor or configured via the `s3fs.cmd` property in the
client-side REX-Ray configuration file.

The client must also have access to the AWS credentials used for mounting and
unmounting S3 buckets. These credentials can be stored in the client-side
REX-Ray configuration file or via
[any means available](https://github.com/s3fs-fuse/s3fs-fuse/wiki/Fuse-Over-Amazon)
to the `s3fs` command.


### Activating the Driver
To activate the AWS S3FS driver please follow the instructions for
[activating storage drivers](./../servers/libstorage.md#storage-drivers),
using `s3fs` as the driver name.

<a name="aws-s3fs-examples"></a>

### Examples
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
