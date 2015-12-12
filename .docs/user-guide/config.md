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

* Command line options
* Environment variables
* Configuration files

The order of the items above is also the order of precedence when considering
options set in multiple locations that may override one another. Values set
via CLI flags have the highest order of precedence, followed by values set by
environment variables, followed, finally, by values set in configuration files.

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
rexray:
    logLevel: warn
    osDrivers:
    - linux
    volumeDrivers:
    - docker
```

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
rexray:
    logLevel: warn
    osDrivers:
    - linux
    storageDrivers:
    volumeDrivers:
    - docker
```

## Configuration Properties
The section [Configuration Methods](#configuration-methods) mentions there are
three ways to configure REX-Ray: config files, environment variables, and the
command line. However, this section will illuminate the relationship between the
names of the configuration file properties, environment variables, and CLI
flags.

These are the global configuration properties with their default values as
represented in a a YAML configuration file:

```yaml
rexray:
    logLevel: warn
    osDrivers:
    - linux
    storageDrivers:
    volumeDrivers:
    - docker
```

The property `rexray.logLevel` is a string and the properties
`rexray.osDrivers`, `rexray.storageDrivers`, and `rexray.volumeDrivers` are all
arrays of strings. These values can also be set via environment variables or the
command line, but to do so requires knowing the names of the environment
variables or CLI flags to use. Luckily those are very easy to figure out just
by knowing the property names.

All properties that might appear in the `REX-Ray` configuration file
fall under some type of heading. For example, take the default configuration
above.

The rule for environment variables is as follows:

  * Each nested level becomes a part of the environment variable name followed
    by an underscore `_` except for the terminating part.
  * The entire environment variable name is uppercase.

Nested properties follow these rules for CLI flags:

  * The root level's first character is lower-cased with the rest of the root
    level's text left unaltered.
  * The remaining levels' first characters are all upper-cased with the the
    remaining text of that level left unaltered.
  * All levels are then concatenated together.
  * See the verbose help for exact global flags using `rexray --help -v`
    as they may be chopped to minimize verbosity.

The following table illustrates the transformations:

Property Name | Environment Variable | CLI Flag
--------------|----------------------|-------------
`rexray.logLevel`    | `REXRAY_LOGLEVEL`    | `--logLevel`
`rexray.osDrivers`   | `REXRAY_OSDRIVERS`   | `--osDrivers`
`rexray.storageDrivers`    | `REXRAY_STORAGEDRIVERS`   | `--storageDrivers`
`rexray.volumeDrivers`    | `REXRAY_VOLUMEDRIVERS`   | `--volumeDrivers`

Another example is a possible configuration of the Amazon Web Services (AWS)
Elastic Compute Cloud (EC2) storage driver:

```yaml
aws:
    accessKey: MyAccessKey
    secretKey: MySecretKey
    region:    USNW
```

Property Name | Environment Variable | CLI Flag
--------------|----------------------|-------------
`aws.accessKey`    | `AWS_ACCESSKEY`    | `--awsAccessKey`
`aws.secretKey`   | `AWS_SECRETKEY`   | `--awsSecretKey`
`aws.region`    | `AWS_REGION`   | `--awsRegion`

### String Arrays
Please note that properties that are represented as arrays in a configuration
file, such as the `rexray.osDrivers`, `rexray.storageDrivers`, and
`rexray.volumeDrivers` above, are not arrays, but multi-valued strings where a
space acts as a delimiter. This is because the Viper project
[does not bind](https://github.com/spf13/viper/issues/112) Go StringSlices
(string arrays) correctly to [PFlags](https://github.com/spf13/pflag).

For example, this is how one would specify the storage drivers `ec2` and
`xtremio` in a configuration file:

```yaml
rexray:
    storageDrivers:
    - ec2
    - xtremio
```

However, to specify the same values in an environment variable,
`REXRAY_STORAGEDRIVERS="ec2 xtremio"`, and as a CLI flag,
`--storageDrivers="ec2 xtremio"`.

## Logging
The `REX-Ray` log level determines the level of verbosity emitted by the
internal logger. The default level is `warn`, but there are three other levels
as well:

 Log Level | Description
-----------|-------------
`error`    | Log only errors
`warn`     | Log errors and anything out of place
`info`     | Log errors, warnings, and workflow messages
`debug`    | Log everything

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

### OS Drivers
Operating system (OS) drivers enable `REX-Ray` to manage storage on
the underlying OS. Currently the following OS drivers are supported:

 Driver | Driver Name
--------|------------
Linux   | linux

The OS driver `linux` is automatically activated when `REX-Ray` is running on
the Linux OS.

### Storage Drivers
Storage drivers enable `REX-Ray` to communicate with direct-attached or remote
storage systems. Currently the following storage drivers are supported:

 Driver | Driver Name
--------|------------
Amazon EC2 | ec2
OpenStack | openstack
Rackspace | rackspace
ScaleIO | scaleio
XtremIO | xtremio

The `rexray.storageDrivers` property can be used to activate storage drivers..

### Volume Drivers
Volume drivers enable `REX-Ray` to manage volumes for consumers of the storage,
such as `Docker` or `Mesos`. Currently the following volume drivers are
supported:

 Driver | Driver Name
--------|------------
Docker   | docker

The volume driver `docker` is automatically activated.

### Pre-emptive Volume Mount
There is a capability to pre-emptively detach any existing attachments to other
instances before attempting a mount.  This will enable use cases for
availability where another instance must be able to take control of a volume
without the current owner instance being involved.  The operation is considered
equivalent to a power off of the existing instance for the device.

Example configuration file follows:
```yaml
rexray:
  storageDrivers:
  - openstack
  volume:
    mount:
      preempt: true
openStack:
  authUrl: https://authUrl:35357/v2.0/
  username: username
  password: password
  tenantName: tenantName
  regionName: regionName
```

Driver|Supported
------|---------
EC2|Yes, HVM only
OpenStack|With Cinder v2
ScaleIO|Yes
Rackspace|No
XtremIO|Yes
