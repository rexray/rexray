# Local

CSI-NFS, CSI-VFS

---

<a name="csi-nfs"></a>

## Network Filesystem
The [CSI-NFS](https://github.com/codedellemc/csi-nfs) project provides
support for network filesystem (NFS) storage.

### Installation
The plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/csi-nfs
```

### Privileges
The plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

### Configuration
The following environment variables can be used to configure the plug-in:

| Environment Variable | Description | Default | Required |
|---------------------|-------------|---------|---------|
| `X_CSI_NFS_VOLUMES` | A list of NFS volume mappings | | |


<a name="csi-vfs"></a>

## Virtual Filesystem
The [CSI-VFS](https://github.com/thecodeteam/csi-vfs) project provides
support for virtual filesystem (VFS) storage.

### Installation
The plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/csi-vfs
```

### Privileges
The plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`
