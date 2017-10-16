# Ceph

RBD

---

<a name="ceph-rbd"></a>

## RADOS Block Device (RBD)
The RBD plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/rbd RBD_DEFAULTPOOL=rbd
```

### Requirements
The RBD plug-in requires that the host has a fully working environment for
mapping Ceph RBDs, including having the RBD kernel module already loaded. The
cluster configuration and authentication files must be present in `/etc/ceph`.

### Privileges
The RBD plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`, `/etc/ceph`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

### Configuration
The following environment variables can be used to configure the RBD
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`RBD_CEPHARGS` | Text to set in the `CEPH_ARGS` environment variable | ""
`RBD_DEFAULTPOOL` | Default Ceph pool for volumes | `rbd`
