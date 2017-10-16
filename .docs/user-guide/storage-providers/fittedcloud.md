# FittedCloud

Right size storage

---

## Overview
Another example of the great community shared by the libStorage project, the
talented people at FittedCloud have provided a driver for their EBS optimizer.

<a name="fittedcloud-ebs"></a>
<a name="fittedcloud-ebs-optimizer"></a>

## EBS Optimizer
The FittedCloud EBS Optimizer driver registers a storage driver named
`fittedcloud` with the libStorage service registry and provides the ability to
connect and manage thin-provisioned EBS volumes for EC2 instances.

!!! note
    This version of the FittedCloud driver only supports configurations where
    client and server are on the same host.  The libStorage server must be
    running on each node along side with the FittedCloud Agent.

!!! note
    This version of the FittedCloud driver does not support co-existing with the
    ebs driver on the same host. As a result it also doesn't support optimizing
    existing EBS volumes. See the [Examples](#fittedcloud-examples) section
    below for a running example.

!!! note
    The FittedCloud driver does not yet support snapshots or tags.

### Requirements
This driver has the following requirements:

* AWS account
* VPC - EBS can be accessed within VPC
* AWS Credentials
* FittedCloud Agent software

<a name="fittedcloud-getting-started"></a>

### Getting Started
Before starting, please make sure to register as a user by visiting the
FittedCloud [customer website](https://customer.fittedcloud.com/register).
Once an account is activated it will be assigned a user ID, which can be found
on the Settings page after logging into the web site.

The following commands will download and install the latest FittedCloud Agent
software. The flags `-o S -m` enable new thin volumes to be created via the
docker command instead of optimizing existing EBS volumes.
Please replace the `<User ID>` with a FittedCloud user ID.

```sh
$ curl -skSL 'https://customer.fittedcloud.com/downloadsoftware?ver=latest' \
  -o fcagent.run
$ sudo bash ./fcagent.run -- -o S -m -d <User ID>
```

Please refer to FittedCloud
[website](https://customer.fittedcloud.com/download) for more details.

<a name="fittedcloud-config"></a>

### Configuration
The following is an example with all possible fields configured.  For a running
example see the [Examples](#fittedcloud-examples) section.

```yaml
ebs:
  accessKey:      XXXXXXXXXX
  secretKey:      XXXXXXXXXX
  kmsKeyID:       abcd1234-a123-456a-a12b-a123b4cd56ef
  statusMaxAttempts:  10
  statusInitialDelay: 100ms
  statusTimeout:      2m
```

#### Configuration Notes
- FittedCloud driver shares the ebs driver's configuration
parameters.
- The `accessKey` and `secretKey` configuration parameters are optional and
should be used when explicit AWS credentials configuration needs to be provided.
FittedCloud driver uses official golang AWS SDK library and supports all other
ways of providing access credentials, like environment variables or instance
profile IAM permissions.
- If the `kmsKeyID` field is specified it will be used as the encryption key for
all volumes that are created with a truthy encryption request field.
- `statusMaxAttempts` is the number of times the status of a volume will be
  queried before giving up when waiting on a status change
- `statusInitialDelay` specifies a time duration used to wait when polling
  volume status. This duration is used in exponential backoff, such that the
  first wait will be for this duration, the second for 2x, the third for 4x,
  etc. The units of the duration must be given (e.g. "100ms" or "1s").
- `statusTimeout` is a maximum length of time that polling for volume status can
  occur. This serves as a backstop against a stuck request of malfunctioning API
  that never returns.

<a name="fittedcloud-examples"></a>

### Examples
The following example illustrates how to configured the FittedCloud driver:

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: fittedcloud
  server:
    services:
      fittedcloud:
        driver: fittedcloud
ebs:
  accessKey:  XXXXXXXXXX
  secretKey:  XXXXXXXXXX
```

Additional information on configuring the FittedCloud driver may be found
at [this](https://goo.gl/I6mf20) location.
