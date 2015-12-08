# Docker

Build, ship, run on storage made easy

---

### Overview
`Docker 1.7` established a Volume Driver API that enables persistent volumes
to be orchestrated with containers.  `REX-Ray` has a `Docker Volume Driver`
which is compatible with 1.7+.

## Examples
REX-Ray must be running as a service to serve requests from Docker.  This can be done by running `rexray start`.  

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
newFsType|FS Type for Volume if filesystem unknown
overwriteFs|Overwrite existing known filesystem
volumeName|Create from an existing volume name
volumeID|Creat from an existing volume ID
snapshotName|Create from an existing snapshot name
snapshotID|Create from an existing snapshot ID
