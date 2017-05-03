# Usage

Report status, Initiate storage operations

---

## Overview
This page reviews how to use the REX-Ray CLI.

```shell
$ rexray [options] commands [flags] [args]
```

Name | Description
-----|------------
`[options]` | REX-Ray's optional, global options, such as log level (`-l|--logLevel`)
`commands` | The commands to execute, such as `rexray volume ls` or `rexray volume create`
`[flags]` | The flags associated with the command. Some flags may be optional, others required. Some flags may be simple switches where other flags require arguments.
`[args]` | The remaining tokens on the command line. These may be arguments to the command. For example, `rexray volume create Vol1 Vol2` has two arguments, `Vol1` and `Vol2`, and are the names of the volume to create. A command's arguments varies by command.


## Getting Help
To print the online help for any command, use the `-?|--help` flag
in conjunction with the command in question.

```shell
$ rexray [command] -?
```

The following example illustrates how to print the online help
for the `rexray volume ls` command:

```shell
$ rexray volume ls -?
List volumes

Usage:
  rexray volume ls [flags]

Aliases:
  ls, l, list, get, inspect

Examples:
rexray volume ls [OPTIONS] [VOLUME...]

Flags:
      --attached    A flag that indicates only volumes attached to this host should be returned
      --available   A flag that indicates only available volumes should be returned
      --path        A flag that indicates only volumes attached to this host should be returned, along with their path info

Global Flags:
  -c, --config string     The path to a custom REX-Ray configuration file
  -n, --dryRun            Show what action(s) will occur, but do not execute them
  -f, --format string     The output format (tmpl, json, jsonp) (default "tmpl")
  -?, --help              Help about the current command
  -h, --host string       The libStorage host.
  -l, --logLevel string   The log level (error, warn, info, debug) (default "warn")
  -q, --quiet             Suppress table headers
  -s, --service string    The libStorage service.
      --template string   The Go template to use when --format is set to 'tmpl'
      --templateTabs      Set to true to use a Go tab writer with the output template (default true)
  -v, --verbose           Print verbose help information
```


## Global Options
As mentioned in the [Overview](#Overview), REX-Ray has several, global options:

Option | Long Form | Description
-------|-----------|------------
`-?` | `--help` | Prints the online help for the current command.
`-c` | `--config` | A path to a custom configuration file. This file will be merged into the default user and global configuration files if they exist.
`-h` | `--host` | The libStorage host used to process the remote parts of the command's workflow.
`-l` | `--logLevel` | The log level. Options include `error`, `warn`, `info`, and `debug`. The default log level is `warn`.
`-s` | `--service` | The libStorage service used to process the remote parts of the command's workflow.
`-v` | `--verbose` | Prints all of the available flags, not just the basic ones.


## Commands
The following commands are available to use with the REX-Ray CLI:

Command | Description
--------|------------
`env` | Prints the effective, current configuration as a list of environment variables.
`token` | The token command is used to create and validate authentication tokens used by libStorage.
`version` | Prints the REX-Ray and libStorage version information for the executing binary.
`volume` | The volume manager used to create, remove, attach, detach, mount, and unmount volumes.


## Examples
This section illustrates several, common examples for using the REX-Ray CLI.

### Print the version
This example shows how to print REX-Ray's version:

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

### Start service
The below example describes how to start REX-Ray as a service:

```shell
$ sudo rexray service start
Starting REX-Ray...SUCCESS!

  The REX-Ray daemon is now running at PID 1455. To
  shutdown the daemon execute the following command:

    sudo /usr/bin/rexray stop
```

### List volumes
This small, but important example highlights how to use REX-Ray to
print a list of volumes:

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

### Create a volume
REX-Ray's volume manager, in addition to listing volumes, can also create
them:

!!! note "note"

    The example below uses the `--size` flag. For a full list of the
    `volume create` command's flags, please use `rexray volume create -?`.

!!! note "note"

    REX-Ray passes these options to
    [libStorage](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/)
    which in turn invokes the API of a storage provider. The settings range,
    and constraints, are generally determined by the storage platform, and
    can have cross-option interactions (e.g. specifying a high IOPS settings,
    reduces the allowed range of size). Cloud providers have been known
    to change constraints over time and across regions. Please consult the
    documentation associated with the storage provider should you
    encounter issues creating a volume.

```shell
$ rexray volume create mysqldata --size 16
ID                     Name       Status     Size
vol-0714d4b348da9e537  mysqldata  available  16
```

### Mount a volume
After a volume has been created, the `rexray volume mount` command
can be used to both attach and mount it to the current host:

```shell
$ rexray volume mount mysqldata
ID                     Name       Status    Size  Path
vol-0762d4b318dade537  mysqldata  attached  16    /var/lib/libstorage/volumes/mysqldata/data
```

### Inspect a volume
The REX-Ray CLI can also inspect and collect additional, detailed
information about a volume:

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

### Unmount a volume
Once a volume is no longer needed, the `rexray volume unmount`
command will unmount the volume:

```shell
$ sudo rexray volume unmount mysqldata
ID                     Name       Status     Size
vol-0764d4b348da9e537  mysqldata  available  16
```

### Remove a volume
If the volume has completely served its purpose and it's time
for its bits to be recycled, use the remove command to completely
delete the volume:

```shell
$ rexray volume rm mysqldata
mysqldata
```
