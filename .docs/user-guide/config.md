# Configuring REX-Ray

Tweak this, turn that, peek behind the curtain...

---

## Overview
This page reviews how to configure `REX-Ray` to suit any environment, beginning
with the the most common use cases, exploring recommended guidelines, and
finally, delving into the details of more advanced settings.

## Basic Configuration
This section outlines the most common configuration scenarios encountered by
`REX-Ray`'s users.

A typical configuration occurs by creating the file `/etc/rexray/config.yml`.
This is the global settings file for `REX-Ray`, and is used whether the
program is used as a command line interface (CLI) application or started as a
background service. Please see the [advanced section](./config.md#advanced-configuration)
for more information and options regarding configuration.

### Example sans Modules
The following example is a YAML configuration for the VirtualBox driver.

```
rexray:
  logLevel: warn
  storageDrivers:
  - virtualbox
docker:
  size: 1
virtualbox:
  endpoint:       http://10.0.2.2:18083
  tls:            false
  volumePath:     $HOME/Repos/vagrant/rexray/Volumes
  controllerName: SATA
```

Settings occur in three primary areas:

 1. `rexray`
 2. `docker`
 3. `virtualbox`

The `rexray` section contains all properties specific to `REX-Ray`. The
YAML property path `rexray.storageDrivers` lists the names of the storage
drivers loaded by `REX-Ray`. In this case, only the `virtualbox` storage
driver is loaded. All of the `rexray` properties are
[documented](#configuration-properties) below.

The `docker` section defines properties specific to Docker. The property
`docker.size` defines in gigabytes the default size for a new Docker volume.
The complete list of properties for the `docker` section are described on the
[Schedulers page](./schedulers.md#docker).

Finally, the `virtualbox` section configures the VirtualBox driver loaded by
`REX-Ray`, as indicated via the `rexray.storageDrivers` property). The
Storage Drivers page has information about the configuration details of
[each driver](./storage-providers.md), including
[VirtualBox](./storage-providers.md#virtualbox).

### Example with Modules
Modules enable a single `REX-Ray` instance to present multiple personalities or
volume endpoints, serving hosts that require access to multiple storage
platforms.

#### Defining Modules
The following example demonstrates a basic configuration that presents two
modules using the VirtualBox driver.

```
rexray:
  logLevel: warn
  storageDrivers:
  - virtualbox
  modules:
    default-docker:
      type: docker
      desc: "The default docker module."
      host: "unix:///run/docker/plugins/vb1.sock"
      docker:
        size: 1
      virtualbox:
        endpoint:      http://10.0.2.2:18083
        tls:           false
        volumePath:    "$HOME/Repos/vagrant/rexray/Volumes"
        controllerName: SATA
    vb2-module:
      type: docker
      desc: "The second docker module."
      host: "unix:///run/docker/plugins/vb2.sock"
      docker:
        size: 1
      virtualbox:
        endpoint:      http://10.0.2.2:18083
        tls:           false
        volumePath:    "$HOME/Repos/vagrant/rexray/Volumes"
        controllerName: SATA
```

Like the previous example that did not use modules, this example begins by
defining the root section `rexray`. Unlike the previous example, the `docker`
and  `virtualbox` sections are no longer at the root. Instead the section
`rexray.modules` is defined. The `modules` key in the `rexray` section is
where all modules are configured. Each key that is a child of `modules`
represents the name of a module.

The above example defines two modules:

 1. `default-module`

    This is a special module, and it's always defined, even if not explicitly
    listed. In the previous example without modules, the `docker` and
    `virtualbox` sections at the root actually informed the configuration of
    the implicit `default-docker` module. In this example the explicit
    declaration of the `default-docker` module enables several of its
    properties to be overridden and given desired values. The Advanced
    Configuration section has more information on
    [Default Modules](#default-modules).

 2. `vb2-module`

    This is a new, custom module configured almost identically to the
    `default-module` with the exception of a unique host address as defined
    by the module's `host` key.

Notice that both modules share many of the same properties and values. In fact,
when defining both modules, the top-level `docker` and `virtualbox` sections
were simply copied into each module as sub-sections. This is perfectly valid
as any configuration path that begins from the root of the `REX-Ray`
configuration file can be duplicated beginning as a child of a module
definition. This allows global settings to be overridden just for a specific
modules.

As noted, each module shares identical values with the exception of the module's
name and host. The host is the address used by Docker to communicate with
`REX-Ray`. The base name of the socket file specified in the address can be
used with `docker --volume-driver=`. With the current example the value of the
`--volume-driver` parameter would be either `vb1` of `vb2`.

#### Modules and Inherited Properties
There is also another way to write the previous example while reducing the
number of repeated, identical properties shared by two modules.

```
rexray:
  logLevel: warn
  storageDrivers:
  - virtualbox
  modules:
    default-docker:
      host: "unix:///run/docker/plugins/vb1.sock"
      docker:
        size: 1
    vb2:
      type: docker
virtualbox:
  endpoint:      http://10.0.2.2:18083
  tls:           false
  volumePath:    "$HOME/Repos/vagrant/rexray/Volumes"
  controllerName: SATA
```

The above example may look strikingly different than the previous one, but it's
actually the same with just a few tweaks.

While there are still two modules defined, the second one has been renamed from
`vb2-module` to `vb2`. The change is a more succinct way to represent the same
intent, and it also provides a nice side-effect. If the `host` key is omitted
from a Docker module, the value for the `host` key is automatically generated
using the module's name. Therefore since there is no `host` key for the `vb2`
module, the value will be `unix:///run/docker/plugins/vb2.sock`.

Additionally, `virtualbox` sections from each module definition have been
removed and now only a single, global `virtualbox` section is present at the
root. When accessing properties, a module will first attempt to access a
property defined in the context of the module, but if that fails the property
lookup will resolve against globally defined keys as well.

Finally, please also note that the `docker` section has **not** been promoted
back to a global property set, and is in fact still located in the context of
the `default-docker` module. This means that create volume requests sent to the
`default-docker` module will result in 1GB volumes by default whereas create
volume requests handled by the `vb2` module will result in 16GB volumes (since
16GB is the default value for the `docker.size` property).

### Logging
The `-l|--logLevel` option or `rexray.logLevel` configuration key can be set
to any of the following values to increase or decrease the verbosity of the
information logged to the console or the `REX-Ray` log file (defaults to
`/var/log/rexray/rexray.log`).

- panic
- fatal
- error
- warn
- info
- debug

### Troubleshooting
The command `rexray env` can be used to print out the runtime interpretation
of the environment, including configured properties, in order to help diagnose
configuration issues.

```
$ rexray env | grep DEFAULT-DOCKER
REXRAY_MODULES_DEFAULT-DOCKER_TYPE=docker
REXRAY_MODULES_DEFAULT-DOCKER_DISABLED=false
REXRAY_MODULES_DEFAULT-DOCKER_VIRTUALBOX_CONTROLLERNAME=SATA
REXRAY_MODULES_DEFAULT-DOCKER_DESC=The default docker module.
REXRAY_MODULES_DEFAULT-DOCKER_HOST=unix:///run/docker/plugins/vb1.sock
REXRAY_MODULES_DEFAULT-DOCKER_VIRTUALBOX_ENDPOINT=http://10.0.2.2:18083
REXRAY_MODULES_DEFAULT-DOCKER_VIRTUALBOX_VOLUMEPATH=$HOME/Repos/vagrant/rexray/Volumes
REXRAY_MODULES_DEFAULT-DOCKER_DOCKER_SIZE=1
REXRAY_MODULES_DEFAULT-DOCKER_SPEC=/etc/docker/plugins/rexray.spec
REXRAY_MODULES_DEFAULT-DOCKER_VIRTUALBOX_TLS=false
```

## Advanced Configuration
The following sections detail every last aspect of how `REX-Ray` works and can
be configured.

### Data Directories
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

### Configuration Methods
There are three ways to configure `REX-Ray`:

* Command line options
* Environment variables
* Configuration files

The order of the items above is also the order of precedence when considering
options set in multiple locations that may override one another. Values set
via CLI flags have the highest order of precedence, followed by values set by
environment variables, followed, finally, by values set in configuration files.

### Configuration Files
There are two `REX-Ray` configuration files - global and user:

* `/etc/rexray/config.yml`
* `$HOME/.rexray/config.yml`

Please note that while the user configuration file is located inside the user's
home directory, this is the directory of the user that starts `REX-Ray`. And
if `REX-Ray` is being started as a service, then `sudo` is likely being used,
which means that `$HOME/.rexray/config.yml` won't point to *your* home
directory, but rather `/root/.rexray/config.yml`.

The next section has an example configuration with the default configuration.

### Configuration Properties
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

#### String Arrays
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

### Logging Configuration
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

### Driver Configuration
There are three types of drivers:

  1. OS Drivers
  2. Storage Drivers
  3. Volume Drivers

#### OS Drivers
Operating system (OS) drivers enable `REX-Ray` to manage storage on
the underlying OS. Currently the following OS drivers are supported:

 Driver | Driver Name
--------|------------
Linux   | linux

The OS driver `linux` is automatically activated when `REX-Ray` is running on
the Linux OS.

#### Storage Drivers
Storage drivers enable `REX-Ray` to communicate with direct-attached or remote
storage systems. Currently the following storage drivers are supported:

 Driver | Driver Name
--------|------------
[Amazon EC2](./storage-providers.md#amazon-ec2) | ec2
[Google Compute Engine](./storage-providers.md#google-compute-engine) | gce
[Isilon](./storage-providers.md#isilon) | isilon
[OpenStack](./storage-providers.md#openstack) | openstack
[Rackspace](./storage-providers.md#rackspace) | rackspace
[ScaleIO](./storage-providers.md#scaleio) | scaleio
[VirtualBox](./storage-providers.md#virtualbox) | virtualbox
[VMAX](./storage-providers.md#vmax) | vmax
[XtremIO](./storage-providers.md#xtremio) | xtremio

The `rexray.storageDrivers` property can be used to activate storage drivers..

#### Volume Drivers
Volume drivers enable `REX-Ray` to manage volumes for consumers of the storage,
such as `Docker` or `Mesos`. Currently the following volume drivers are
supported:

 Driver | Driver Name
--------|------------
Docker   | docker

The volume driver `docker` is automatically activated.

### Module Configuration
This section reviews exposing multiple, differently configured endpoints by
using modules.

#### Default Modules
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

#### Additional Modules
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

#### Inferred Properties
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

#### Overriding Defaults
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

#### Overriding Inherited Properties
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

#### Disabling Modules
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

### Volume Configuration
This section describes various global configuration options related to
operations such as mounting and unmounting volumes.

#### Disable Create
The disable create feature enables you to disallow any volume creation activity.
Any requests will be returned in a successful manner, but the create will not
get passed to the backend storage platform.

```yaml
rexray:
  volume:
    create:
      disable: true
```

#### Disable Remove
The disable remove feature enables you to disallow any volume removal activity.
Any requests will be returned in a successful manner, but the remove will not
get passed to the backend storage platform.

```yaml
rexray:
  volume:
    remove:
      disable: true
```

#### Preemption
There is a capability to preemptively detach any existing attachments to other
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

#### Ignore Used Count
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

#### Volume Path Disable Cache (0.3.2)
In order to minimize the impact to return `Path` requests, a caching
capability has been introduced by default. A `List` request will cause the
returned volumes and paths to be evaluated and those with active mounts are
recorded. Subsequent `Path` requests for volumes that have no recorded mounts
will not result in active path lookups. Once the mount counter is initialized or
a `List` operation occurs where a mount is recorded, the volume will be looked
up for future `Path` operations.

```yaml
rexray:
  volume:
    path:
      disableCache: true
```

#### Volume Root Path (0.3.1)
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

#### Volume FileMode (0.3.1)
The permissions of the `linux.volume.rootPath` can be set to default values.  At
each mount, the permissions will be written based on this value.  The default
is to include the `0700` mode.

```yaml
rexray:
linux:
  volume:
    fileMode: 0700
```
