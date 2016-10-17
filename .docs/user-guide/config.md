# Configuring REX-Ray

Tweak this, turn that, peek behind the curtain...

---

## Overview
This page reviews how to configure REX-Ray to suit any environment, beginning
with the most common use cases, exploring recommended guidelines, and
finally, delving into the details of more advanced settings.


## Basic Configuration
This section outlines the two most common configuration scenarios encountered
by REX-Ray's users:

 1. REX-Ray as a stand-alone CLI tool
 2. REX-Ray as a service

!!! note "note"

    Please remember to replace the placeholders in the following examples
    with values valid for the systems on which the examples are executed.

    The example below specifies the `volumePath` property as
    `$HOME/VirtualBox/Volumes`. While the text `$HOME` will be replaced with
    the actual value for that environment variable at runtime, the path may
    still be invalid. The `volumePath` property should reflect a path on the
    system on which the VirtualBox server is running, and that is not always
    the same system on which the `libStorage` server is running.

    So please, make sure to update the `volumePath` property for the VirtualBox
    driver to a path valid on the system on which the VirtualBox server is
    running.

    The same goes for VirtualBox property `endpoint` as the VirtualBox
    web service is not always available at `10.0.2.2:18083`.

### Stand-alone CLI Mode
It is possible to use REX-Ray directly from the command line without any
configuration files. The following example uses REX-Ray to list the storage
volumes available to a Linux VM hosted by VirtualBox:

