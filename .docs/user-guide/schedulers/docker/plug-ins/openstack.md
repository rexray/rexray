# OpenStack

Cinder

---

## Cinder
The Cinder plug-in can be installed with the following command:

```bash
$ docker plugin install rexray/cinder \
  CINDER_AUTHURL=http://xxxx \
  CINDER_USERNAME=rexray \
  CINDER_PASSWORD=xxx \
  CINDER_TENANTID=xxxxxxx
```
The safer way to install it, is get your api.rc file, and source it:
```bash
$ source yourprojectname.rc
```
And use the OS_variables to install the plugin

```bash
$ docker plugin install rexray/cinder CINDER_AUTHURL=$OS_AUTH_URL \
CINDER_USERNAME=$OS_USERNAME CINDER_PASSWORD=$OS_PASSWORD \
CINDER_TENANTID=$OS_PROJECT_ID CINDER_DOMAINNAME=$OS_USER_DOMAIN_NAME
```


##### Requirements
The Cinder plug-in requires that GCE compute instance has Read/Write Cloud API
access to the Compute Engine and Storage services.

##### Privileges
The Cinder plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

##### Configuration
The following environment variables can be used to configure the Cinder
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`CINDER_AUTHURL` | The keystone authentication API |  | ✓
`CINDER_USERID` | OpenStack userId for cinder access | |
`CINDER_USERNAME` | OpenStack username for cinder access | |
`CINDER_PASSWORD` | OpenStack user password for cinder access | |
`CINDER_TOKENID` | OpenStack tokenId for cinder access | |
`CINDER_TRUSTID` | OpenStack trustId for cinder access | |
`CINDER_TENANTID` | OpenStack tenantId | |
`CINDER_TENANTNAME` | OpenStack tenantId | |
`CINDER_DOMAINID` | OpenStack domainId to authenticate | |
`CINDER_DOMAINNAME` | OpenStack domainName to authenticate | |
`CINDER_REGIONNAME` | OpenStack regionName to authenticate | |
`CINDER_AVAILABILITYZONENAME` | OpenStack availability zone for volumes | |
`CINDER_ATTACHTIMEOUT` | Timeout for attaching volumes | `1m` |
`CINDER_CREATETIMEOUT` | Timeout for creating volumes | `10m` |
`CINDER_DELETETIMEOUT` | Timeout for creating volumes | `10m` |
`HTTP_PROXY` | Address of HTTP proxy server to gain access to API endpoint | |

##### Troubleshooting

Most errors occurs due invalid Cinder configuration, to make sure if everiting is ok, source your api file and install the openstack client, then:

```bash
$openstack volume list
```

If youir volumes are listed adn/or no error is displayed, you are good to go

Otherwise you can use the debug mode:
```bash
REXRAY_DEBUG=true
```



