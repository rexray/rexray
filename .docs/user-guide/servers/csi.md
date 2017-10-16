# CSI

YEEEEAAAAAAAAHH! Won't get fooled again!

---

The Container Storage Interface ([CSI](https://github.com/container-storage-interface/spec))
specification is an industry-wide collaboration between the companies that
make container orchestration platforms such as Kubernetes, Mesos, and
Docker.

REX-Ray supports CSI and adheres to all of the specification's requirements,
such as idempotency. This means that all existing
[storage providers](../storage-providers.md) support CSI and are
idempotent as well. Not only that, but REX-Ray supports native CSI plug-ins!

| Provider              | Storage Platform  | <center>[Docker](https://docs.docker.com/engine/extend/plugins_volume/)</center> | <center>[CSI](https://github.com/container-storage-interface/spec)</center> | <center>Containerized</center> |
|-----------------------|----------------------|:---:|:---:|:---:|
| Amazon EC2 | [EBS](../storage-providers/aws.md#aws-ebs) | ✓ | ✓ | ✓  |
| | [EFS](../storage-providers/aws.md#aws-efs) | ✓ | ✓ | ✓ |
| | [S3FS](../storage-providers/aws.md#aws-s3fs) | ✓ | ✓ | ✓ |
| Ceph | [RBD](../storage-providers/ceph.md#ceph-rbd) | ✓ | ✓ | ✓ |
| Local | [CSI-BlockDevices](https://github.com/codedellemc/csi-blockdevices) | | ✓ | ✓ |
| | [CSI-NFS](https://github.com/codedellemc/csi-nfs) | ✓ | ✓ | ✓ |
| | [CSI-VFS](https://github.com/codedellemc/csi-vfs) | | ✓ | ✓ |
| Dell EMC | [Isilon](../storage-providers/dellemc.md#dell-emc-isilon) | ✓ | ✓ | ✓ |
| | [ScaleIO](../storage-providers/dellemc.md#dell-emc-scaleio) | ✓ | ✓ | ✓ |
| DigitalOcean | [Block Storage](../storage-providers/digitalocean.md#do-block-storage) | ✓ | ✓ | ✓ |
| FittedCloud | [EBS Optimizer](../storage-providers/fittedcloud.md/#ebs-optimizer) | ✓ | ✓ | |
| Google | [GCE Persistent Disk](../storage-providers/google.md#gce-persistent-disk) | ✓ | ✓ | ✓ |
| Microsoft | [Azure Unmanaged Disk](../storage-providers/microsoft.md#azure-ud) | ✓ | ✓ | ✓ |
| OpenStack | [Cinder](../storage-providers/openstack.md#cinder) | ✓ | ✓ | ✓ |
| VirtualBox | [Virtual Media](../storage-providers/virtualbox.md#virtualbox) | ✓ | ✓ | |

## Start a Server
The REX-Ray CSI server has two modes:

| Server Mode | Description |
|-------------|-------------|
| [Bridge](#bridge-server) | Supports [existing](../storage-providers.md), libStorage-based storage platforms |
| [Native](#native-server) | Supports native CSI plug-ins |

### Bridge Server
This server mode provides a bridge between CSI and the storage platforms supported
by REX-Ray via libStorage drivers.

!!! note "note"

    The rest of this example assumes that REX-Ray is configured properly
    following the same conventions as earlier versions of REX-Ray.

    REX-Ray using CSI is configured no differently than without CSI. Please
    see the [configuration](../config.md) documentation for information on
    getting started with REX-Ray.

The following example illustrates how to start a REX-Ray CSI server in
bridge mode:

```bash
$ X_CSI_DRIVER=libstorage \
  CSI_ENDPOINT=csi.sock rexray start
```

The environment variable `X_CSI_DRIVER` in the above command is
explicitly set to `libstorage`, the name of the bridge driver. This
is also the default value for the `X_CSI_DRIVER` configuration option.

### Native Server
This server mode eschews the storage drivers that target libStorage
and uses only native CSI plug-ins. It's possible to start a REX-Ray
CSI server using native mode that has the Docker module and libStorage
completely disabled:

```bash
$ X_CSI_DRIVER=csi-vfs \
  X_CSI_NATIVE=true \
  CSI_ENDPOINT=csi.sock rexray start
```

The above command looks similar to the one used to start bridge mode
except for a new environment variable:

| Name | Description |
|------|-------------|
| `X_CSI_NATIVE` | A flag that disables both the CSI->libStorage bridge and REX-Ray's embedded libStorage server |

In native mode he environment variable `X_CSI_DRIVER` is set to the
name of a native CSI plug-in, in this case `csi-vfs`. In fact, `X_CSI_DRIVER`
is automatically set to `csi-vfs` if `X_CSI_NATIVE` is set to a truthy
value.

## Use a Client
A CSI client does not care whether or not REX-Ray's CSI server is
operating in bridge or native mode. However, for the purposes of the
examples below, native mode will be used in order to leverage the
[CSI-VFS](https://github.com/codedellemc/csi-vfs) plug-in as it's
portable and works on Linux and Darwin.

First, export the `CSI_ENDPOINT` location for both the server and
client to use.

```bash
$ export CSI_ENDPOINT=csi.sock
```

Next, start a REX-Ray CSI server in native mode (the default CSI plug-in
in native mode is [CSI-VFS](https://github.com/codedellemc/csi-vfs)):

```bash
$ X_CSI_DRIVER=csi-vfs \
  X_CSI_NATIVE=true \
  rexray start &> rexray.log &
```

!!! note "note"

    Please note the above command runs the REX-Ray server in the
    background.

Now that a CSI server is running, it's time to access it. To do that a
CSI client is needed, such as
[`csc`](https://github.com/codedellemc/gocsi/tree/master/csc):

```bash
$ go get github.com/codedellemc/gocsi/csc
```

Once installed, the `csc` program can be used to create a volume:

```bash
$ csc new MyNewVolume
path=/Users/akutz/.csi-vfs/vol/MyNewVolume
```

List volumes:

```bash
$ csc ls
path=/Users/akutz/.csi-vfs/vol/MyNewVolume
```

Publish a volume:

```bash
$ csc att path=/Users/akutz/.csi-vfs/vol/MyNewVolume
path=/Users/akutz/.csi-vfs/dev/MyNewVolume
```

And more! The `csc` program supports all of the CSI specifications
remote procedure calls (RPC). It's a great method to develop and test
against a CSI endpoint.

## Configuration
There are a few CSI-specific configuration options when using REX-Ray:

| Environment Variable | Flag | Default | Description |
|-------|--------|-------------|-----|
| `CSI_ENDPOINT` | | `/run/docker/plugins/rexray.sock` | The endpoint used by the CSI module |
| `X_CSI_DRIVER` | `--csiDriver` | `libstorage` | The name of the CSI plug-in used by the CSI module. If `X_CSI_NATIVE` is set to a truthy value the default value for `X_CSI_DRIVER` becomes `csi-vfs`. |
| `X_CSI_NATIVE` | | `false` | A flag that disables the CSI to libStorage bridge. |

## Multiplexing Docker
All of REX-Ray's managed Docker plug-ins include support for CSI. However, there's
an issue. Today it is not possible to create a managed Docker plug-in that does not
adhere to one of Docker's
[supported plug-in interfaces](https://docs.docker.com/engine/extend/config/#config-field-descriptions).

In short, when Docker starts a plug-in, Docker will validate that the service
hosted on the plug-in's advertised sock file is compatible with the plug-in's
configured interface type. If this is not the case then Docker marks the plug-in
in error.

REX-Ray's managed Docker plug-ins circumvent this restriction by multiplexing a
CSI server -- bridge or native -- and a Docker Volume API server at the same
endpoint.

Here's a quick example showcasing this functionality. First, start
REX-Ray in CSI native mode with debug logging enabled:

```bash
$ X_CSI_NATIVE=true \
  X_CSI_DRIVER=csi-vfs \
  rexray start -l debug
```

Now, in a new terminal use the [`csc` client](#use-a-client) to create
a new volume:

```bash
$ CSI_ENDPOINT=/run/docker/plugins/rexray.sock csc new MyNewVolume
path=/root/.csi-vfs/vol/MyNewVolume
```

So far, so normal. However, this is where the multiplexing comes into
play. Use the `curl` command to issue a Docker Volume API command to
the same socket file:

```bash
$ curl -H "Content-Type: application/json" -XPOST -d '{}' --unix-socket /run/docker/plugins/rexray.sock http://localhost/VolumeDriver.List
{"Volumes":[{"Name":"MyNewVolume"}]}
```

The logs from the first terminal will also indicate when CSI is running
in multiplexed mode:

```bash
INFO[0000] multiplexed csi+docker endpoint               sockFile=/run/docker/plugins/rexray.sock time=1505081594112
```
