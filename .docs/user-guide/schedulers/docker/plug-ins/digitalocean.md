# DigitalOcean

Block Storage

---

<a name="digitalocean-block-storage"></a>
<a name="dobs"></a>

## DO Block Storage
The DOBS plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/dobs \
  DOBS_REGION=sfo2 \
  DOBS_TOKEN=0907868f343d86076f261958123638248ae2321434dd4f1b74773ddb9320de43
```

##### Requirements
The DOBS plug-in requires that your DigitalOcean droplet is running in a region that
supports block storage.

**NOTE:** DigitalOcean volumes cannot be created if their name contains an underscore.
Docker will automatically append prefixes with underscores to your volume names when
they are created as part of a compose file, so if you're creating volumes with this plugin
using compose (or stack deploy), be sure to set `DOBS_CONVERTUNDERSCORES` to `true`.

##### Privileges
The DOBS plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

##### Configuration
The following environment variables can be used to configure the DOBS
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`DOBS_CONVERTUNDERSCORES` | Set to `true` if the plugin will create volumes at DigitalOcean via e.g. a `docker-compose.yml` file | `false` |  
`DOBS_REGION` | The region where volumes should be created | | ✓
`DOBS_STATUSINITIALDELAY` | Time duration used to wait when polling volume status | `100ms` |
`DOBS_STATUSMAXATTEMPTS` | Number of times the status of a volume will be queried before giving up | `10` |
`DOBS_STATUSTIMEOUT` | Maximum length of time that polling for volume status can occur | `2m` |
`DOBS_TOKEN` | Your DigitalOcean access token | | ✓
`HTTP_PROXY` | Address of HTTP proxy server to gain access to API endpoint | |
