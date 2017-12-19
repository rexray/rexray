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

## Examples

To use the `csi-nfs` plug-in, you must first create a volume using the
`docker volume create` command (unless predefined using `X_CSI_NFS_VOLUMES`, see
`Limitations` below), specifying both the NFS server host and the
exported path:

```sh
$ docker volume create -d rexray/csi-nfs -o host=192.168.75.2 -o export=/data test
$ docker volume ls
DRIVER                           VOLUME NAME
rexray/csi-nfs                   test
```

With the volume created, you can then use it in your containers:

```sh
$ docker run -it -v test:/mnt alpine sh
/ # mount | grep nfs
192.168.75.2:/data on /mnt type nfs4 (rw,relatime,vers=4.1,rsize=65536,wsize=65536,namlen=255,hard,proto=tcp,port=0,timeo=600,retrans=2,sec=sys,clientaddr=10.0.2.15,local_lock=none,addr=192.168.75.2)
```

### Limitations

`csi-nfs` is an early proof of concept to demonstrate a Docker managed plug-in functioning with [CSI](https://github.com/container-storage-interface/spec). It is
not intended to replace Docker's
[native NFS](https://docs.docker.com/engine/reference/commandline/volume_create/#driver-specific-options)
functionality.

Currently, the plug-in only supports NFSv4, and does not allow for custom options
to be passed through to the mount command. In other words, the NFS mount command
will always be of the form `mount -t nfs {server}:{export} {mountpath}`.

When doing a `docker volume create` command, the `csi-nfs` plug-in merely creates
a reference to an existing NFS volume with the details you have provided. This
reference is only persisted across the lifetime of the plug-in, which means if the
plug-in is removed and recreated, or upgraded, the volumes previously seen in
`docker volume ls` will disappear. To circumvent this behavior, you can seed the
plug-in with the NFS volume definitions using the `X_CSI_NFS_VOLUMES` env var.
For example:

```sh
$ docker plugin set rexray/csi-nfs X_CSI_NFS_VOLUMES="vol1=192.168.75.2:/nfsshare vol2=192.168.75.2:/share2"
```

The env var can be set during plug-in installation, or the plug-in can be disabled,
the var set, and then enabled. However, currently only the env var or the persisted
volumes from `docker volume create` can be used, which means if
`docker volume create` is used to define an NFS volume, a later setting of
`X_CSI_NFS_VOLUMES` will be ignored.
