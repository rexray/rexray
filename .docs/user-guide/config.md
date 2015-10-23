#Configuring REX-Ray

Tweak this, turn that, peek behind the curtain...

---

## Data Directories
The first time `REX-Ray` is executed it will create several directories if
they do not already exist:

* `/etc/rexray`
* `/var/log/rexray`
* `/var/run/rexray`
* `/var/lib/rexray`

The above directories will contain configuration files, logs, PID files, and
mounted volumes. However, the location of these directories can also be
influenced with the environment variable `REXRAY_HOME`.

`REXRAY_HOME` can be used to define a custom home directory for `REX-Ray`.
This directory is irrespective of the actual `REX-Ray` binary. Instead, the
directory specified in `REXRAY_HOME` is the root directory where the `REX-Ray`
binary expects all of the program's data directories to be located.

For example, the following command sets a custom value for `REXRAY_HOME` and
then gets a volume list:

```
env REXRAY_HOME=/tmp/rexray rexray volume
```

The above command would produce a list of volumes and create the following
directories in the process:

* `/tmp/rexray/etc/rexray`
* `/tmp/rexray/var/log/rexray`
* `/tmp/rexray/var/run/rexray`
* `/tmp/rexray/var/lib/rexray`

The entire configuration section will refer to the global configuration file as
a file located inside of `/etc/rexray`, but it should be noted that if
`REXRAY_HOME` is set the location of the global configuration file can be
changed.

## Configuration Methods
There are three ways to configure `REX-Ray`:

* Configuration files
* Environment variables
* Command line options

The order of the items above is also the order of precedence when considering
options set in multiple locations that may override one another.

## Configuration Files
There are two `REX-Ray` configuration files - global and user:

* `/etc/rexray/config.yml`
* `$HOME/.rexray/config.yml`

Please note that while the user configuration file is located inside the user's
home directory, this is the directory of the user that starts `REX-Ray`. And
if `REX-Ray` is being started as a service, then `sudo` is likely being used,
which means that `$HOME/.rexray/config.yml` won't point to *your* home
directory, but rather `/root/.rexray/config.yml`.

This is an example configuration with the default configuration for the general
options as described in the following section:

```yaml
logLevel: warn
osDrivers: linux
volumeDrivers: docker
```

## Log Level
The `REX-Ray` log level determines the level of verbosity emitted by the
internal logger. The default level is `warn`, but there are three other levels
as well:

 Log Level | Description
-----------|-------------
`error`    | Log only errors
`warn`     | Log errors and anything out of place
`info`     | Log errors, warnings, and workflow messages
`debug`    | Log everything

The log level can be set by environment variable, a configuration file, or via
the command line:

Environment Variable | Config File Property | CLI Flag(s)
---------------------|----------------------|-------------
`REXRAY_LOGLEVEL`    | `logLevel`           | `--logLevel`, `-l`

For example, the following two commands may look slightly different, but they
are functionally the same, both printing a list of volumes using the `debug`
log level:

*Use the `debug` log level - Example 1*
```bash
rexray volume get -l debug
```

*Use the `debug` log level - Example 2*
```bash
env REXRAY_LOGLEVEL=debug rexray volume get
```

## OS Drivers
Operating system (OS) drivers enable `REX-Ray` to manage storage on
the underlying OS.

### OS Driver Types
Currently the following OS drivers are supported:

 Driver | Driver Name
--------|------------
Linux   | linux

### Automatic OS Drivers
The OS driver `linux` is automatically activated when `REX-Ray` is running on
the Linux OS.

### Activating OS Drivers
While all drivers are automatically registered, they are not all automatically
activated. OS drivers can be activated by by environment variable, a
configuration file, or via the command line:

Environment Variable | Config File Property | CLI Flag(s)
---------------------|----------------------|-------------
`REXRAY_OSDRIVERS`   | `osDrivers`          | `--osDrivers`

The environment variable and CLI flag both expect a space-delimited list of
OS driver names such as:

```bash
env REXRAY_OSDRIVERS=linux rexray volume get
```

or

```bash
rexray volume get --osDrivers=linux
```

However, when specifying the `osDrivers` option in a configuration file, since
the option is multi-valued, it would look like the following:

```yaml
osDrivers:
- linux
```

## Storage Drivers
Storage drivers enable `REX-Ray` to communicate with direct-attached or remote
storage systems.

### Storage Driver Types
Currently the following storage drivers are supported:

 Driver | Driver Name
--------|------------
Amazon EC2 | ec2
OpenStack | openstack
Rackspace | rackspace
ScaleIO | scaleio
XtremIO | xtremio

### Automatic Storage Drivers
No storage drivers are automatically activated.

### Activating Storage Drivers
While all drivers are automatically registered, not all are automatically
activated. Storage drivers can be activated by by environment variable, a
configuration file, or via the command line:

Environment Variable | Config File Property | CLI Flag(s)
---------------------|----------------------|-------------
`REXRAY_STORAGEDRIVERS`   | `storageDrivers`          | `--storageDrivers`

The environment variable and CLI flag both expect a space-delimited list of
storage driver names such as:

```bash
env REXRAY_STORAGEDRIVERS="ec2 xtremio" rexray volume get
```

or

```bash
rexray volume get --osDrivers="ec2 xtremio"
```

However, when specifying the `storageDrivers` option in a configuration file,
since the option is multi-valued, it would look like the following:

```yaml
storageDrivers:
- ec2
- xtremio
```

## Volume Drivers
Volume drivers enable `REX-Ray` to manage volumes for consumers of the storage,
such as `Docker` or `Mesos`.

### Volume Driver Types
Currently the following volume drivers are supported:

 Driver | Driver Name
--------|------------
Docker   | docker

### Automatic Volume Drivers
The volume driver `docker` is automatically activated.

### Activating Volume Drivers
While all drivers are automatically registered, they are not all automatically
activated. Volume drivers can be activated by by environment variable, a
configuration file, or via the command line:

Environment Variable | Config File Property | CLI Flag(s)
---------------------|----------------------|-------------
`REXRAY_VOLUMEDRIVERS`   | `volumeDrivers`          | `--volumeDrivers`

The environment variable and CLI flag both expect a space-delimited list of
volume driver names such as:

```bash
env REXRAY_VOLUMEDRIVERS=docker rexray volume get
```

or

```bash
rexray volume get --volumeDrivers=docker
```

However, when specifying the `volumeDrivers` option in a configuration file,
since the option is multi-valued, it would look like the following:

```yaml
volumeDrivers:
- docker
```
