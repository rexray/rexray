# Configuring libStorage

Tweak this, turn that, peek behind the curtain...

---

## Overview
This page reviews how to configure `libStorage` to suit any environment,
beginning with the the most common use cases, exploring recommended guidelines,
and finally, delving into the details of more advanced settings.

### Client/Server Configuration
Except when specified otherwise, the configuration examples below assume the
`libStorage` client and server exist on the same host. However, that is not at
all a requirement. It is fully possible, and in fact the entire purpose of
`libStorage`, that the client and server be able to function on different
systems. One `libStorage` server should be able to support hundreds of clients.
Yet for the sake of completeness, the examples below show both configurations
merged.

When configuring a `libStorage` client and server for different systems, there
will be a few differences from the examples below:

  * The examples show `libStorage` configured with its server component hosted
    on a UNIX socket. This is ideal for when the client/server exist on the same
    host as it reduces security risks. However, in most real-world scenarios
    the client and server are *not* residing on the same host, the
    `libStorage`  server should use a TCP endpoint so it can be accessed
    remotely.

  * In a distributed configuration the actual driver configuration sections
    need only occur on the server-side. The entire purpose of `libStorage`'s
    distributed nature is to enable clients without any knowledge of how to
    access a storage platform the ability to connect to a remote server that
    maintains that storage platform access information.

## Basic Configuration
This section outlines the most common configuration scenarios encountered by
`libStorage`'s users.

### Simple
The first example is a simple `libStorage` configuration with the VirtualBox
storage driver. The below example omits the host property, but the configuration
is still valid. If the `libstorage.host` property is not found, the server is
hosted via a temporary UNIX socket file in `/var/run/libstorage`.

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

```yaml
libstorage:
  server:
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

### TCP
The following example illustrates how to configure a `libStorage` client and
server running on the same host. The server has one endpoint on which it is
accessible - a single TCP port, 7979, bound to the localhost network interface.

```yaml
libstorage:
  host: tcp://127.0.0.1:7979
  server:
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

### TCP+TLS
The following example illustrates how to configure a `libStorage` client and
server running on the same host. The server has one endpoint on which it is
accessible - a single TCP port, 7979, bound to all of the host's network
interfaces. This means that the server is accessible via external clients, not
just those running on the same host.

Because of the public nature of this `libStorage` server, it is a good idea to
encrypt communications between client and server.

```yaml
libstorage:
  host: tcp://127.0.0.1:7979
  client:
    tls:
      certFile: $HOME/.libstorage/libstorage-client.crt
      keyFile: $HOME/.libstorage/libstorage-client.key
      trustedCertsFile: $HOME/.libstorage/trusted-certs.crt
  server:
    tls:
      certFile: /etc/libstorage/libstorage-server.crt
      keyFile: /etc/libstorage/libstorage-server.key
      trustedCertsFile: /etc/libstorage/trusted-certs.crt
      clientCertRequired: true
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

Please note that in the above example the property `libstorage.client` has been
introduced. This property is always present, even if not explicitly specified.
It exists to override `libStorage` properties for the client only, such as TLS
settings, logging, etc.

### UNIX Socket
For the security conscious, there is no safer way to run a client/server setup
on a single system than the option to use a UNIX socket. The socket offloads
authentication and relies on the file system file access to ensure authorized
users can use the `libStorage` API.

```yaml
libstorage:
  host: unix:///var/run/libstorage/localhost.sock
  server:
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

It is possible to apply TLS to the UNIX socket. Refer to the TCP+TLS section
for applying TLS to the UNIX sockets.

### Multiple Endpoints
There may be occasions when it is desirable to provide multiple ingress vectors
for the `libStorage` API. In these situations, configuring multiple endpoints
is the solution. The below example illustrates how to configure three endpoints:

 endpoint | protocol    | address | tls | localhost only
----------|-------------|---------|-----|-----------
sock | unix socket | /var/run/libstorage/localhost.sock | no | yes
private | tcp | 127.0.0.1:7979 | no | yes
public | tcp | \*:7980 | yes | no

