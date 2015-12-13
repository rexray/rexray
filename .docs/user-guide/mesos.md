#Mesos

Pooling storage has never been easier...

---

### Overview
`Mesos` currently includes two containerizer options for tasks.  The first is
the `Docker` containizer that leverages a Docker runtime to run containers.  In
this case, refer to the
[Docker](/user-guide/docker.md) page for more information.  Otherwise,
the Mesos containerizer is used and the following is important for it.

`Mesos 0.23+` includes modules that enables extensibility for different
portions the architecture.  The [dvdcli](https://github.com/emccode/dvdcli) and
[mesos-module-dvdi](https://github.com/emccode/mesos-module-dvdi) projects
enable external volume support with Mesos.

## Examples
Below is an example `config.yml` file that can be used.  In this case we are
showing a `virtualbox` driver configuration, but you can use anything here.  We
do suggest two optional options for the `messos-module-dvdi`.  Setting the
`volume.mount.preempt` flag ensures any host can pre-empt control of a volume
from other hosts.  The `volume.unmount.ignoreusedcount` ensures that
`mesos-module-dvdi` is authoritative when it comes to deciding when to unmount
volumes.

```yaml
rexray:
  storageDrivers:
  - virtualbox
  volume:
    mount:
      preempt: true
    unmount:
      ignoreusedcount: true
virtualbox:
  endpoint: http://yourlaptop:18083
  volumePath: /Users/youruser/VirtualBox Volumes
  controllerName: SATA
```      

REX-Ray must be running as a service to serve requests from Docker.  This can
be done by running `rexray start`.  Make sure you restart REX-Ray if you make
configuration changes.

    root@ubuntu:/home/ubuntu# ./rexray start
    Starting REX-Ray...SUCESS!

      The REX-Ray daemon is now running at PID 18141. To
      shutdown the daemon execute the following command:

        sudo /home/ubuntu/rexray stop

Following this you can now leverage volumes with Docker.

### Mesos with Volumes

Run containers with volumes (0.23+).  The `dvdcli` is used by `mesos-module-dvdi` to call `REX-Ray`.  The following is an example of this.

    dvdcli mount --volumedriver=rexray --volumename=test123456789  \
      --volumeopts=size=5 --volumeopts=iops=150 --volumeopts=volumetype=io1 \
      --volumeopts=newFsType=ext4 --volumeopts=overwritefs=true


You can use frameworks like Marathon to specify external volumes.

    "env": {
      "DVDI_VOLUME_NAME": "testing",
      "DVDI_VOLUME_DRIVER": "platform1",
      "DVDI_VOLUME_OPTS": "size=5,iops=150,volumetype=io1,newfstype=ext4,overwritefs=false",
      "DVDI_VOLUME_NAME1": "testing2",
      "DVDI_VOLUME_DRIVER1": "platform2",
      "DVDI_VOLUME_OPTS1": "size=6,volumetype=gp2,newfstype=xfs,overwritefs=true"
    }
