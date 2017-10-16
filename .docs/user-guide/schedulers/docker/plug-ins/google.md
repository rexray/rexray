# Google

Cloud storage

---

<a name="gce-persistent-disk"></a>

## GCE Persistent Disk
The GCEPD plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/gcepd \
  GCEPD_TAG=rexray
```

##### Requirements
The GCEPD plug-in requires that GCE compute instance has Read/Write Cloud API
access to the Compute Engine and Storage services.

**NOTE:** GCE persistent disks cannot be created if their name contains an underscore.
Docker will automatically append prefixes with underscores to your volume names when
they are created as part of a compose file, so if you're creating volumes with this plugin
using compose (or stack deploy), be sure to set `GCEPD_CONVERTUNDERSCORES` to `true`.

##### Privileges
The GCEPD plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

##### Configuration
The following environment variables can be used to configure the GCEPD
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`GCEPD_CONVERTUNDERSCORES` | Set to `true` if the plugin will reference persistent disks through a `docker-compose.yml` file | `false` |
`GCEPD_DEFAULTDISKTYPE` | The default disk type to consume | `pd-ssd` |
`GCEPD_STATUSINITIALDELAY` | Time duration used to wait when polling volume status | `100ms` |
`GCEPD_STATUSMAXATTEMPTS` | Number of times the status of a volume will be queried before giving up | `10` |
`GCEPD_STATUSTIMEOUT` | Maximum length of time that polling for volume status can occur | `2m` |
`GCEPD_TAG` | Only use volumes that are tagged with a label | |
`GCEPD_ZONE` | GCE Availability Zone | |
`HTTP_PROXY` | Address of HTTP proxy server to gain access to API endpoint | |
