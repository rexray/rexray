# Mesos

The sphere of influence

---

## Overview
In Mesos the frameworks are responsible for receiving requests from
consumers and then proceeding to schedule and manage tasks. While some
frameworks, like Marathon, are open to run any workload for sustained periods
of time, others are use case specific, such as Cassandra. Frameworks may
also receive requests from other platforms in addition to schedulers instead of
consumers such as Cloud Foundry, Kubernetes, and Swarm.

Once a resource offer is accepted from Mesos, tasks are launched to support the
associated workloads. These tasks are eventually distributed to Mesos agents in
order to spin up containers.

REX-Ray enables on-demand storage allocation for agents receiving tasks via
two deployment configurations:

 1. Docker Containerizer with Marathon

 2. Mesos Containerizer with Marathon

### Docker Containerizer with Marathon
When the framework leverages the Docker containerizer, Docker and REX-Ray
should both already be configured and working. The following example shows
how to use Marathon in order to bring an application online with external
volumes:

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
Mesos 0.23+ includes modules that enable extensibility for different
portions of the architecture. The [dvdcli](https://github.com/codedellemc/dvdcli) and
[mesos-module-dvdi](https://github.com/codedellemc/mesos-module-dvdi) projects are
required to enable external volume support with the native containerizer.

The next example is similar to the one above, except in this instance the
native containerizer is preferred and volume requests are handled by the
`env` section.

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

This example also illustrates several important settings for the native method.
While the VirtualBox driver is being used, any validated storage platform
should work. Additionally, there are two options recommended for this type of
configuration:

 Property | Recommendation
 ---------|---------------
 `libstorage.integration.volume.operations.mount.preempt` | Setting this flag to true ensures any host can preempt control of a volume from other hosts
 `libstorage.integration.volume.operations.unmount.ignoreUsedCount` | Enabling this flag declares that `mesos-module-dvdi` is the authoritative source for deciding when to unmount volumes

Please refer to the libStorage documentation for more information on
[Volume Configuration](http://libstorage.readthedocs.io/en/stable/user-guide/config/#volume-configuration)
options.

!!! note "note"

    The `libstorage.integration.volume.operations.remove.disable` property can
	prevent the scheduler from removing volumes. Setting this flag to `true` is
	recommended when using Mesos with Docker 1.9.1 or earlier.

```yaml
libstorage:
  service: virtualbox
  integration:
    volume:
      operations:
        mount:
          preempt: true
        unmount:
          ignoreusedcount: true
        remove:
          disable: true
virtualbox:
  volumePath: $HOME/VirtualBox/Volumes
```