```yaml
libstorage:
  host: unix:///var/run/libstorage/localhost.sock
  server:
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
    endpoints:
      sock:
        address: unix:///var/run/libstorage/localhost.sock
      private:
        address: tcp://127.0.0.1:7979
      public:
        address: tcp://:7980
        tls:
          certFile: /etc/libstorage/libstorage-server.crt
          keyFile: /etc/libstorage/libstorage-server.key
          trustedCertsFile: /etc/libstorage/trusted-certs.crt
          clientCertRequired: true
```

With all three endpoints defined explicitly in the above example, why leave the
property `libstorage.host` in the configuration at all? When there are no
endpoints defined, the `libStorage` server will attempt to create a default
endpoint using the value from the property `libstorage.host`. However, even
when there's at least one explicitly defined endpoint, the `libstorage.host`
property still serves a very important function -- it is the property used
by the `libStorage` client to determine which to which endpoint to connect.

### Multiple Services
All of the previous examples have used the VirtualBox storage driver as the
sole measure of how to configure a `libStorage` service. However, it is possible
to configure many services at the same time in order to provide access to
multiple storage drivers of different types, or even different configurations
of the same driver.

The following example demonstrates how to configure three `libStorage` services:

service | driver
--------|--------
virtualbox-00 | virtualbox
virtualbox-01 | virtualbox
scaleio | scaleio

Notice how the `virtualbox-01` service includes an added `integration` section.
The integration definition refers to the integration interface and parameters
specific to incoming requests through this layer. In this case we defined
`libstorage.server.services.virtualbox-01` with the
`integration.volume.operations.create.default.size` parameter set. This enables all
create requests that come in through `virtualbox-01` to have a default size of
1GB. So although it is technically the same platform below the covers,
`virtualbox-00` requests may have different default values than those defined
in `virtualbox-01`.

```yaml
libstorage:
  server:
    services:
      virtualbox-00:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes-00
          controllerName: SATA
      virtualbox-01:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes-01
          controllerName: SATA
        integration:
          volume:
            operations:
              create:
                default:
                  size: 1 # GB
      scaleio:
        driver: scaleio
        scaleio:
          endpoint: https://gateway_ip/api
          insecure: true
          userName: username
          password: password
          systemName: tenantName
          protectionDomainName: protectionDomainName
          storagePoolName: storagePoolName
```

A very important point to make about the relationship between services and
endpoints is that all configured services are available on all endpoints. In
the future this may change, and `libStorage` may support endpoint-specific
service definitions, but for now if a service is configured, it is accessible
via any of the available endpoint addresses.

Between the three services above, clearly one major difference is that two
services host one driver, VirtualBox, and the third service hosts ScaleIO.
However, why two services for one driver, in this case, VirtualBox? Because,
in addition to services being configured to host different types of drivers,
services can also host different driver configurations. In service
`virtualbox-00`, the volume path is `$HOME/VirtualBox/Volumes-00`,
whereas for service `virtualbox-01`, the volume path is
`$HOME/VirtualBox/Volumes-01`.

### Logging
Sometimes it helps to see a little more, or maybe even a little less,
information in the logs. Configuring logging is quite straight-forward:

