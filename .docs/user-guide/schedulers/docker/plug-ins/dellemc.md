# Dell EMC

Isilon, ScaleIO

---

<a name="dell-emc-isilon"></a>

## Isilon
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

### Requirements
The Isilon plug-in requires that nfs utilities be installed on the
same host on which Docker is running. You should be able to mount an
nfs export to the host.

### Privileges
The Isilon plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

### Configuration
The following environment variables can be used to configure the Isilon
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`ISILON_ENDPOINT` | The Isilon web interface endpoint | | ✓
`ISILON_GROUP` | The group to use when creating a volume | group of the user specified in the configuration |
`ISILON_INSECURE` | Flag for insecure gateway connection | `false` |
`ISILON_USERNAME` | Isilon user for connection | | ✓
`ISILON_PASSWORD` | Isilon password | | ✓
`ISILON_VOLUMEPATH` | The path for volumes (eg: /ifs/rexray) | | ✓
`ISILON_NFSHOST` | The host or ip of your isilon nfs server | | ✓
`ISILON_DATASUBNET` | The subnet for isilon nfs data traffic | | ✓
`ISILON_QUOTAS` | Wanting to use quotas with isilon? | `false` |
`HTTP_PROXY` | Address of HTTP proxy server to gain access to API endpoint | |

<a name="dell-emc-scaleio"></a>

## ScaleIO
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

### Requirements
The ScaleIO plug-in requires that the SDC toolkit must be installed on the
same host on which Docker is running.

### Privileges
The ScaleIO plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
 | `/bin/emc`
 | `/opt/emc/scaleio/sdc`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

### Configuration
The following environment variables can be used to configure the ScaleIO
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`REXRAY_FSTYPE` | The type of file system to use | `xfs` |
`SCALEIO_ENDPOINT` | The ScaleIO gateway endpoint | | ✓
`SCALEIO_GUID` | The ScaleIO client GUID | |
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
`HTTP_PROXY` | Address of HTTP proxy server to gain access to API endpoint | |
