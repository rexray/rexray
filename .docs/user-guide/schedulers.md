# Schedulers

Scheduling storage one resource at a time...

---

## Overview
This page reviews the scheduling systems supported by `REX-Ray`.

## Docker
`REX-Ray` has a `Docker Volume Driver` which is compatible with 1.7+.

It is suggested that you are running `Docker 1.10.2+` with `REX-Ray` especially
if you are sharing volumes between containers, or you want interactive
volume commands through `docker volume`.

### Example Configuration
Below is an example `config.yml` that can be used.  The `volume.mount.preempt`
is an optional parameter here which enables any host to take control of a
volume irrespective of whether other hosts are using the volume.  If this is
set to `false` then plugins should ensure `safety` first by locking the
volume from to the current owner host. We also specify `docker.size` which will
create all new volumes at the specified size in GB.

```yaml
rexray:
  storageDrivers:
  - virtualbox
  volume:
    mount:
      preempt: true
docker:
  size: 1
virtualbox:
  endpoint: http://yourlaptop:18083
  volumePath: /Users/youruser/VirtualBox Volumes
  controllerName: SATA
```

#### Extra Global Parameters
These are all valid parameters that can be configured for the service.

parameter|description
------|-----------
docker.size|Size in GB
docker.iops|IOPS
docker.volumeType|Type of Volume or Storage Pool
docker.fsType|Type of filesystem for new volumes (ext4/xfs)
docker.availabilityZone|Extensible parameter per storage driver
linux.volume.rootPath|The path within the volume to private mount (/data)
rexray.volume.mount.preempt|Forcefully take control of volumes when requested
rexray.volume.create.disable|Disable the ability for a volume to be created
rexray.volume.remove.disable|Disable the ability for a volume to be removed

Note: With Docker 1.9.1 or below a `rexray.volume.remove.disable` is suggested
since Docker will remove external volumes when containers that are using volumes
are forcefully removed.

### Starting Volume Driver

REX-Ray must be running as a service to serve requests from Docker. This can be
done by running `rexray start`.  Make sure you restart REX-Ray if you make
configuration changes.

    $ sudo rexray start
    Starting REX-Ray...SUCESS!

      The REX-Ray daemon is now running at PID 18141. To
      shutdown the daemon execute the following command:

        sudo rexray stop

Following this you can now leverage volumes with Docker.

### Creating and Using Volumes
There are two ways to interact with volumes. You can use the `docker run`
command in combination with `--volume-driver` for new volumes, or
specify `-v volumeName` by itself for existing volumes. The `--volumes-from`
will also work when sharing existing volumes with a new container.

The `docker volume` sub-command
enables complete management to create, remove, and list existing volumes. All
volumes are returned from the underlying storage platform.

  1. Run containers with volumes (1.7+)

        docker run -ti --volume-driver=rexray -v test:/test busybox

  2. Create volume with options (1.8+)

        docker volume create --driver=rexray --opt=size=5 --name=test

### Extra Volume Create Options
option|description
------|-----------
size|Size in GB
IOPS|IOPS
volumeType|Type of Volume or Storage Pool
volumeName|Create from an existing volume name
volumeID|Creat from an existing volume ID
snapshotName|Create from an existing snapshot name
snapshotID|Create from an existing snapshot ID

### Caveats
If you restart the REX-Ray instance while volumes *are shared between
Docker containers* then problems may arise when stopping one of the containers
sharing the volume.  It is suggested that you avoid stopping these containers
at this point until all containers sharing the volumes can be stopped.  This
will enable the unmount process to proceed cleanly.

## Mesos
In Mesos the frameworks are responsible for receiving requests from
consumers and then proceeding to schedule and manage tasks.  Some frameworks
are open to run any workload for sustained periods of time (ie. Marathon), and
others are use case specific (ie. Cassandra).  Further than this, frameworks can
receive requests from other platforms or schedulers instead of consumers such as
Cloud Foundry, Kubernetes, and Swarm.

Once frameworks decide to accept resource offers from Mesos, tasks are launched
to support workloads.  These tasks eventually make it down to Mesos agents
to spin up containers.  

`REX-Ray` provides the ability for any agent receiving a task to request
storage be orchestrated for that task.  

There are two primary methods that `REX-Ray` functions with Mesos.  It is up to
the framework to determine which is most appropriate.  Mesos (0.26) has two
containerizer options for tasks, `Docker` and `Mesos`.

### Docker Containerizer with Marathon
If the framework uses the Docker containerizer, it is required that both
`Docker` and `REX-Ray` are configured ahead of time and working.  It is best to
refer to the [Docker](#docker) page for more
information.  Once this is configured across all appropriate agents, the
following is an example of using Marathon to start an application with external
volumes.

```json
{
	"id": "nginx",
	"container": {
		"docker": {
			"image": "million12/nginx",
			"network": "BRIDGE",
			"portMappings": [{
				"containerPort": 80,
				"hostPort": 0,
				"protocol": "tcp"
			}],
			"parameters": [{
				"key": "volume-driver",
				"value": "rexray"
			}, {
				"key": "volume",
				"value": "nginx-data:/data/www"
			}]
		}
	},
	"cpus": 0.2,
	"mem": 32.0,
	"instances": 1
}
```

### Mesos Containerizer with Marathon
`Mesos 0.23+` includes modules that enables extensibility for different
portions the architecture.  The [dvdcli](https://github.com/emccode/dvdcli) and
[mesos-module-dvdi](https://github.com/emccode/mesos-module-dvdi) projects are
required for this method to enable external volume support with the native
containerizer.

The following is a similar example to the one above.  But here we are specifying
to use the the native containerizer and requesting volumes through The `env`
section.

```json
{
  "id": "hello-play",
  "cmd": "while [ true ] ; do touch /var/lib/rexray/volumes/test12345/hello ; sleep 5 ; done",
  "mem": 32,
  "cpus": 0.1,
  "instances": 1,
  "env": {
    "DVDI_VOLUME_NAME": "test12345",
    "DVDI_VOLUME_DRIVER": "rexray",
    "DVDI_VOLUME_OPTS": "size=5,iops=150,volumetype=io1,newfstype=xfs,overwritefs=true"
  }
}
```

This example also comes along with a couple of important settings for the
native method.  This is a `config.yml` file that can be used.  In this case we
are showing a `virtualbox` driver configuration, but you can use anything here.  
We suggest two optional options for the `mesos-module-dvdi`.  Setting the
`volume.mount.preempt` flag ensures any host can preempt control of a volume
from other hosts.  Refer to the [User-Guide](./config.md#preemption) for
more information on preempt.  The `volume.unmount.ignoreusedcount` ensures that
`mesos-module-dvdi` is authoritative when it comes to deciding when to unmount
volumes.

Note: We have added a `rexray.volume.remove.disable` flag to disable the ability
for the scheduler to remove volumes. With Mesos + Docker 1.9.1 this setting
is suggested.

```yaml
rexray:
  storageDrivers:
  - virtualbox
  volume:
    mount:
      preempt: true
    unmount:
      ignoreusedcount: true
    remove:
      disable: true
virtualbox:
  endpoint: http://yourlaptop:18083
  volumePath: /Users/youruser/VirtualBox Volumes
  controllerName: SATA
```
