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

The next section has an example configuration with the default configuration.

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

## Logging Configuration
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

## Driver Configuration
There are three types of drivers:

  1. OS Drivers
  2. Storage Drivers
  3. Volume Drivers

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

## Module Configuration
This section reviews exposing multiple, differently configured endpoints by
using modules.

### Default Modules
If not explicitly specified in a configuration source, `REX-Ray` always
considers the following, default modules:

```yaml
rexray:
    modules:
        default-admin:
            type:     admin
            desc:     The default admin module.
            host:     tcp://127.0.0.1:7979
            disabled: false
        default-docker:
            type:     docker
            desc:     The default docker module.
            host:     unix:///run/docker/plugins/rexray.sock
            spec:     /etc/docker/plugins/rexray.spec
            disabled: false
```

The first module, `default-admin`, is used by the CLI to communicate with the
REX-Ray service API. For security reasons the `default-admin` module is bound
to the loopback IP address by default.

The second default module, `default-docker`, exposes `REX-Ray` as a Docker
Volume Plug-in via the specified sock and spec files.

### Additional Modules
It's also possible to create additional modules via the configuration file:

```yaml
rexray:
    storageDrivers:
    - isilon
    modules:
        isilon2:
            type: docker
            desc: A second docker module.
            host: unix:///run/docker/plugins/isilon2.sock
            spec: /etc/docker/plugins/isilon2.spec
isilon:
    endpoint:   https://172.17.177.230:8080
    insecure:   true
    username:   root
    password:   P@ssword1!
    volumePath: /rexray/default
    nfsHost:    172.17.177.230
    quotas:     true
    dataSubnet: 172.17.177.0/24
```

The above example defines three modules:

1. `default-admin`
2. `default-docker`
3. `isilon2`

The first two, default modules are not included in the configuration file as
they are implicit. The only reason to define them explicitly is to override
their properties, a feature discussed in the next section.

Ignoring the `default-admin` module, the `default-docker` and `isilon2` modules
are both Docker modules as indicated by a module's `type` property. Just like
the `default-docker` module, the custom module `isilon2` is configured to use
the default isilon settings from the root key, `isilon`. Therefore the modules
`default-docker` and `isilon2` are configured exactly the same except for they
are exposed via different sock and spec files.

### Inferred Properties
The following example is nearly identical to the previous one except this
example is missing the `host`, `spec`, `desc`, and `disabled` properties for
the `isilon2` module:

```yaml
rexray:
    storageDrivers:
    - isilon
    modules:
        isilon2:
            type: docker
            desc: A second docker module.
isilon:
    endpoint:   https://172.17.177.230:8080
    insecure:   true
    username:   root
    password:   P@ssword1!
    volumePath: /rexray/default
    nfsHost:    172.17.177.230
    quotas:     true
    dataSubnet: 172.17.177.0/24
```

A module is not required to provide a description (`desc`), and `disabled` will
always default to `false` unless explicitly specified as `true`. Docker modules
(`type: docker`) will also infer the values of the `host` and `spec`
properties if they are not explicitly provided. Because the name of the
module above is `isilon2` and the `host` and `spec` properties are not defined,
those values will be automatically set to
`unix:///run/docker/plugins/isilon2.sock` and `/etc/docker/plugins/isilon2.spec`
respectively.

### Overriding Defaults
It is also possible to override a default module's configuration. What if it's
determined the `default-admin` module should be accessible externally and the
`default-docker` module should use a different sock file? Simply override those
keys and only those keys:

```yaml
rexray:
    storageDrivers:
    - isilon
    modules:
        default-admin:
            host: tcp://:7979
        default-docker:
            host: unix:///run/docker/plugins/isilon1.sock
isilon:
    endpoint:   https://172.17.177.230:8080
    insecure:   true
    username:   root
    password:   P@ssword1!
    volumePath: /rexray/default
    nfsHost:    172.17.177.230
    quotas:     true
    dataSubnet: 172.17.177.0/24
```

### Overriding Inherited Properties
In all of the module configuration examples so far there has been a root key
named `isilon` that provides the settings for the storage driver used by the
modules. Thanks to scoped configuration support and inherited properties, it's
also quite simple to provide adjustments to a default configuration at the
module level. For example, imagine the `isilon2` module should load a driver
that points to a different `volumePath`?

