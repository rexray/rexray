CSI-NFS
-------

CSI-NFS is an implementation of a
[CSI](https://github.com/container-storage-interface) plugin for NFS volumes.

It is structured such that it can be compiled into a standalone golang binary
that can be executed to meet the requirements of a CSI plugin. Furthermore, the
core NFS logic is separated into a `nfs` go package that can be imported for use
by other programs.

Installation
-------------

You'll need a working [Go](https://golang.org) installation. From there,
download and installation is as simple as:

`go get github.com/codedellemc/csi-nfs`

This will download the source to `$GOPATH/src/github.com/codedellemc/csi-nfs`,
and will build install the binary `csi-nfs` to `$GOPATH/bin/csi-nfs`.

Starting the plugin
-------------------

In order to execute the binary, you **must** set the env var `CSI_ENDPOINT`. CSI
is intended to only run over UNIX domain sockets, so a simple way to set this
endpoint to a `.sock` file in the same directory as the project is

`export CSI_ENDPOINT=unix://$(go list -f '{{.Dir}}' github.com/codedellemc/csi-nfs)/csi-nfs.sock`

With that in place, you can start the plugin
(assuming that $GOPATH/bin is in your $PATH):

```sh
$ ./csi-nfs
INFO[0000] .Serve                                        name=csi-nfs
```

Use ctrl-C to exit.

You can enable debug logging (all logging goes to stdout) by setting the
`X_CSI_NFS_DEBUG` env var. It doesn't matter what value you set it to, just that
it is set. For example:

```sh
$ X_CSI_NFS_DEBUG= ./csi-nfs
INFO[0000] .Serve                                        name=csi-nfs
DEBU[0000] Added Controller Service
DEBU[0000] Added Node Service
^CINFO[0002] Shutting down server
```

Configuring the plugin
----------------------

The behavior of CSI-NFS can be modified with the following environment variables

| name | purpose | default |
| - | - | - |
| CSI_ENDPOINT | Set path to UNIX domain socket file | n/a |
| X_CSI_NFS_DEBUG | enable debug logging to stdout | n/a |
| X_CSI_NFS_NODEONLY | Only run the Node Service (no Controller service) | n/a |
| X_CSI_NFS_CONTROLLERONLY | Only run the Controller Service (no Node service) | n/a |

Note that the Identity service is required to always be running, and that the
default behavior is to also run both the Controller and the Node service

Using the plugin
----------------

All communication with the plugin is done via gRPC. The easiest way to interact
with a CSI plugin via CLI is to use the `csc` tool found in
[GoCSI](https://github.com/codedellemc/gocsi).

You can install this tool with:

```sh
go get github.com/codedellemc/gocsi
go install github.com/codedellemc/gocsi/csc
```

With $GOPATH/bin in your $PATH, you can issue commands using the `csc` command.
You will want to use a separate shell from where you are running the `csi-nfs`
binary, and as such you will once again need to do:

`export CSI_ENDPOINT=unix://$(go list -f '{{.Dir}}' github.com/codedellemc/csi-nfs)/csi-nfs.sock`

Here are some sample commands:

```sh
$ csc gets
0.1.0
$ csc getp
csi-nfs	0.1.0
$ csc cget
LIST_VOLUMES
$ showmount -e 192.168.75.2
Exports list on 192.168.75.2:
	/data                             192.168.75.1
$ csc mnt -targetPath /tmp/mnt -mode 1 host=192.168.75.2 export=/data
$ ls -al /tmp/mnt
total 1
drwxr-xr-x   2 root  wheel    18 Jul 22 20:25 .
drwxrwxrwt  85 root  wheel  2890 Aug 17 15:32 ..
-rw-r--r--   1 root  wheel     0 Jul 22 20:25 test
$ csc umount -targetPath /tmp/mnt host=192.168.75.2 export=/data
$ ls -al /tmp/mnt
total 0
drwxr-xr-x   2 travis  wheel    68 Aug 16 15:01 .
drwxrwxrwt  85 root    wheel  2890 Aug 17 15:32 ..
```
