# Schedulers

Scheduling storage one resource at a time...

---

## Overview
This page reviews the scheduling systems supported by `libStorage`.

## Docker
`libStorage`'s '`Docker Integration Driver` is compatible with 1.10+.

However, `Docker 1.10.2+` is suggested if volumes are shared between containers
or interactive volume inspection requests are desired via the `/volumes`,
`/volumes/{service}`, and  `/volumes/{service}/{volumeID}` resources.

Please  note that this is *not* the same as
[Docker's Volume Plug-in](https://docs.docker.com/engine/extend/plugins_volume/).
`libStorage` does not provide a way to expose the `Docker Integration Driver`
via the `Docker Volume Plug-in`, but `REX-Ray`, which embeds `libStorage`,
does.

### Example Configuration
Below is an example `config.yml` that can be used.  The `volume.mount.preempt`
is an optional parameter here which enables any host to take control of a
volume irrespective of whether other hosts are using the volume.  If this is
set to `false` then plugins should ensure `safety` first by locking the
volume from to the current owner host. We also specify `docker.size` which will
create all new volumes at the specified size in GB.

```yaml
libstorage:
  host: unix:///var/run/libstorage/localhost.sock
  integration:
    volume:
      mount:
        preempt: true
      create:
        default:
          size: 1 # GB
  server:
    endpoints:
      localhost:
        address: unix:///var/run/libstorage/localhost.sock
    services:
      virtualbox:
        driver: virtualbox
        virtualbox:
          endpoint:       http://10.0.2.2:18083
          tls:            false
          volumePath:     $HOME/VirtualBox/Volumes
          controllerName: SATA
```

### Configuration Properties
The Docker integration driver adheres to the properties described in the
section on an
[Integration driver's volume-related properties](./config.md#volume-properties).

Please note that with `Docker` 1.9.1 or below, it is recommended that the
property `libstorage.integration.volume.remove.disable` be set to `true` in
order to prevent `Docker` from removing external volumes in-use by containers
that are forcefully removed.

### Caveats
If you restart the process which embeds `libStorage` and hosts the
`Docker Volume Plug-in` while volumes *are shared between Docker containers*,
then problems may arise when stopping one of the containers sharing the volume.

It is suggested to avoid stopping these containers at this point until all
containers sharing the volumes can be stopped. This will enable the unmount
process to proceed cleanly.