```yaml
libstorage:
  logging:
    level: warn
  server:
    logging:
      level: info
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

The `libStorage` configuration shown above uses a global log level of `warn`,
and a more verbose, `info` log level for just the server.

## Advanced Configuration
The following sections detail every last aspect of how `libStorage` works and can
be configured.

### Embedded Configuration
If `libStorage` is embedded into another application, such as
[`REX-Ray`](https://github.com/emccode/rexray), then that application may
manage its own configuration and supply the embedded `libStorage` instance
directly with a configuration object. In this scenario, the `libStorage`
configuration files are ignored in deference to the embedding application.

### Data Directories
The first time `libStorage` is executed it will create several directories if
they do not already exist:

* `/etc/libstorage`
* `/var/log/libstorage`
* `/var/run/libstorage`
* `/var/lib/libstorage`

The above directories will contain configuration files, logs, PID files, and
mounted volumes. However, the location of these directories can also be
influenced with the environment variable `LIBSTORAGE_HOME`. All of the above
data directories will be placed in their same paths, but prefixed by the path
specified via `LIBSTORAGE_HOME`, if `LIBSTORAGE_HOME` is in fact specified.

### Configuration Methods
There are three ways to configure `libStorage`:

* Command line options
* Environment variables
* Configuration files

The order of the items above is also the order of precedence when considering
options set in multiple locations that may override one another. Values set
via CLI flags have the highest order of precedence, followed by values set by
environment variables, followed, finally, by values set in configuration files.

### Configuration Files
There are two `libStorage` configuration files - global and user:

* `/etc/libstorage/config.yml`
* `$HOME/.libstorage/config.yml`

Please note that while the user configuration file is located inside the user's
home directory, this is the directory of the user that starts `libStorage`. And
if `libStorage` is being started as a service, then `sudo` is likely being used,
which means that `$HOME/.libstorage/config.yml` won't point to *your* home
directory, but rather `/root/.libstorage/config.yml`.

### Configuration Properties
The section [Configuration Methods](#configuration-methods) mentions there are
three ways to configure libStorage: config files, environment variables, and the
command line. However, this section will illuminate the relationship between the
names of the configuration file properties, environment variables, and CLI
flags.

Below is a simple configuration file that tells the `libStorage` client where
the `libStorage` server is hosted:

```yaml
libstorage:
  host: tcp://192.168.0.20:7979
```

The property `libstorage.host` is a string. This value can also be set via
environment variables or the command line, but to do so requires knowing the
names of the environment variables or CLI flags to use. Luckily those are very
easy to figure out just by knowing the property names.

All properties that might appear in the `libStorage` configuration file
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

The following example builds on the previous. In this case we have added logging
directives to the client instance and reference how their transformation in
the table below the example.

```yaml
  libstorage:
    host: tcp://192.168.0.20:7979
    logging:
      level: warn
      stdout:
      stderr:
      httpRequests: false
      httpResponses: false
```

The following table illustrates the transformations:

Property Name | Environment Variable | CLI Flag
--------------|----------------------|-------------
`libstorage.host`    | `LIBSTORAGE_HOST`    | `--libstorageHost`
`libstorage.logging.level`    | `LIBSTORAGE_LOGGING_LEVEL`    | `--libstorageLoggingLevel`
`libstorage.logging.stdout`    | `LIBSTORAGE_LOGGING_STDOUT`    | `--libstorageLoggingStdout`
`libstorage.logging.stderr`    | `LIBSTORAGE_LOGGING_STDERR`    | `--libstorageLoggingStderr`
`libstorage.logging.httpRequests`    | `LIBSTORAGE_LOGGING_HTTPREQUESTS`    | `--libstorageLoggingHttpRequests`
`libstorage.logging.httpResponses`    | `LIBSTORAGE_LOGGING_HTTPRESPONSES`    | `--libstorageLoggingHttpResponses`

### Inherited Properties
Referring to the section on defining
[Multiple Services](./config.md#multiple-services), there is also another way
to define the TLS settings for the external TCP endpoint. The same configuration
can be rewritten and simplified in the process:

```yaml
libstorage:
  integration:
    volume:
      operations:
        create:
          default:
            size: 1 # GB
  server:
    virtualbox:
      endpoint:       http://10.0.2.2:18083
      tls:            false
      controllerName: SATA
    services:
      virtualbox-00:
        driver: virtualbox
        virtualbox:
          volumePath: $HOME/VirtualBox/Volumes-00
      virtualbox-01:
        driver: virtualbox
        virtualbox:
          volumePath: $HOME/VirtualBox/Volumes-01
```

The above example may look different than the previous one, but it's actually
the same with a minor tweak in order to simplify configuration.

While there are still two VirtualBox services defined, `virtualbox-00` and
`virtualbox-01`, neither service contains configuration information about the
VirtualBox driver other than the `volumePath` property. This is because the
change affected above is to take advantage of inherited properties.

When a property is omitted, `libStorage` traverses the configuration instance
upwards, checking certain, predefined levels known as "scopes" to see if the
property value exists there. All configured services represent a valid
configuration scope as does `libstorage.server`.

Thus when the VirtualBox driver is initialized and it checks for its properties,
while the driver may only find the `volumePath` property defined under the
configured service scope, the property access attempt travels up the
configuration stack until it hits the `libstorage.server` scope where the
remainder of the VirtualBox driver's properties *are* defined.

#### Overriding Inherited Properties
It's also possible to override inherited properties as is demonstrated in the
[Logging configuration example](#logging) above:

```yaml
libstorage:
  logging:
    level: warn
  integration:
    volume:
      operations:
        create:
          default:
            size: 1 # GB
  server:
    logging:
      level: info
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