!!! note "note"

    The examples below assume that the VirtualBox web server is running on the
    host OS with authentication disabled and accessible to the guest OS. For
    more information please refer to the VirtualBox storage driver
    [documentation](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/#virtualbox).

```sh
$ rexray volume --service virtualbox ls
ID                                    Name             Status    Size
1b819454-a280-4cff-aff5-141f4e8fd154  libStorage.vmdk  attached  16
```
In addition to listing volumes, the REX-Ray CLI can be used to create and
remove them as well as manage volume snapshots. For an end-to-end example of
volume creation, see [Hello REX-Ray](.././index.md#hello-rex-ray).

#### Embedded Server Mode
When operating as a stand-alone CLI, REX-Ray actually loads an embedded
libStorage server for the duration of the CLI process and is accessible by
only the process that hosts it. This is known as _Embedded Server Mode_.

While commonly used when executing one-off commands with REX-Ray as a
stand-alone CLI tool, Embedded Server Mode can be utilized when configuring
REX-Ray to advertise a static libStorage server as well. The following
qualifications must be met for Embedded Server Mode to be activated:

 * The property `libstorage.host` must not be defined via configuration file,
   environment variable, or CLI flag

 * If the `libstorage.host` property *is* defined then the property
   `libstorage.embedded` can be set to `true` to explicitly activate
   Embedded Server Mode.

 * If the `libstorage.host` property is set and `libtorage.embedded` is
   set to true, Embedded Server Mode will still only activate if the address
   specified by `libstorage.host` (whether a UNIX socket or TCP port) is
   not currently in use.

#### Auto Service Mode
The Stand-alone CLI Mode [example](#stand-alone-cli-mode) also uses the
`--service` flag. This flag's argument sets the `libstorage.service` property,
which has a special meaning inside of REX-Ray -- it serves to enabled
_Auto Service Mode_.

Services represent unique libStorage endpoints that are available to libStorage
clients. Each service is associated with a storage driver. Thus
Auto Service Mode minimizes configuration for simple environments.

The value of the `libstorage.service` property is used to create a default
service configured with a storage driver. This special mode is only activated
if all of the following conditions are met:

  * The `libstorage.service` property is set via:
    * The CLI flags `-s|--service` or `--libstorageService`
    * The environment variable `LIBSTORAGE_SERVICE`
    * The configuration file property `libstorage.service`
  * The `libstorage.host` property is *not* set. This property can be set via:
    * The CLI flags `-h|--host` or `--libstorageHost`
    * The environment variable `LIBSTORAGE_HOST`
    * The configuration file property `libstorage.host`
  * The configuration property `libstorage.server.services` must *not* be set.
    This property is only configurable via a configuration file.

Because the above example met the auto service mode conditions, REX-Ray
created a service named `virtualbox` configured to use the `virtualbox` driver.
This service runs on the libStorage server embedded inside of REX-Ray and is
accessible only by the executing CLI process for the duration of said process.
When used in this manner, the service name must also be a valid driver name.

### Service Mode
REX-Ray can also run as a persistent service that advertises both
[Docker Volume Plug-in](https://docs.docker.com/engine/extend/plugins_volume/)
and [libStorage](http://libstorage.readthedocs.io/en/stable/) endpoints.

#### Docker Volume Plug-in
This section refers to the only operational mode that REX-Ray supported in
versions 0.3.3 and prior. A UNIX socket is created by REX-Ray that serves as a
Docker Volume Plugin compliant API endpoint. Docker is able to leverage this
endpoint to deliver on-demand, persistent storage to containers.

The following is a simple example of a configuration file that should be
located at `/etc/rexray/config.yml`. This file can be used to configure the
same options that were specified in the previous CLI example. Please see the
[advanced section](./config.md#advanced-configuration) for a complete list of
configuration options.

```yaml
libstorage:
  service: virtualbox
virtualbox:
  volumePath: $HOME/VirtualBox/Volumes
```

Once the configuration file is in place, `rexray service start` can be used to
start the service. Sometimes it is also useful to add `-l debug` to enable
more verbose logging. Additionally, it's also occasionally beneficial to
start the service in the foreground with the `-f` flag.

```sh
$ rexray start

Starting REX-Ray...SUCCESS!

  The REX-Ray daemon is now running at PID 15724. To
  shutdown the daemon execute the following command:

    sudo /usr/bin/rexray stop
```

At this point requests can now be made to the default Docker Volume Plugin
and Volume Driver advertised by the UNIX socket `rexray` at
`/run/docker/plugins/rexray.sock`. More details on configuring the Docker
Volume Plug-in are available on the [Schedulers](./schedulers.md) page.

#### libStorage Server and Client
In addition to [Embedded Server Mode](#embedded-server-mode), REX-Ray can also
expose the libStorage API statically. This enables REX-Ray to serve a
libStorage server and perform only a storage abstraction role.

If the desire is to establish a centralized REX-Ray server that is called
on from remote REX-Ray instances then the following example will be useful.
The first configuration is for running REX-Ray purely as a libStorage server.
The second defines how one would would use one or more REX-Ray instances in a
libStorage client role.

The following examples require multiple systems in order to fulfill these
different roles. The [Hello REX-Ray](.././index.md#hello-rex-ray) section on
the front page has an end-to-end illustration of this use case that leverages
Vagrant to provide and configure the necessary systems.

#### libStorage Server
The example below illustrates the necessary settings for configuring REX-Ray
as a libStorage server:

```yaml
rexray:
  modules:
    default-docker:
      disabled: true
libstorage:
  host: tcp://127.0.0.1:7979
  embedded: true
  client:
    type: controller
  server:
    endpoints:
      public:
        address: tcp://:7979
    services:
      virtualbox:
        driver: virtualbox
virtualbox:
  volumePath: $HOME/VirtualBox/Volumes
```

In the above sample, the default Docker module is disabled. This means that
while the REX-Ray service would be running, it would not be available to
Docker on that host.

The `libstorage` section defines the settings that configure the libStorage
server:

Property | Description
---------|------------
`libstorage.host` | Instructs local clients which libStorage endpoint to access
`libstorage.embedded` | Indicates the libStorage server should be started even though the `libstorage.host` property is defined
`libstorage.client.type` | When set to `controller` this property indicates local clients perform no integration activities
`libstorage.server.endpoints` | The available libStorage server HTTP endpoints
`libstorage.server.services` | The configured libStorage services

Start the REX-Ray service with `rexray service start`.

#### libStorage Client
On a separate OS instance running REX-Ray, the follow command can be used to
list the instance's available VirtualBox storage volumes:

```sh
$ rexray volume -h tcp://REXRAY_SERVER:7979 -s virtualbox
```

An alternative to the above CLI flags is to add them as persistent settings
to the `/etc/rexray/config.yml` configuration file on this instance:

```yaml
libstorage:
  host:    tcp://REXRAY_SERVER:7979
  service: virtualbox
```

Now the above command can be simplified further:

```sh
$ rexray volume
```

Once more, the REX-Ray service can be started with `rexray service start` and
the REX-Ray Docker Volume Plug-in endpoint will utilize the remote libStorage
server as its method for communicating with VirtualBox.

Again, a complete end-to-end Vagrant environment for the above example is
available at [Hello REX-Ray](.././index.md#hello-rex-ray).

#### Example sans Modules
Lets review the major sections of the configuration file:

```yaml
rexray:
  logLevel: warn
libstorage:
  service: virtualbox
  integration:
    volume:
      operations:
        create:
          default:
            size: 1
virtualbox:
  volumePath: $HOME/VirtualBox/Volumes
```

Settings occur in three primary areas:

 1. `rexray`
 2. `libstorage`
 3. `virtualbox`

The `rexray` section contains all properties specific to REX-Ray. The
YAML property path `rexray.logLevel` defines the log level for REX-Ray and its
child components. All of the `rexray` properties are
[documented](#configuration-properties) below.

Next, the `libstorage` section defines the service with which REX-Ray will
communicate via the property `libstorage.service`. This property also enables
the [Auto Service Mode](#auto-service-mode) discussed above since this
configuration example does not define a host or services section. For all
information related to libStorage and its properties, please refer to the
[libStorage documentation](http://libstorage.readthedocs.io/).

Finally, the `virtualbox` section configures the VirtualBox driver selected
or loaded by REX-Ray, as indicated via the `libstorage.service` property. The
libStorage Storage Drivers page has information about the configuration details
of [each driver](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers),
including [VirtualBox](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/#virtualbox).

### Logging
The `-l|--logLevel` option or `rexray.logLevel` configuration key can be set
to any of the following values to increase or decrease the verbosity of the
information logged to the console or the REX-Ray log file (defaults to
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
$ rexray env | grep DEFAULT | sort -r
REXRAY_MODULES_DEFAULT-DOCKER_TYPE=docker
REXRAY_MODULES_DEFAULT-DOCKER_SPEC=/etc/docker/plugins/rexray.spec
REXRAY_MODULES_DEFAULT-DOCKER_LIBSTORAGE_SERVICE=vfs
REXRAY_MODULES_DEFAULT-DOCKER_HOST=unix:///run/docker/plugins/rexray.sock
REXRAY_MODULES_DEFAULT-DOCKER_DISABLED=false
REXRAY_MODULES_DEFAULT-DOCKER_DESC=The default docker module.
REXRAY_MODULES_DEFAULT-ADMIN_TYPE=admin
REXRAY_MODULES_DEFAULT-ADMIN_HOST=unix:///var/run/rexray/server.sock
REXRAY_MODULES_DEFAULT-ADMIN_DISABLED=false
REXRAY_MODULES_DEFAULT-ADMIN_DESC=The default admin module.
LIBSTORAGE_INTEGRATION_VOLUME_OPERATIONS_CREATE_DEFAULT_TYPE=
LIBSTORAGE_INTEGRATION_VOLUME_OPERATIONS_CREATE_DEFAULT_SIZE=16
LIBSTORAGE_INTEGRATION_VOLUME_OPERATIONS_CREATE_DEFAULT_IOPS=
LIBSTORAGE_INTEGRATION_VOLUME_OPERATIONS_CREATE_DEFAULT_FSTYPE=ext4
LIBSTORAGE_INTEGRATION_VOLUME_OPERATIONS_CREATE_DEFAULT_AVAILABILITYZONE=
```

## Advanced Configuration
The following sections detail every last aspect of how REX-Ray works and can
be configured.

### Example with Modules
Modules enable a single REX-Ray instance to present multiple personalities or
volume endpoints, serving hosts that require access to multiple storage
platforms.

#### Defining Modules
The following example demonstrates a basic configuration that presents two
modules using the VirtualBox driver: `default-docker` and `vb2-module`.

```yaml
rexray:
  logLevel: warn
  modules:
    default-docker:
      type: docker
      desc: The default docker module.
      host: unix:///run/docker/plugins/vb1.sock
      libstorage:
        service: virtualbox
        integration:
          volume:
            operations:
              create:
                default:
                  size: 1
      virtualbox:
        volumePath: $HOME/VirtualBox/Volumes
    vb2-module:
      type: docker
      desc: The second docker module.
      host: unix:///run/docker/plugins/vb2.sock
      libstorage:
        service: virtualbox
        integration:
          volume:
            operations:
              create:
                default:
                  size: 1
      virtualbox:
        volumePath: $HOME/VirtualBox/Volumes
libstorage:
  service: virtualbox
```

Whereas the previous example did not use modules and the example above does,
they both begin by defining the root section `rexray`. Unlike the previous
example, however, the majority of the `libstorage` section and all of the
`virtualbox` section are no longer at the root. Instead the section
`rexray.modules` is defined. The `modules` key in the `rexray` section is where
all modules are configured. Each key that is a child of `modules` represents the
name of a module.

!!! note "note"

    Please note that while most of the `libstorage` section has been relocated
    as a child of each module, the `libstorage.service` property is still
    defined at the root to activate [Auto Service Mode](#auto-service-mode) as
    a quick-start method of property configuring the embedded libStorage server.

The above example defines two modules:

 1. `default-module`

    This is a special module, and it's always defined, even if not explicitly
    listed. In the previous example without modules, the `libstorage` and
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
when defining both modules, the top-level `libstorage` and `virtualbox` sections
were simply copied into each module as sub-sections. This is perfectly valid
as any configuration path that begins from the root of the REX-Ray
configuration file can be duplicated beginning as a child of a module
definition. This allows global settings to be overridden just for a specific
modules.

As noted, each module shares identical values with the exception of the module's
name and host. The host is the address used by Docker to communicate with
REX-Ray. The base name of the socket file specified in the address can be
used with `docker --volume-driver=`. With the current example the value of the
`--volume-driver` parameter would be either `vb1` of `vb2`.

#### Modules and Inherited Properties
There is also another way to write the previous example while reducing the
number of repeated, identical properties shared by two modules.

```yaml
rexray:
  logLevel: warn
  modules:
    default-docker:
      host: unix:///run/docker/plugins/vb1.sock
      libstorage:
        integration:
          volume:
            operations:
              create:
                default:
                  size: 1
    vb2:
      type: docker
libstorage:
  service: virtualbox
virtualbox:
  volumePath: $HOME/VirtualBox/Volumes
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

Finally, the `libstorage` section has been completely removed from the `vb2`
module whereas it still remains in the `default-docker` section. Volume
creation requests without an explicit size value sent to the `default-docker`
module will result in 1GB volumes whereas the same request sent to the `vb2`
module will result in 16GB volumes (since 16GB is the default value for the
`libstorage.integration.volume.operations.create.default.size` property).

### Defining Service Endpoints
Multiple libStorage services can be defined in order to leverage several
different combinations of storage provider drivers and their respective
configurations. The following section illustrates how to define two separate
services, one using the ScaleIO driver and one using VirtualBox:

```yaml
rexray:
  modules:
    default-docker:
      host:     unix:///run/docker/plugins/virtualbox.sock
      spec:     /etc/docker/plugins/virtualbox.spec
      libstorage:
        service: virtualbox
    scaleio-docker:
      type:     docker
      host:     unix:///run/docker/plugins/scaleio.sock
      spec:     /etc/docker/plugins/scaleio.spec
      libstorage:
        service: scaleio
libstorage:
  server:
    services:
      scaleio:
        driver: scaleio
      virtualbox:
        driver: virtualbox
scaleio:
  endpoint:             https://SCALEIO_GATEWAY/api
  insecure:             true
  userName:             SCALEIO_USER
  password:             SCALEIO_PASS
  systemName:           SCALEIO_SYSTEM_NAME
  protectionDomainName: SCALEIO_DOMAIN_NAME
  storagePoolName:      SCALEIO_STORAG_NAME
virtualbox:
  volumePath: $HOME/VirtualBox/Volumes
```

Once the services have been defined, it is then up to the modules to specify
which service to use. Notice how the `default-docker` module specifies
the `virtualbox` service as its `libstorage.service`. Any requests to the
Docker Volume Plug-in endpoint `/run/docker/plugins/virtualbox.sock` will
utilize the libStorage service `virtualbox` on the backend.

#### Defining a libStorage Server
The following example is very similar to the previous one, but in this instance
there is a centralized REX-Ray server which services requests from many
REX-Ray clients.

```yaml
rexray:
  modules:
    default-docker:
      disabled: true
libstorage:
  host:     tcp://127.0.0.1:7979
  embedded: true
  client:
    type: controller
  server:
    endpoints:
      public:
        address: tcp://:7979
    services:
      scaleio:
        driver: scaleio
      virtualbox:
        driver: virtualbox
scaleio:
  endpoint:             https://SCALEIO_GATEWAY/api
  insecure:             true
  userName:             SCALEIO_USER
  password:             SCALEIO_PASS
  systemName:           SCALEIO_SYSTEM_NAME
  protectionDomainName: SCALEIO_DOMAIN_NAME
  storagePoolName:      SCALEIO_STORAG_NAME
virtualbox:
  volumePath: $HOME/VirtualBox/Volumes
```

One of the larger differences between the above example and the previous one is
the removal of the module definitions. Docker does not communicate with the
central REX-Ray server directly; instead Docker interacts with the REX-Ray
services running on the clients via their Docker Volume Endpoints. The client
REX-Ray instances then send all storage-related requests to the central REX-Ray
server.

Additionally, the above sample configuration introduces a few new properties:

Property | Description
---------|------------
`libstorage.host` | Instructs local clients which libStorage endpoint to access
`libstorage.embedded` | Indicates the libStorage server should be started even though the `libstorage.host` property is defined
`libstorage.client.type` | When set to `controller` this property indicates local clients perform no integration activities
`libstorage.server.endpoints` | The available libStorage server HTTP endpoints

#### Defining a libStorage Client
The client configuration is still rather simple. As mentioned in the previous
section, the `rexray.modules` configuration occurs here. This enables the Docker
engines running on remote instances to communicate with local REX-Ray exposed
Docker Volume endpoints that then handle the storage-related requests via the
centralized REX-Ray server.

```yaml
rexray:
  modules:
    default-docker:
      host:     unix:///run/docker/plugins/virtualbox.sock
      spec:     /etc/docker/plugins/virtualbox.spec
      libstorage:
        service: virtualbox
    scaleio-docker:
      type:     docker
      host:     unix:///run/docker/plugins/scaleio.sock
      spec:     /etc/docker/plugins/scaleio.spec
      libstorage:
        service: scaleio
libstorage:
  host: tcp://REXRAY_SERVER:7979
```

### libStorage Configuration
REX-Ray embeds both the libStorage client as well as the libStorage server. For
information on configuring the following, please refer to the
[libStorage documentation](http://libstorage.readthedocs.io/en/stable):

 - [Volume options](http://libstorage.readthedocs.io/en/stable/user-guide/config/#volume-configuration)
   such as preemption, disabling operations, etc.
 - Fine-tuning [logging](http://libstorage.readthedocs.io/en/stable/user-guide/config/#logging-configuration)
 - [Configuring](http://libstorage.readthedocs.io/en/stable/user-guide/config/#driver-configuration)
   OS, integration, and storage drivers

### Data Directories
The first time REX-Ray is executed it will create several directories if
they do not already exist:

* `/etc/rexray`
* `/var/log/rexray`
* `/var/run/rexray`
* `/var/lib/rexray`

The above directories will contain configuration files, logs, PID files, and
mounted volumes. However, the location of these directories can also be
influenced with the environment variable `REXRAY_HOME`.

`REXRAY_HOME` can be used to define a custom home directory for REX-Ray.
This directory is irrespective of the actual REX-Ray binary. Instead, the
directory specified in `REXRAY_HOME` is the root directory where the REX-Ray
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
There are three ways to configure REX-Ray:

* Command line options
* Environment variables
* Configuration files

The order of the items above is also the order of precedence when considering
options set in multiple locations that may override one another. Values set
via CLI flags have the highest order of precedence, followed by values set by
environment variables, followed, finally, by values set in configuration files.

### Configuration Files
There are two REX-Ray configuration files - global and user:

* `/etc/rexray/config.yml`
* `$HOME/.rexray/config.yml`

Please note that while the user configuration file is located inside the user's
home directory, this is the directory of the user that starts REX-Ray. And
if REX-Ray is being started as a service, then `sudo` is likely being used,
which means that `$HOME/.rexray/config.yml` won't point to *your* home
directory, but rather `/root/.rexray/config.yml`.

The next section has an example configuration with the default configuration.

### Configuration Properties
The section [Configuration Methods](#configuration-methods) mentions there are
three ways to configure REX-Ray: config files, environment variables, and the
command line. However, this section will illuminate the relationship between the
names of the configuration file properties, environment variables, and CLI
flags.

Here is a sample REX-Ray configuration:

```yaml
rexray:
  logLevel: warn
libstorage:
  service: virtualbox
virtualbox:
  volumePath: $HOME/VirtualBox/Volumes
```

The properties `rexray.logLevel`, `libstorage.service`, and
`virtualbox.volumePath` are strings. These values can also be set via
environment variables or command line interface (CLI) flags, but to do so
requires knowing the names of the environment variables or CLI flags to use.
Luckily those are very easy to figure out just by knowing the property names.

All properties that might appear in the REX-Ray configuration file
fall under some type of heading. For example, take the configuration above:

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
`libstorage.service`   | `LIBSTORAGE_SERVICE`   | `--libstorageService`
`virtualbox.volumePath`    | `VIRTUALBOX_VOLUMEPATH`   | `--virtualboxVolumePath`

### Logging Configuration
The REX-Ray log level determines the level of verbosity emitted by the
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
rexray volume -l debug ls
```

*Use the `debug` log level - Example 2*
```bash
env REXRAY_LOGLEVEL=debug rexray volume ls
```
