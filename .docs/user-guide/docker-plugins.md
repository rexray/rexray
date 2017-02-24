# Docker Volume Plug-ins

Plug it in, plug it in...

---

## Overview
This page reviews the REX-Ray Docker volume plug-ins, available for
Docker 1.13+.

## Getting Started
This section describes how to get started with REX-Ray Docker volume plug-ins!

### Installation
Docker plug-ins can be installed with following command:

```bash
$ docker plugin install rexray/driver[:version]
```

In the above command line, if  `[:version]` is omitted, it's equivalent to
the following command:

```bash
$ docker plugin install rexray/driver:latest
```

The `latest` tag refers to the most recent, GA version of a plug-in. The
`[:version]` component is known as a Docker _tag_. It follows the semantic
versioning model. However, in addition to `latest`, there is also the `edge`
tag which refers to the most recent version built from the `master` development
branch.

!!! note "note"
    Please note that most of REX-Ray's plug-ins must be configured and
    installed at the same time since Docker starts the plug-in when installed.
    Otherwise the plug-in will fail since it is not yet configured. Please
    see the sections below for platform-specific configuration options.

### Configuration
Docker volume plug-ins are configured via environment variables, and all
REX-Ray plug-ins share the following, common configuration options:

Environment Variable | Description | Default Value
---------------------|-------------|--------------
`REXRAY_FSTYPE` | The type of file system to use | `ext4`
`REXRAY_LOGLEVEL` | The log level | `warn`
`REXRAY_PREEMPT` | Enable preemption | `false`

## Amazon
REX-Ray has plug-ins for multiple Amazon Web Services (AWS) storage services.

### Elastic Block Service
The EBS plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/ebs \
  EBS_ACCESSKEY=abc \
  EBS_SECRETKEY=123
```

#### Privileges
The EBS plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

#### Configuration
The following environment variables can be used to configure the EBS
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`EBS_ACCESSKEY` | The AWS access key | | ✓
`EBS_SECRETKEY` | The AWS secret key | | ✓
`EBS_REGION` | The AWS region | `us-east-1` |

### Elastic File System
The EFS plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/efs \
  EFS_ACCESSKEY=abc \
  EFS_SECRETKEY=123 \
  EFS_SECURITYGROUPS="sg-123 sg-456" \
  EFS_TAG=rexray
```

#### Requirements
The EFS plug-in requires that nfs utilities be installed on the
same host on which Docker is running. You should be able to mount an
nfs export to the host.

#### Privileges
The EFS plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

#### Configuration
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

### Simple Storage Service (S3)
The S3FS plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/s3fs \
  S3FS_ACCESSKEY=abc \
  S3FS_SECRETKEY=123
```

#### Privileges
The S3FS plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

#### Configuration
The following environment variables can be used to configure the S3FS
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`S3FS_ACCESSKEY` | The AWS access key | | ✓
`S3FS_SECRETKEY` | The AWS secret key | | ✓
`S3S_REGION` | The AWS region | |

## Dell EMC
REX-Ray includes plug-ins for several Dell EMC storage platforms.

### Isilon
The Isilon plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/isilon \
  ISILON_ENDPOINT=https://isilon:8080 \
  ISILON_USERNAME=user \
  ISILON_PASSWORD=pass \
  ISILON_VOLUMEPATH=/ifs/rexray \
  ISILON_NFSHOST=isilon_ip \
  ISILON_DATASUBNET=192.168.1.0/24
```

#### Requirements
The Isilon plug-in requires that nfs utilities be installed on the
same host on which Docker is running. You should be able to mount an
nfs export to the host.

#### Privileges
The Isilon plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

#### Configuration
The following environment variables can be used to configure the Isilon
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`ISILON_ENDPOINT` | The Isilon web interface endpoint | | ✓
`ISILON_INSECURE` | Flag for insecure gateway connection | `false` |
`ISILON_USERNAME` | Isilon user for connection | | ✓
`ISILON_PASSWORD` | Isilon password | | ✓
`ISILON_VOLUMEPATH` | The path for volumes (eg: /ifs/rexray) | | ✓
`ISILON_NFSHOST` | The host or ip of your isilon nfs server | | ✓
`ISILON_DATASUBNET` | The subnet for isilon nfs data traffic | | ✓
`ISILON_QUOTAS` | Wanting to use quotas with isilon? | `false` |

### ScaleIO
The ScaleIO plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/scaleio \
  SCALEIO_ENDPOINT=https://gateway/api \
  SCALEIO_USERNAME=user \
  SCALEIO_PASSWORD=pass \
  SCALEIO_SYSTEMNAME=scaleio \
  SCALEIO_PROTECTIONDOMAINNAME=default \
  SCALEIO_STORAGEPOOLNAME=default
