# Usage
Report status, Initiate storage operations

---

## Summary
General format:

```sh
$ rexray [option flags] [command]
```

Get general or command specific help

```sh
$ rexray --help
```

```sh
$ rexray [command] --help
```

## Global Option Flags
These can be invoked with all commands

| Flag | Alternate form | Description |
| --- | --- | --- |
| `-c` | `--config=""` | Path to a custom REX-Ray configuration file |
| `-h` | `--host=""` | designate libStorage host to execute command |
| `-l` | `--logLevel=""` | error, warn, info, debug |
| `-s` | `--service=""` | designate specific libStorage service command pertains to |
| `-v` | `--verbose=true` | emit verbose information to console |

## Commands
Choose one of these REX-Ray commands

| Command | Description |
| --- | --- |
| `version` | return REX-Ray and libStorage versions |
| `env` | return effective current configuration, in form normalized as environment variables |
| `adapter` | return list of configured services, or list of host instances manageable by a service |
| `device` | list or operate (mount,unmount,format) on OS devices |
| `service` | manage REX-Ray running as an OS service |
| `volume` | list or operate (create,remove,attach,detach,mount,unmount) on volumes |
| `help` | return list of commands and global option flags |

## Examples
**Display Version**

```shell
$ rexray version
REX-Ray
-------
Binary: /usr/bin/rexray
Flavor: client+agent+controller
SemVer: 0.8.1
OsArch: Linux-x86_64
Branch: v0.8.1
Commit: 30e9082dd9917f0d218ca981f886d701110ce5f5
Formed: Sat, 25 Feb 2017 03:00:28 UTC

libStorage
----------
SemVer: 0.5.1
OsArch: Linux-x86_64
Branch: v0.8.1
Commit: 35c7b6d96d5f17aa0c0379924615ae22c1ad3d45
Formed: Sat, 25 Feb 2017 02:59:00 UTC
```

**Start REX-Ray Service**

```shell
$ sudo rexray service start
Starting REX-Ray...SUCCESS!

  The REX-Ray daemon is now running at PID 1455. To
  shutdown the daemon execute the following command:

    sudo /usr/bin/rexray stop
```

**List Volumes**

```shell
$ rexray volume ls
ID                     Name       Status       Size
vol-03ff12d1be6e8a65d             attached     8
vol-549bd6d4                      unavailable  20
vol-867e4906           mesos1-sw  available    16
```

| Status | Description |
| --- | --- |
| `attached` | volume is currently attached to this instance |
| `unavailable` | volume is in use attached to another (different) instance |
| `available` | volume is not currently attached to an instance |

**List Configured Drivers**

```
$ rexray adapter types
Name  Driver
ebs   ebs
```

**List Host Instances**

List all host instances manageable by configured services.

```shell
$ rexray adapter instances
ID                   Name                 Provider  Region
i-0e46f04c4b7449bc1  i-0e46f04c4b7449bc1  ebs       us-west-2
```

**Create Volume**

Volume create options. This is a partial list of the most commonly used 
options. To see the complete list of options, 
invoke `rexray volume create --help`

!!! note "note"

    REX-Ray passes these options to 
    [libStorage](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/) 
    which in turn invokes the API of a storage provider. The settings range, 
    and constraints, are generally determined by the storage platform, and 
    can have cross-option interactions (e.g. specifying a high iops settings, 
    reduces the allowed range of size). Cloud providers have been known 
    to change constraints over time and across regions. Please consult the 
    documentation associated with the storage provider should you 
    encounter issues creating a volume. 

| Flag | Range | Description | Mandatory |
| --- | --- | --- | --- |
| `--iops=` | range and support for this option varies by driver | specify performance class of storage | no, default determined by platform |
| `--size=` | supported range varies by driver | create will fail if outside supported range for driver | yes |


```shell
$ rexray volume create mysqldata --size=16
ID                     Name       Status     Size
vol-0714d4b348da9e537  mysqldata  available  16
```

**Mount Volume**

```shell
$ sudo rexray volume mount mysqldata --fsType=ext4
ID                     Name       Status    Size  Path
vol-0762d4b318dade537  mysqldata  attached  16    /var/lib/libstorage/volumes/mysqldata/data
```

| Flag | Range | Description | Mandatory |
| --- | --- | --- | --- |
| `--fsType=` | `ext4,xfs` | file system format for block device mount | no, default=ext4 |
| `--overwriteFS=` | `true,false` | reformat any existing filesytem if true | no, default=false |

**List  Volume Details**

```shell
$ rexray --format=jsonp volume ls mysqldata
[
  {
    "attachmentState": 3,
    "availabilityZone": "us-west-2c",
    "name": "mysqldata",
    "size": 1,
    "status": "available",
    "id": "vol-0ca6d0b4577b3d1b7",
    "type": "standard"
  }
```

**Unmount Volume**

```shell
$ sudo rexray volume unmount mysqldata
ID                     Name       Status     Size
vol-0764d4b348da9e537  mysqldata  available  16
```

**Remove Volume**

```shell
$ rexray volume rm mysqldata
mysqldata
```

**List Devices + Mounts**

```shell
$ rexray device ls
ID  Device      MountPoint
22  /dev/xvda1  /
32  /dev/xvdf   /var/lib/libstorage/volumes/mysqldata
20  devpts      /dev/pts
27  none        /run/lock
30  none        /sys/fs/pstore
28  none        /run/shm
23  none        /sys/fs/cgroup
24  none        /sys/fs/fuse/connections
29  none        /run/user
26  none        /sys/kernel/security
25  none        /sys/kernel/debug
18  proc        /proc
17  sysfs       /sys
31  systemd     /sys/fs/cgroup/systemd
21  tmpfs       /run
19  udev        /dev
```