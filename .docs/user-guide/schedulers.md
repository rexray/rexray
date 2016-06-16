# Schedulers

Scheduling storage one resource at a time...

---

## Overview
This page reviews the scheduling systems supported by REX-Ray.

## Docker
The [Docker documentation](http://libstorage.readthedocs.io/en/stable/user-guide/schedulers/#docker)
can now be found in the libStorage project.

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

REX-Ray provides the ability for any agent receiving a task to request
storage be orchestrated for that task.  

There are two primary methods that REX-Ray functions with Mesos.  It is up to
the framework to determine which is most appropriate.  Mesos (0.26) has two
containerizer options for tasks, `Docker` and `Mesos`.

### Docker Containerizer with Marathon
If the framework uses the Docker containerizer, it is required that both
`Docker` and REX-Ray are configured ahead of time and working.  It is best to
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