```

#### Requirements
The ScaleIO plug-in requires that the SDC toolkit must be installed on the
same host on which Docker is running.

#### Privileges
The ScaleIO plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
 | `/bin/emc`
 | `/opt/emc/scaleio/sdc`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

#### Configuration
The following environment variables can be used to configure the ScaleIO
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`REXRAY_FSTYPE` | The type of file system to use | `xfs` |
`SCALEIO_ENDPOINT` | The ScaleIO gateway endpoint | | ✓
`SCALEIO_INSECURE` | Flag for insecure gateway connection | `true` |
`SCALEIO_USECERTS` | Flag indicating to require certificate validation | `false` |
`SCALEIO_USERNAME` | ScaleIO user for connection | | ✓
`SCALEIO_PASSWORD` | ScaleIO password | | ✓
`SCALEIO_SYSTEMID` | The ID of the ScaleIO system to use | | If `SCALEIO_SYSTEMID` is omitted
`SCALEIO_SYSTEMNAME` | The name of the ScaleIO system to use | | If `SCALEIO_SYSTEMNAME` is omitted
`SCALEIO_PROTECTIONDOMAINID` | The ID of the protection domain to use | | If `SCALEIO_PROTECTIONDOMAINNAME` is omitted
`SCALEIO_PROTECTIONDOMAINNAME` | The name of the protection domain to use | | If `SCALEIO_PROTECTIONDOMAINID` is omitted
`SCALEIO_STORAGEPOOLID` | The ID of the storage pool to use | | If `SCALEIO_STORAGEPOOLNAME` is omitted
`SCALEIO_STORAGEPOOLNAME` | The name of the storage pool to use | | If `SCALEIO_STORAGEPOOLID` is omitted
`SCALEIO_THINORTHICK` | The provision mode `(Thin|Thick)Provisioned` | |
`SCALEIO_VERSION` | The version of ScaleIO system | |

## Google
REX-Ray ships with plug-ins for Google Compute Engine (GCE) as well.

### GCE Persistent Disk
The GCEPD plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/gcepd \
  GCEPD_TAG=rexray
```

#### Requirements
The GCEPD plug-in requires that GCE compute instance has Read/Write Cloud API
access to the Compute Engine and Storage services.

#### Privileges
The GCEPD plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

#### Configuration
The following environment variables can be used to configure the GCEPD
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`GCEPD_DEFAULTDISKTYPE` | The default disk type to consume | `pd-ssd` |
`GCEPD_PROJECTID` | GCE Project ID | `true` |
`GCEPD_REGION` | GCE Region | `false` |
`GCEPD_TAG` | Only use volumes that are tagged with a label | |
`GCEPD_ZONE` | GCE Availability Zone | |

## Examples
This section reviews examples of how to use the REX-Ray Docker Volume plug-ins.
For the purposes of the examples the EBS plug-in will be demonstrated, but
each example would work for any of the plug-ins above.

### Create a volume
The following example illustrates creating a volume:

```bash
$ docker volume create --driver rexray/ebs --name test-vol-1
```

Verify the volume was successfully created by listing the volumes:

```bash
$ docker volume ls
DRIVER          VOLUME NAME
rexray/ebs      test-vol-1
```

### Inspect a volume
The following example illustrates inspecting a volume:

```bash
$ docker volume inspect test-vol-1
```

```json
[
    {
        "Driver": "rexray/ebs",
        "Labels": {},
        "Mountpoint": "/var/lib/docker/plug-ins/9f30ec546a4b1bb19574e491ef3e936c2583eda6be374682eb42d21bbeec0dd8/rootfs",
        "Name": "test-vol-1",
        "Options": {},
        "Scope": "global",
        "Status": {
            "availabilityZone": "default",
            "fields": null,
            "iops": 0,
            "name": "test-vol-1",
            "server": "ebs",
            "service": "ebs",
            "size": 16,
            "type": "default"
        }
    }
]
```

### Use a volume
The following example illustrates using a volume:

```bash
$ docker run -v test-vol-1:/data busybox mount | grep "/data"
/dev/xvdf on /data type ext4 (rw,seclabel,relatime,nouuid,attr2,inode64,noquota)
```

### Remove a volume
The following example illustrates removing a volume created:

```bash
$ docker volume rm test-vol-1
```

Validate the volume was deleted successfully by listing the volumes:

```bash
$ docker volume ls
DRIVER              VOLUME NAME
```
