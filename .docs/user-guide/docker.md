# Docker

Build, ship, run on storage made easy

---

### Overview
`REX-Ray` has a `Docker Volume Driver` which is compatible with 1.7+.

It is suggested that you are running `Docker 1.9.1+` with `REX-Ray` especially
if you are sharing volumes between containers.

## Examples
Below is an example `config.yml` that can be used.  The `volume.mount.preempt`
is an optional parameter here which enables any host to take control of a
volume irrespective of whether other hosts are using the volume.  If this is
set to `false` then mostly plugins ensure `safeety` first for locking the
volume.

```yaml
rexray:
  storageDrivers:
  - virtualbox
  volume:
    mount:
      preempt: true
virtualbox:
  endpoint: http://yourlaptop:18083
  volumePath: /Users/youruser/VirtualBox Volumes
  controllerName: SATA
```

REX-Ray must be running as a service to serve requests from Docker.  This can be
 done by running `rexray start`.  Make sure you restart REX-Ray if you make
 configuration changes.

    root@ubuntu:/home/ubuntu# ./rexray start
    Starting REX-Ray...SUCESS!

      The REX-Ray daemon is now running at PID 18141. To
      shutdown the daemon execute the following command:

        sudo /home/ubuntu/rexray stop

Following this you can now leverage volumes with Docker.

### Docker with Volumes

Run containers with volumes (1.7+)

    docker run -ti --volume-driver=rexray -v test:/test busybox

Create volume with options (1.8+)

    docker volume create --driver=rexray --opt=size=5 --name=test

### Extra Options
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
If you restart the REX-Ray instance while volumes are mounted with Docker,
then you should also be resetting Docker.  The volume mount accounting will
be out of sync unless REX-Ray this happens.  In the case of sharing volumes
between containers, problems will arise when stopping the first container since
the volume will be unmounted pre-maturely.