```yaml
rexray:
    storageDrivers:
    - isilon
    modules:
        isilon2:
            type:       docker
            isilon:
                volumePath: /rexray/isilon2
isilon:
    endpoint:   https://172.17.177.230:8080
    insecure:   true
    username:   root
    password:   P@ssword1!
    volumePath: /rexray/default
    nfsHost:    172.17.177.230
    quotas:     true
    dataSubnet: 172.17.177.0/24
```

The above example will load two Docker modules, the `default-docker` module and
the `isilon2` module. The `default-docker` module's `isilon.volumePath` will be
set to `/rexray/default` whereas the `isilon2` module's `isilon.volumePath` is
overridden and set to `/rexray/isilon2`.

Any key path structure can be duplicated under the module's name, and the value
at the terminus of that path will be used in place of any inherited value.
Another example is overriding the type of storage driver used by the `isilon2`
module. There may be a case where the `isilon2` module needs to use an
enhanced version of the `isilon` storage driver but still use the same
configuration:

```yaml
rexray:
    storageDrivers:
    - isilon
    modules:
        isilon2:
            type:       docker
            rexray:
                storageDrivers:
                - isilonEnhanced
isilon:
    endpoint:   https://172.17.177.230:8080
    insecure:   true
    username:   root
    password:   P@ssword1!
    volumePath: /rexray/default
    nfsHost:    172.17.177.230
    quotas:     true
    dataSubnet: 172.17.177.0/24
```

The above example recreates the key path structure `rexray.storageDrivers`
beneath the key path structure `rexray.modules.isilon2`. Whenever any query is
made for `rexray.storageDrivers` inside the `isilon2` module, the value
`[]string{"isilonEnhanced"}` is returned instead of `[]string{"isilon"}`.

### Disabling Modules
Both default and custom modules can be disabled by setting the key `disabled` to
true inside a module definition:

```yaml
rexray:
    storageDrivers:
    - isilon
    modules:
        default-docker:
            disabled: true
        isilon2:
            type:     docker
        isilon3:
            type:     docker
            disabled: true
isilon:
    endpoint:   https://172.17.177.230:8080
    insecure:   true
    username:   root
    password:   P@ssword1!
    volumePath: /rexray/default
    nfsHost:    172.17.177.230
    quotas:     true
    dataSubnet: 172.17.177.0/24
```

The above example disables the `default-docker` and `isilon3` modules such that
`isilon2` is the only Docker module loaded.

## Volume Configuration
This section describes various global configuration options related to
operations such as mounting and unmounting volumes.

### Pre-Emption
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
EC2|Yes, no Ubuntu support
Isilon|Not yet
OpenStack|With Cinder v2
ScaleIO|Yes
Rackspace|No
VirtualBox|Yes
VMAX|Not yet
XtremIO|Yes

### Ignore Used Count
By default accounting takes place during operations that are performed
on `Mount`, `Unmount`, and other operations.  This only has impact when running
as a service through the HTTP/JSON interface since the counts are persisted
in memory.  The purpose of respecting the `Used Count` is to ensure that a
volume is not unmounted until the unmount requests have equaled the mount
requests.  

In the `Docker` use case if there are multiple containers sharing a volume
on the same host, the the volume will not be unmounted until the last container
is stopped.  

The following setting should only be used if you wish to *disable* this
functionality.  This would make sense if the accounting is being done from
higher layers and all unmount operations should proceed without control.
```yaml
rexray:
  volume:
    unmount:
      ignoreUsedCount: true
```

Currently a reset of the service will cause the counts to be reset.  This
will cause issues if *multiple containers* are sharing a volume.  If you are
sharing volumes, it is recommended that you reset the service along with the
accompanying container runtime (if this setting is false) to ensure they are
synchronized.  

### Volume Path (0.3.1)
When volumes are mounted there can be an additional path that is specified to
be created and passed as the valid mount point.  This is required for certain
applications that do not want to place data from the root of a mount point.  
The default is the `/data` path.  If a value is set by `linux.volume.rootPath`,
then the default will be overwritten.

If upgrading to 0.3.1 then you can either set this to an empty value, or move
the internal directory in your existing volumes to `/data`.

```yaml
rexray:
linux:
  volume:
    rootPath: /data
```

### Volume FileMode (0.3.1)
The permissions of the `linux.volume.rootPath` can be set to default values.  At
each mount, the permissions will be written based on this value.  The default
is to include the `0700` mode.

```yaml
rexray:
linux:
  volume:
    fileMode: 0700
```
