# Troubleshooting

It's not doing what I expected...

---

## Solving problems
This section details the usual places and methods to look and use when
investigating a problem.

### Is REX-Ray running?
Confirm REX-Ray is running with the following command:
```shell
$ sudo rexray service status

Active: active (running) since Sun 2017-02-12 02:19:16 UTC; 1 day 16h ago
```

### Is REX-Ray listening?
Confirm REX-Ray is listening on a UNIX socket with the following command:
```shell
$ sudo lsof -noPU -a -c rexray
COMMAND  PID USER   FD   TYPE             DEVICE OFFSET  NODE NAME
rexray  3228 root    5u  unix 0xffff8800cc6eb800    0t0 38218 /var/run/libstorage/277666194.sock
rexray  3228 root    7u  unix 0xffff880117792000    0t0 34741 socket
rexray  3228 root    8u  unix 0xffff880117792400    0t0 38221 /var/run/libstorage/277666194.sock
rexray  3228 root    9u  unix 0xffff8800cc6e8c00    0t0 38226 /run/docker/plugins/rexray.sock
```

Confirm REX-Ray is listening on a TCP port with the following command:
```shell
$ sudo lsof -noP -i -sTCP:LISTEN -a -c rexray
COMMAND  PID USER   FD   TYPE DEVICE OFFSET NODE NAME
rexray  3400 root    4u  IPv4  38868    0t0  TCP 127.0.0.1:5002 (LISTEN)
```

### Is REX-Ray talking?
After confirming that the service is listening as expected, you can use `curl`
to actually invoke the libStorage REST API. This example shows doing it using
localhost but in a client server topology deployment you can also invoke it
from your client nodes (using the server’s external IP) to confirm that there
are no routing or firewall issues. This example uses http, but if you have
installed certificates, you should use https instead. This particular
invocation of the REST API lists the services (storage provider classes)
that are available to clients.

```shell
$ curl http://127.0.0.1:5002/services
```

```json
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

## Common Errors
This section reviews common errors encountered when using REX-Ray.

!!! note "note"
    General note: When running on a public cloud provider the error messages in
    the log resulting from calls to a cloud provider API are repeated from the
    cloud provider’s API. Sometimes these are not intuitive. For example a bad
    credential might result in an error message related to a non-existent
    object. Performing a google search in the context of your cloud provider,
    rather than REX-Ray can sometimes prove to be helpful.

### Starting REX-Ray
This error occurs when attempting to start REX-Ray service as a normal account,
rather than root or with `sudo`:
```shell
$ rexray service start
Failed to start rexray.service: Interactive authentication required.
```

### Omitted service flag
This error occurs when there are multiple services configured and the service
specification is omitted from the command line:
```shell
$ rexray volume ls
FATA[0000] http error                                    status=404
```

### Missing REX-Ray config file
This error can occur when the REX-Ray configuration file is in the incorrect
location or one does not exist at all:
```shell
$ rexray volume ls
ERRO[0000] error starting libStorage server              error.configKey=libstorage.server.services error.obj=<nil> time=1486771348853`
```

### Invalid provider credentials
This error occurs when invalid credentials are provided for the storage
provider. The example below uses the EBS storage provider:

```shell
time="2017-02-10T22:34:44Z" level=error msg="error getting volume" host="tcp://127.0.0.1:7979" inner="AuthFailure: AWS was not able to validate the provided access credentials\n\tstatus code: 401, request id: ea6587a90-f29d-4f14-99da-6a7ec7cb05c1" instanceID="ebs=i-0213cc11c4ade43fb,availabilityZone=us-west-2a&region=us-west-2" route=volumesForService server=dew-lady-tw service=ebs storageDriver=ebs task=0 time=1486766084055 tls=false txCR=1486726083 txID=05a5e1fd-094f-40ec-63f6-448d26ddde4f
```

### Omitted service definition
This error occurs when the configuration file omits a service definition or
one is named erroneously:

#### Console
```shell
ERRO[0000] error starting libStorage server              error.obj=<nil> error.configKey=libstorage.server.services time=1486772035132
```

#### Service Log
```shell
time="2017-02-11T00:13:25Z" level=error msg="error starting libStorage server" error.configKey=libstorage.server.services error.obj=<nil> time=1486772005732
time="2017-02-11T00:13:25Z" level=error msg="default module(s) failed to initialize" error.obj=<nil> error.configKey=libstorage.server.services time=1486772005732
time="2017-02-11T00:13:25Z" level=error msg="daemon failed to initialize" error.configKey=libstorage.server.services error.obj=<nil> time=1486772005732
time="2017-02-11T00:13:25Z" level=error msg="error starting rex-ray" error.obj=<nil> error.configKey=libstorage.server.services time=1486772005732
```
