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

```shell
$ docker plugin install rexray/driver[:version]
```

In the above command line, if  `[:version]` is omitted, it's equivalent to
the following command:

```shell
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

### Elastic File System

### Simple Storage Service

## Ceph
REX-Ray includes plug-ins for the following Ceph storage technologies.

### RADOS Block Device

## Dell EMC
REX-Ray includes plug-ins for several Dell EMC storage platforms.

### Isilon

### ScaleIO
The ScaleIO plug-in can be installed with the following command:

```shell
docker plugin install rexray/scaleio
```

#### Requirements
The only requirement for ScaleIO plug-in is that the ScaleIO SDC toolkit must
be installed on the same host on which Docker is running.

#### Configuration
The following environment variables can be used to configure the ScaleIO
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`SCALEIO_ENDPOINT` | The ScaleIO gateway endpoint | | ✓
`SCALEIO_INSECURE` | Flag for insecure gateway connection | `true` |
`SCALEIO_USECERTS` | Flag indicating to require certificate validation | `false` |
`SCALEIO_USERNAME` | ScaleIO user for connection | | ✓
`SCALEIO_PASSWORD` | ScaleIO password | | ✓
`SCALEIO_SYSTEMID` | The ID of the ScaleIO system to use | | ✓
`SCALEIO_SYSTEMNAME` | The name of the ScaleIO system to use | | If `SCALEIO_SYSTEMID` is omitted
`SCALEIO_PROTECTIONDOMAINID` | The ID of the protection domain to use | |
`SCALEIO_PROTECTIONDOMAINNAME` | The name of the protection domain to use | | If `SCALEIO_PROTECTIONDOMAINID` is omitted
`SCALEIO_STORAGEPOOLID` | The ID of the storage pool to use | |
`SCALEIO_STORAGEPOOLNAME` | The name of the storage pool to use | | If `SCALEIO_STORAGEPOOLID` is omitted
`SCALEIO_THINORTHICK` | The provision mode `(Thin|Thick)Provisioned` | |
`SCALEIO_VERSION` | The version of ScaleIO system | |

#### Installation
The following example illustrates how to install version `0.7.20` of the
ScaleIO plug-in:

```shell
$ docker plug-in install rexray/scaleio:0.7.20 \
  REXRAY_FSTYPE=xfs \
  REXRAY_LOGLEVEL=warn \
  REXRAY_PREEMPT=false \
  SCALEIO_ENDPOINT=https://localhost/api \
  SCALEIO_INSECURE=true \
  SCALEIO_USERNAME=admin \
  SCALEIO_PASSWORD=MySCaleio123 \
  SCALEIO_SYSTEMNAME=scaleio \
  SCALEIO_PROTECTIONDOMAINNAME=default \
  SCALEIO_STORAGEPOOLNAME=default
```

The above command prompts the user to acknowledge the plug-ins required
permissions:

```shell
Plug-in "rexray/scaleio:0.7.20" is requesting the following privileges:
 - network: [host]
 - mount: [/dev]
 - mount: [/bin/emc]
 - mount: [/opt/emc/scaleio/sdc]
 - allow-all-devices: [true]
 - capabilities: [CAP_SYS_ADMIN]
Do you grant the above permissions? [y/N]
```

Once installed, the status of the plug-in can be retrieved like so:

```shell
$ docker plug-in ls
ID                  NAME                            DESCRIPTION                    ENABLED
5c08e5947d8f        rexray/scaleio:0.7.20           REX-Ray for EMC Dell ScaleIO   true
```

#### Create a volume
The following example illustrates creating a volume using the ScaleIO plug-in:

```shell
$ docker volume create --driver rexray/scaleio:0.7.0-20 --name test-vol-1
```

Verify the volume was successfuly created by listing the volumes:

```shell
$ docker volume ls
DRIVER                          VOLUME NAME
rexray/scaleio:0.7.0            test-vol-1
```

#### Inspect a volume
The following example illustrates inspecting a volume created using the
ScaleIO plug-in:

```shell
$ docker volume inspect test-vol-1
[
    {
        "Driver": "rexray/scaleio:0.7.0-20",
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
            "server": "scaleio",
            "service": "scaleio",
            "size": 16,
            "type": "default"
        }
    }
]
```

#### Use a volume
The following example illustrates using a volume created using the
ScaleIO plug-in:

```shell
$ docker run -v test-vol-1:/data busybox mount | grep "/data"
/dev/scinia on /data type xfs (rw,seclabel,relatime,nouuid,attr2,inode64,noquota)
```

#### Remove a volume
The following example illustrates removing a volume created using the
ScaleIO plug-in:

```shell
$ docker volume rm test-vol-1
```

Validate the volume was deleted successfully by listing the volumes:

```shell
$ docker volume ls
DRIVER              VOLUME NAME
```

## Google
REX-Ray ships with plug-ins for Google Compute Engine (GCE) as well.

### GCE Persistent Disk

## Microsoft
Microsoft Azure plug-ins are included with REX-Ray as well.

### Azure Unmanaged Disk
