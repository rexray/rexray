# Ceph

RBD

---

<a name="ceph-rbd"></a>

## RADOS Block Device (RBD)
The Ceph RBD driver registers a driver named `rbd` with the `libStorage` driver
manager and is used to connect and mount RADOS Block Devices from a Ceph
cluster.

### Requirements

* The `ceph` and `rbd` binary executables must be installed on the host
* The `rbd` kernel module must be installed
* A `ceph.conf` file must be present in its default location
  (`/etc/ceph/ceph.conf`)
* The ceph `admin` key must be present in `/etc/ceph/`

### Configuration
The following is an example with all possible fields configured. For a running
example see the [Examples](#ceph-rbd-examples) section.

```yaml
rbd:
  defaultPool: rbd
  testModule: true
```

#### Configuration Notes

* The `defaultPool` parameter is optional, and defaults to "rbd". When set, all
  volume requests that do not reference a specific pool will use the
  `defaultPool` value as the destination storage pool.
* The `testModule` parameter is optional, and defaults to "true". This setting
  indicates whether the libStorage client should test if the `rbd` kernel module
  is loaded, with the side-effect of loading it if is not already loaded. This
  setting should be disabled when the driver is executing inside of a container.

### Runtime behavior

The Ceph RBD driver only works when the client and server are on the same node.
There is no way for a centralized `libStorage` server to attach volumes to
clients, therefore the `libStorage` server must be running on each node that
wishes to mount RBD volumes.

The RBD driver uses the format of `<pool>.<name>` for the volume ID. This allows
for the use of multiple pools by the driver. During a volume create, if the
volume ID is given as `<pool>.<name>`, a volume named *name* will be created in
the *pool* storage pool. If no pool is referenced, the `defaultPool` will be
used.

Both *pool* and *name* may only contain alphanumeric characters, underscores,
and dashes.

When querying volumes, the driver will return all RBDs present in all pools in
the cluster, prefixing each volume with the appropriate `<pool>.` value.

All RBD creates are done using the default 4MB object size, and using the
"layering" feature bit to ensure greatest compatibility with the kernel clients.

### Activating the Driver
To activate the Ceph RBD driver please follow the instructions for
[activating storage drivers](../servers/libstorage.md#storage-drivers), using `rbd` as the
driver name.

### Troubleshooting

* Make sure that `ceph` and `rbd` commands work without extra parameters for
  ID, key, and monitors. All configuration must come from `ceph.conf`.
* Check status of the ceph cluster with `ceph -s` command.

<a name="ceph-rbd-examples"></a>

### Examples

Below is a full `config.yml` that works with RBD

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: rbd
  server:
    services:
      rbd:
        driver: rbd
        rbd:
          defaultPool: rbd
```

### Caveats
* Snapshot and copy functionality is not yet implemented
* libStorage Server must be running on each host to mount/attach RBD volumes
* There is not yet options for using non-admin cephx keys or changing RBD create
  features
* Volume pre-emption is not supported. Ceph does not provide a method to
  forcefully detach a volume from a remote host -- only a host can attach and
  detach volumes from itself.
* RBD advisory locks are not yet in use. A volume is returned as "unavailable"
  if it has a watcher other than the requesting client. Until advisory locks are
  in place, it may be possible for a client to attach a volume that is already
  attached to another node. Mounting and writing to such a volume could lead to
  data corruption.