Note that while the log level is defined at the root of the config, it's also
defined at `libstorage.server.logging.level`. The latter value of `info`
overrides the former value of `warn`. Also please remember that even had the
latter, server-specific value of `info` not been defined, an attempt by to
access the log level by the server would be perfectly valid since the attempt
would traverse up the configuration data until it found the log level defined
at the root of the configuration.

### Logging Configuration
The `libStorage` log level determines the level of verbosity emitted by the
internal logger. The default level is `warn`, but there are three other levels
as well:

 Log Level | Description
-----------|-------------
`error`    | Log only errors
`warn`     | Log errors and anything out of place
`info`     | Log errors, warnings, and workflow messages
`debug`    | Log everything

### Tasks Configuration
All operations received by the libStorage API are immediately enqueued into a
Task Service in order to divorce the business objective from the scope of the
HTTP request that delivered it. If a task completes before the HTTP request
times out, the result of the task is written to the HTTP response and sent to
the client. However, if the operation is long-lived and continues to execute
after the original HTTP request has timed out, the goroutine running the
operation will finish regardless.

In the case of such a timeout event, the client receives an HTTP status 408 -
Request Timeout. The HTTP response body also includes the task ID which can
be used to monitor the state of the remote call. The following resource URI can
be used to retrieve information about a task:

```
GET /tasks/${taskID}
```

For systems that experience heavy loads the task system can also be a source of
potential resource issues. Because tasks are kept indefinitely at this point in
time, too many tasks over a long period of time can result in a massive memory
consumption, with reports of up to 50GB and more.

That's why the configuration property `libstorage.server.tasks.logTimeout` is
available to adjust how long a task is logged before it is removed from memory.
The default value is `0` -- that is, do not log the task in memory at all.

While this is in contradiction to the task retrieval example above --
obviously a task cannot be retrieved if it is not retained -- testing and
benchmarks have shown it is too dangerous to enable task retention by default.
Instead tasks are removed immediately upon completion.

The follow configuration example illustrates a libStorage server that keeps
tasks logged for 10 minutes before purging them from memory:

```yaml
libstorage:
  server:
    tasks:
      logTimeout: 10m
```

The `libstorage.server.tasks.logTimeout` property can be set to any value that
is parseable by the Golang
[time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) function. For
example, `1000ms`, `10s`, `5m`, and `1h` are all valid values.

### Driver Configuration
There are three types of drivers:

  1. OS Drivers
  2. Storage Drivers
  3. Integration Drivers

#### OS Drivers
Operating system (OS) drivers enable `libStorage` to manage storage on
the underlying OS. Currently the following OS drivers are supported:

 Driver | Driver Name
--------|------------
Linux   | linux

The OS driver `linux` is automatically activated when `libStorage` is running on
the Linux OS.

#### Storage Drivers
Storage drivers enable `libStorage` to communicate with direct-attached or
remote storage systems. Currently the following storage drivers are supported:

 Driver | Driver Name
