# CSI-VFS
CSI-VFS is a Container Storage Interface
([CSI](https://github.com/container-storage-interface/spec)) plug-in
that provides virtual filesystem (VFS) support.

This project may be compiled as a stand-alone binary using Golang that,
when run, provides a valid CSI endpoint. This project can also be
vendored or built as a Golang plug-in in order to extend the functionality
of other programs.

## Installation
CSI-VFS can be installed with Go and the following command:

```bash
$ go get github.com/thecodeteam/csi-vfs
```

The resulting binary will be installed to `$GOPATH/bin/csi-vfs`.

## Starting the Plug-in
Before starting the plug-in please set the environment variable
`CSI_ENDPOINT` to a valid Go network address such as `csi.sock`:

```bash
$ CSI_ENDPOINT=csi.sock csi-vfs
INFO[0000] serving                                       address="unix://csi.sock" service=csi-vfs
```

The server can be shutdown by using `Ctrl-C` or sending the process
any of the standard exit signals.

## Using the Plug-in
The CSI specification uses the gRPC protocol for plug-in communication.
The easiest way to interact with a CSI plug-in is via the Container
Storage Client (`csc`) program provided via the
[GoCSI](https://github.com/thecodeteam/gocsi) project:

```bash
$ go get github.com/thecodeteam/gocsi
$ go install github.com/thecodeteam/gocsi/csc
```

## Configuring the Plug-in
The VFS plug-in attempts to approximate the normal workflow of a storage platform
by having separate directories for volumes, devices, and private mounts. These
directories can be configured with the following environment variables:

| Name | Default | Description |
|------|---------|-------------|
| `X_CSI_VFS_DATA` | `$HOME/.csi-vfs` | The root data directory |
| `X_CSI_VFS_VOL` | `$X_CSI_VFS_DATA/vol` | Where volumes (directories) are created |
| `X_CSI_VFS_VOL_GLOB` | `*` | The pattern used to match volumes in `$X_CSI_VFS_VOL` |
| `X_CSI_VFS_DEV` | `$X_CSI_VFS_DATA/dev` | A directory from `$X_CSI_VFS_VOL` is bind mounted to an eponymous directory in this location when `ControllerPublishVolume` is called |
| `X_CSI_VFS_MNT` | `$X_CSI_VFS_DATA/mnt` | A directory from `$X_CSI_VFS_DEV` is bind mounted to an eponymous directory in this location when `NodePublishVolume` is called |
