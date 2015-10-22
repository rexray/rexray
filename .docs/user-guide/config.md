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

```
logLevel: warn
osDrivers: linux
volumeDrivers: docker
```

## General Options
The following are the general configuration options for `REX-Ray`.

 EnvVar | YAML | CLI  | Description | Default
--------|------|------|-------------|---------
`REXRAY_LOGLEVEL` | `logLevel` | `--logLevel`, `-l` | The valid log levels are `error`, `warn`, `info`, `debug` | `warn`
`REXRAY_OSDRIVERS` | `osDrivers` | `--osDrivers` |  A space-delimited list of OS driver names which instructs `REX-Ray` to only do checks using the specified drivers | `linux`
`REXRAY_VOLUMEDRIVERS` | `volumeDrivers` | `--volumeDrivers` |  A space-delimited list of storage driver names which instructs `REX-Ray` to only do checks using the specified drivers | `docker`
`REXRAY_STORAGEDRIVERS` | `storageDrivers` | `--storageDrivers` |  A space-delimitedlist of volume driver names which instructs `REX-Ray` to only do checks using the specified drivers |