--------|------------
[Isilon](./storage-providers.md#isilon) | isilon
[ScaleIO](./storage-providers.md#scaleio) | scaleio
[VirtualBox](./storage-providers.md#virtualbox) | virtualbox
[EBS](./storage-providers.md#ebs) | ebs, ec2
[EFS](./storage-providers.md#efs) | efs
..more coming|

The `libstorage.server.libstorage.storage.driver` property can be used to
activate a storage drivers. That is not a typo; the `libstorage` key is repeated
beneath `libstorage.server`. This is because configuration property paths are
absolute, and when nested under an architectural component, such as
`libstorage.server`, the entire key path must be replicated.

That said, and this may seem to contradict the last point, the storage driver
property is valid *only* on the server. Well, not really. Internally the
`libStorage` client uses the same configuration property to denote its own
storage driver. This internal storage driver is actually how the `libStorage`
client communicates with the `libStorage` server.

#### Integration Drivers
Integration drivers enable `libStorage` to integrate with schedulers and other
storage consumers, such as `Docker` or `Mesos`. Currently the following
integration drivers are supported:

 Driver | Driver Name
--------|------------
Docker   | docker

The integration driver `docker` provides necessary functionality to enable
most consuming platforms to work with storage volumes.

### Volume Configuration
This section describes various global configuration options related to an
integration driver's volume operations, such as mounting and unmounting volumes.

#### Volume Properties
The properties listed below are the global properties valid for an integration
driver's volume-related properties.

parameter|description
---------|-----------
`libstorage.integration.volume.operations.mount.preempt`|Forcefully take control of volumes when requested
`libstorage.integration.volume.operations.mount.path`|The default host path for mounting volumes
`libstorage.integration.volume.operations.mount.rootPath`|The path within the volume to return to the integrator (ex. `/data`)
`libstorage.integration.volume.operations.create.disable`|Disable the ability for a volume to be created
`libstorage.integration.volume.operations.remove.disable`|Disable the ability for a volume to be removed

The properties in the next table are the configurable parameters that affect
the default values for volume creation requests.

parameter|description
---------|-----------
`libstorage.integration.volume.operations.create.default.size`|Size in GB
`libstorage.integration.volume.operations.create.default.iops`|IOPS
`libstorage.integration.volume.operations.create.default.type`|Type of Volume or Storage Pool
`libstorage.integration.volume.operations.create.default.fsType`|Type of filesystem for new volumes (ext4/xfs)
`libstorage.integration.volume.operations.create.default.availabilityZone`|Extensible parameter per storage driver

#### Disable Create
The disable create feature enables you to disallow any volume creation activity.
Any requests will be returned in a successful manner, but the create will not
get passed to the backend storage platform.

```yaml
libstorage:
  integration:
    volume:
      operations:
        create:
          disable: true
```

#### Disable Remove
The disable remove feature enables you to disallow any volume removal activity.
Any requests will be returned in a successful manner, but the remove will not
get passed to the backend storage platform.

```yaml
libstorage:
  integration:
    volume:
      operations:
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
libstorage:
  integration:
    volume:
      operations:
        mount:
          preempt: true
```

Driver|Supported
------|---------
Isilon|Not yet
ScaleIO|Yes
VirtualBox|Yes
EBS|Yes
EFS|No

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
libstorage:
  integration:
    volume:
      operations:
        unmount:
          ignoreUsedCount: true
```

Currently a reset of the service will cause the counts to be reset.  This
will cause issues if *multiple containers* are sharing a volume.  If you are
sharing volumes, it is recommended that you reset the service along with the
accompanying container runtime (if this setting is false) to ensure they are
synchronized.  

#### Volume Path Cache
In order to optimize `Path` requests, the paths of actively mounted volumes
returned as the result of a `List` request are cached. Subsequent `Path`
requests for unmounted volumes will not dirty the cache. Only once a volume
has been mounted will the cache be marked dirty and the volume's path retrieved
and cached once more.

The following configuration example illustrates the two path cache properties:

```yaml
libstorage:
  integration:
    volume:
      operations:
        path:
          cache:
            enabled: true
            async:   true
```

Volume path caching is enabled and asynchronous by default, so it's possible to
entirely omit the above configuration excerpt from a production deployment, and
the system will still use asynchronous caching. Setting the `async` property to
`false` simply means that the initial population of the cache will be handled
synchronously, slowing down the program's startup time.

#### Volume Root Path
When volumes are mounted there can be an additional path that is specified to
be created and passed as the valid mount point.  This is required for certain
applications that do not want to place data from the root of a mount point.

The default is the `/data` path.  If a value is set by
`linux.integration.volume.operations.mount.rootPath`, then the default will be
overwritten.

```yaml
libstorage:
  integration:
    volume:
      operations:
        mount:
          rootPath: /data
```
