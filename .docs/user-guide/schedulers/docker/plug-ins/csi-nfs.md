# CSI NFS

Network File System

---

## Installation
The plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/csi-nfs
```

## Privileges
The plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

## Configuration
The following environment variables can be used to configure the plug-in:

| Environment Variable | Description | Default | Required |
|---------------------|-------------|---------|---------|
| `X_CSI_NFS_VOLUMES` | A list of NFS volume mappings | | |
