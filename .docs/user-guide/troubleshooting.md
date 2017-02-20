# Troubleshooting

It's not doing what I expected...

---

## Configuration Validation
REX-Ray configuration will usually be set in the yaml file located 
at `/etc/rexray/config.yml`

If you are having issues, or anytime you edit your configuration, you may wish 
to validate the result by cut and paste to an online validator such as this 
one (if your configuration contains sensitive security credentials install and use a local validator instead): 
[http://yaml-online-parser.appspot.com/](http://yaml-online-parser.appspot.com/)

If REX-Ray is running as a service, the service needs to be restarted to pick 
up the change.

```sh
$ sudo rexray service restart
```

The configuration file is the conventional method to set configuration, but 
this can be supplemented or replaced by environment variables and command 
line option flags. For example, when running a centralized REX-Ray controller 
service in a Docker container, you would likely use an environment variables 
configuration. 

If configuration is set in in multiple locations, the precedence order is:

Command line overrides -> environment overrides -> config file

The command `rexray env` can be used to dump the current _effective_ 
configuration - in an environment variable form, regardless of whether 
environment variables were used to set the configuration.

### Ultra simplified configuration 
A minimal complexity configuration can be used as an experiment to eliminate 
issues external to configuration.

1. Shut down the REX-Ray service: `sudo rexray service stop`
2. Save your original config: `sudo mv /etc/rexray/config.yml /etc/rexray/config.$(date +%Y-%m-%d-%T).yml`
3. Create a new minimal config `vi \etx\rexray\config.yml`

This is a minimal configuration for **AWS EBS** volumes. Other storage drivers 
are similar and will mostly vary along the lines of authorization credentials, 
and possibly storage API address identification.

```yaml
libstorage:
  service: ebs
ebs:
  accessKey: YOUR-ACCESS-KEY
  secretKey: your-secret-key
  region:us-west-2
```

This will parse the config and emit it in environment variable form 

```sh
$ rexray env
```

Next test that your REX-Ray installation works in interactive (non service) 
form:

```sh
$ rexray volume ls –s ebs
```


if this works, move on to operating as a service:

```sh
$ sudo rexray service start
```

Look at the log (`/var/log/rexray/rexray.log`) which should show the version 
and no errors. 

A few other minimal REX-Ray configuration examples for other platforms:

**ScaleIO** (host_ip varies)

```yaml
libstorage:
  service: scaleio
scaleio:
  endpoint: https://SCALEIO_GATEWAY/api
  insecure: true
  thinOrThick: ThinProvisioned
```

**VirtualBox** (endpoint IP and volumePath varies)

```yaml
libstorage:
  service: virtualbox
virtualbox:
  endpoint: http://10.0.2.2:18083
  volumePath: /Users/<your-name>/VirtualBox/Volumes
  controllerName: SATA
```

## Investigating REX-Ray in server (controller) mode

### Confirm the server is running (from the server host):

```sh
$ sudo rexray service status

Active: active (running) since Sun 2017-02-12 02:19:16 UTC; 1 day 16h ago
```

### Confirm the server is listening on a Unix socket, and identify the socket:

```sh
$ ss -l | grep rexray
u_str  LISTEN     0      128    /run/docker/plugins/rexray.sock 20959                 * 0
```

### Confirm the server is listening on a network port, and identify the port. 

```sh
$ lsof -i -P -sTCP:LISTEN | grep rexray
rexray  949 root    5u  IPv4  20916      0t0  TCP localhost:7979 (LISTEN)
rexray  949 root    7u  IPv6  20917      0t0  TCP *:7980 (LISTEN)
```

If the `lsof` package is not present in your host, you can either install it, 
or try using `ss` as a replacement:

```sh
$ ss -ltnp | grep rexray
LISTEN     0      128    127.0.0.1:7979                     *:*                   users:(("rexray",pid=949,fd=5))
LISTEN     0      128         :::7980                    :::*                   users:(("rexray",pid=949,fd=7))
```

### Confirm server access from clients

After confirming that the service is listening as expected, you can use `curl` 
to actually invoke the libStorage REST API. This example shows doing it using 
localhost but in a client server topology deployment you can also invoke it 
from your client nodes (using the server’s external IP) to confirm that there 
are no routing or firewall issues. This example uses http, but if you have 
installed certificates, you should use https instead. This particular 
invocation of the REST API lists the services (storage provider classes) 
that are available to clients.

```sh
$ curl http://127.0.0.1:7979/services

{
  "ebs": {
    "name": "ebs",
    "driver": {
      "name": "ebs",
      "type": "block",
      "nextDevice": {
        "ignore": false,
        "prefix": "xvd",
        "pattern": "[f-p]"
      }
    }
  }
}
```
## Some common errors

General note: When running on a public cloud provider the error messages in 
the log resulting from calls to a cloud provider API are repeated from the 
cloud provider’s API. Sometimes these are not intuitive. For example a bad 
credential might result in an error message related to a non-existent object. 
Performing a google search in the context of your cloud provider, rather than 
REX-Ray can sometimes prove to be helpful. 

#### Issue: Attempt to start REX-Ray service as a normal account, rather that root or other privileged:
```sh
$ rexray service start
Failed to start rexray.service: Interactive authentication required.
```

#### Issue: Leave out service specification on command line, when multiple services are configured:
```sh
$ rexray volume ls
FATA[0000] http error                                    status=404
```

#### Issue: Place REX-Ray configuration file in wrong place, or fail to create one at all:
```sh
$ rexray volume ls
ERRO[0000] error starting libStorage server              error.configKey=libstorage.server.services error.obj=<nil> time=1486771348853`
```

#### Issue: Define bad credentials for the storage provider API

In this case (AWS), results in a log entry like this:

```
time="2017-02-10T22:34:44Z" level=error msg="error getting volume" host="tcp://127.0.0.1:7979" inner="AuthFailure: AWS was not able to validate the provided access credentials\n\tstatus code: 401, request id: ea6587a90-f29d-4f14-99da-6a7ec7cb05c1" instanceID="ebs=i-0213cc11c4ade43fb,availabilityZone=us-west-2a&region=us-west-2" route=volumesForService server=dew-lady-tw service=ebs storageDriver=ebs task=0 time=1486766084055 tls=false txCR=1486726083 txID=05a5e1fd-094f-40ec-63f6-448d26ddde4f
```

#### Issue: Leave out a service definition, or erroneously name it, in the REX-Ray configuration:

```
On start ERRO[0000] error starting libStorage server              error.obj=<nil> error.configKey=libstorage.server.services time=1486772035132

In log:
time="2017-02-11T00:13:25Z" level=error msg="error starting libStorage server" error.configKey=libstorage.server.services error.obj=<nil> time=1486772005732 
time="2017-02-11T00:13:25Z" level=error msg="default module(s) failed to initialize" error.obj=<nil> error.configKey=libstorage.server.services time=1486772005732 
time="2017-02-11T00:13:25Z" level=error msg="daemon failed to initialize" error.configKey=libstorage.server.services error.obj=<nil> time=1486772005732 
time="2017-02-11T00:13:25Z" level=error msg="error starting rex-ray" error.obj=<nil> error.configKey=libstorage.server.services time=1486772005732 
```

## Logging

Enable detailed logging through the REX-Ray configuration file. Available 
levels from least detailed, to most are: 

1. panic
2. fatal
2. error
3. warn
3. info
4. debug

There is also a setting to enable logging of transactions over the internal 
REST API.

Also note that for any command line invocation, the `-l debug` flag can be 
added to get verbose output.

Because REX-Ray is based on a second libStorage component, REX-Ray has two 
independent log settings in it’s configuration, rexray, and libstorage. For 
troubleshooting, set the log level for both REX-Ray and libStorage to `debug` 
and enable HTTP request and response tracing for libStorage. Note that 
additional lines may be present in your configuration - this example has 
been simplified to highlight the relevant settings.

REX-Ray config: (`/etc/rexray/config.yml`)

```yaml
rexray:
  logLevel:        debug
libstorage:
  logging:
    level:         debug
    httpRequests:  true
    httpResponses: true
```

Restart the REX-Ray service after changing the configuration:

```sh
$ rexray service restart
```

The REX-Ray service log file is stored at `/var/log/rexray/rexray.log` by 
default

Note that in specialized scenarios using REX-Ray in a client mode, debug 
output is emitted to the console.

## Installation related troubleshooting

Installables are available here [https://dl.bintray.com/emccode/rexray/](https://dl.bintray.com/emccode/rexray/)

The REX-Ray bintray repository may be used to access .rpm, .deb and other 
formats for specialized installation requirements.

You can open this URL from a browser, or do a simple curl or wget -O - to 
verify connectivity to the repository, to eliminate routing or firewall issues.

A normal installation downloads and invokes an installation script in one 
line. You can forcibly install a particular version like this:

```sh
$ curl -sSL https://dl.bintray.com/emccode/rexray/install | sh -s -- stable 0.8.0
```

The version shows during an install, but to see it again later:

```sh
$ rexray version
```

To see the full set of options for the install script, while proving you have 
basic connectivity to your target (eliminate routing, firewall, DNS, etc, 
issues), execute this.

```sh
$ curl -sSL -o install https://dl.bintray.com/emccode/rexray/install
```

This downloads the install script as a file. It’s a bash script, with 
available options described internally, as comments.

Supported installation targets are here - but some platforms may have 
version restrictions: [https://rexray.readthedocs.io/en/stable/#operating-system-support](https://rexray.readthedocs.io/en/stable/#operating-system-support)

### Uninstall

On Linux platforms using the rpm package manager `rpm -e rexray` can be 
used to remove REX-Ray prior to installation of an alternate version.

On Linux platforms using the deb package manager, `dpkg --remove rexray` can 
be used to remove REX-Ray prior to installation of an alternate version
