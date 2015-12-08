#Mesos

Pooling storage has never been easier...

---

### Overview
`Mesos 0.23+` includes modules that enables extensibility for different
portions the architecture.  The [dvdcli](https://github.com/emccode/dvdcli) and [mesos-module-dvdi](https://github.com/emccode/mesos-module-dvdi) projects enable external volume support with Mesos.

## Examples
REX-Ray must be running as a service to serve requests from Docker.  This can be done by running `rexray start`.  

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
