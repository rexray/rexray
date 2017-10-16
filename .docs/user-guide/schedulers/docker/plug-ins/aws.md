# Amazon Web Services

EBS, EFS, S3FS

---

<a name="aws-ebs"></a>

## Elastic Block Service
The EBS plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/ebs \
  EBS_ACCESSKEY=abc \
  EBS_SECRETKEY=123
```

### Privileges
The EBS plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

### Configuration
The following environment variables can be used to configure the EBS
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`EBS_ACCESSKEY` | The AWS access key | | ✓
`EBS_KMSKEYID` | The encryption key for all volumes that are created with a truthy encryption request field | |
`EBS_MAXRETRIES` | the number of retries that will be made for failed operations by the AWS SDK | 10 |
`EBS_REGION` | The AWS region | `us-east-1` |
`EBS_SECRETKEY` | The AWS secret key | | ✓
`EBS_STATUSINITIALDELAY` | Time duration used to wait when polling volume status | `100ms` |
`EBS_STATUSMAXATTEMPTS` | Number of times the status of a volume will be queried before giving up | `10` |
`EBS_STATUSTIMEOUT` | Maximum length of time that polling for volume status can occur | `2m` |
`EBS_USELARGEDEVICERANGE` | Use largest available device range `/dev/xvd[b-c][a-z]` for EBS volumes | false |
`HTTP_PROXY` | Address of HTTP proxy server to gain access to API endpoint | |

<a name="aws-efs"></a>

## Elastic File System
The EFS plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/efs \
  EFS_ACCESSKEY=abc \
  EFS_SECRETKEY=123 \
  EFS_SECURITYGROUPS="sg-123 sg-456" \
  EFS_TAG=rexray
```

### Requirements
The EFS plug-in requires that nfs utilities be installed on the
same host on which Docker is running. You should be able to mount an
nfs export to the host.

### Privileges
The EFS plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

### Configuration
The following environment variables can be used to configure the EFS
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`EFS_ACCESSKEY` | The AWS access key | | ✓
`EFS_SECRETKEY` | The AWS secret key | | ✓
`EFS_REGION` | The AWS region | |
`EFS_SECURITYGROUPS` | The AWS security groups to bind to | `default` |
`EFS_TAG` | Only consume volumes with tag (tag\volume_name)| |
`EFS_DISABLESESSIONCACHE` | new AWS connection is established with every API call | `false` |
`EFS_STATUSINITIALDELAY` | Time duration used to wait when polling volume status | `1s` |
`EFS_STATUSMAXATTEMPTS` | Number of times the status of a volume will be queried before giving up | `6` |
`EFS_STATUSTIMEOUT` | Maximum length of time that polling for volume status can occur | `2m` |
`HTTP_PROXY` | Address of HTTP proxy server to gain access to API endpoint | |

<a name="aws-s3fs"></a>

## Simple Storage Service
The S3FS plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/s3fs \
  S3FS_ACCESSKEY=abc \
  S3FS_SECRETKEY=123
```

### Privileges
The S3FS plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

### Configuration
The following environment variables can be used to configure the S3FS
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`S3FS_ACCESSKEY` | The AWS access key | | ✓
`S3FS_DISABLEPATHSTYLE` | Disables use of path style for bucket endpoints | `false` |
`S3FS_MAXRETRIES` | the number of retries that will be made for failed operations by the AWS SDK | 10 |
`S3FS_OPTIONS` | Additional options to pass to S3FS | |
`S3FS_REGION` | The AWS region | |
`S3FS_SECRETKEY` | The AWS secret key | | ✓
`HTTP_PROXY` | Address of HTTP proxy server to gain access to API endpoint | |
