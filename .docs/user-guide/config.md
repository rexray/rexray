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
logLevel: warn
osDrivers:
- linux
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
logLevel: warn
osDrivers:
- linux
storageDrivers:
volumeDrivers:
- docker
```

The property `logLevel` is a string and the properties `osDrivers`,
`storageDrivers`, and `volumeDrivers` are all arrays of strings. These values
can also be set via environment variables or the command line, but to do so
requires knowing the names of the environment variables or CLI flags to use.
Luckily those are very easy to figure out just by knowing the property names.

### Top-Level Properties
For any configuration file property there is really only one exception to
consider when deriving the name of the property's equivalent environment
variable or CLI flag: is the property a top-level property? The top-level
properties are:

  * `logLevel`
  * `osDrivers`
  * `storageDrivers`
  * `volumeDrivers`

For any top-level property the name of its equivalent environment variable is
the name of the property, in all caps, prefixed with `REXRAY_`. Additionally,
the name of the property's CLI flag is equivalent to the property name, as is.
The following table illustrates the transformations:

Property Name | Environment Variable | CLI Flag
--------------|----------------------|-------------
`logLevel`    | `REXRAY_LOGLEVEL`    | `--logLevel`
`osDrivers`   | `REXRAY_OSDRIVERS`   | `--osDrivers`
`storageDrivers`    | `REXRAY_STORAGEDRIVERS`   | `--storageDrivers`
`volumeDrivers`     | `REXRAY_VOLUMEDRIVERS`    | `--volumeDrivers`

### All Other Properties
The other properties that might appear in the `REX-Ray` configuration file all
fall under some type of heading. For example, a possible configuration of the
Amazon Web Services (AWS) Elastic Compute Cloud (EC2) storage driver might look
like this:

```yaml
aws:
    accessKey: MyAccessKey
    secretKey: MySecretKey
    region:    USNW
```

For any property that's nested, the rule for environment variables is as
follows:

  * Each nested level becomes a part of the environment variable name followed
    by an underscore `_` except for the terminating part.
  * The entire environment variable name is uppercase.

Nested properties follow these rules for CLI flags:

  * The root level's first character is lower-cased with the rest of the root
    level's text left unaltered.
  * The remaining levels' first characters are all upper-cased with the the
    remaining text of that level left unaltered.
  * All levels are then concatenated together.

The following table illustrates the transformations:

Property Name | Environment Variable | CLI Flag
--------------|----------------------|-------------
`aws.accessKey`    | `AWS_ACCESSKEY`    | `--awsAccessKey`
`aws.secretKey`   | `AWS_SECRETKEY`   | `--awsSecretKey`
`aws.region`    | `AWS_REGION`   | `--awsRegion`

### String Arrays
Please note that properties that are represented as arrays in a configuration
file, such as the `osDrivers`, `storageDrivers`, and `volumeDrivers` above, are
not arrays, but multi-valued strings where a space acts as a delimiter. This is
because the Viper project
[does not bind](https://github.com/spf13/viper/issues/112) Go StringSlices
(string arrays) correctly to [PFlags](https://github.com/spf13/pflag).

For example, this is how one would specify the storage drivers `ec2` and
`xtremio` in a configuration file:

```yaml
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
