# Google

Cloud storage

---

## Overview
REX-Ray ships with support for Google Compute Engine (GCE) as well.

<a name="gce-persistent-disk"></a>

## GCE Persistent Disk
The Google Compute Engine Persistent Disk (GCEPD) driver registers a driver
named `gcepd` with the libStorage service registry and is used to connect and
mount Google Compute Engine (GCE) persistent disks with GCE machine instances.

### Requirements
* GCE account
* The libStorage server must be running on a GCE instance created with a Service
  Account with appropriate permissions, or a Service Account credentials file
  in JSON format must be supplied. If not using the Compute Engine default
  Service Account with the Cloud Platform/"all cloud APIs" scope, create a new
  Service Account via the [IAM Portal](https://console.cloud.google.com/iam-admin/serviceaccounts).
  This Service Account requires the `Compute Engine/Instance Admin`,
  `Compute Engine/Storage Admin`, and `Project/Service Account Actor` roles.
  Then create/download a new private key in JSON format. see
  [creating a service account](https://developers.google.com/identity/protocols/OAuth2ServiceAccount#creatinganaccount)
  for details. The libStorage service must be restarted in order for permissions
  changes on a service account to take effect.

### Configuration
The following is an example with all possible fields configured. For a running
example see the [Examples](#gce-persistent-disk-examples)
section.

```yaml
gcepd:
  keyfile: /etc/gcekey.json
  zone: us-west1-b
  defaultDiskType: pd-ssd
  tag: rexray
  statusMaxAttempts:  10
  statusInitialDelay: 100ms
  statusTimeout:      2m
  convertUnderscores: false
```

#### Configuration Notes
* The `keyfile` parameter is optional. It specifies a path on disk to a file
  containing the JSON-encoded Service Account credentials. This file can be
  downloaded from the GCE web portal. If `keyfile` is specified, the GCE
  instance's service account is not considered, and is not necessary. If
  `keyfile` is *not* specified, the application will try to lookup
  [application default credentials](https://developers.google.com/identity/protocols/application-default-credentials).
  This has the effect of looking for credentials in the priority described
  [here](https://godoc.org/golang.org/x/oauth2/google#FindDefaultCredentials).
* The `zone` parameter is optional, and configures the driver to *only* allow
  access to the given zone. Creating and listing disks from other zones will be
  denied. If a zone is not specified, the zone from the client Instance ID will
  be used when creating new disks.
* The `defaultDiskType` parameter is optional and specifies what type of disk
  to create, either `pd-standard` or `pd-ssd`. When not specified, the default
  is `pd-ssd`.
* The `tag` parameter is optional, and causes the driver to create or return
  disks that have a matching tag. The tag is implemented by using the GCE
  label functionality available in the beta API. The value of the `tag`
  parameter is used as the value for a label with the key `libstoragetag`.
  Use of this parameter is encouraged, as the driver will only return volumes
  that have been created by the driver, which is most useful to eliminate
  listing the boot disks of every GCE disk in your project/zone. If you wsih to
  "expose" previously created disks to the `GCEPD` driver, you can edit the
  labels on the existing disk to have a key of `libstoragetag` and a value
  matching that given in `tag`.
* `statusMaxAttempts` is the number of times the status of a volume will be
  queried before giving up when waiting on a status change
* `statusInitialDelay` specifies a time duration used to wait when polling
  volume status. This duration is used in exponential backoff, such that the
  first wait will be for this duration, the second for 2x, the third for 4x,
  etc. The units of the duration must be given (e.g. "100ms" or "1s").
* `statusTimeout` is a maximum length of time that polling for volume status can
  occur. This serves as a backstop against a stuck request of malfunctioning API
  that never returns.
* `convertUnderscores` is a boolean flag that controls whether the driver will
  automatically convert underscores to dashes during a volume create request.
  GCE does not allow underscores in the volume name, but some container
  orchestrators (e.g. Docker Swarm) automatically prefix volume names
  with a string containing a dash. This flag enables such requests to proceed,
  but with the volume name modified.

### Runtime behavior
* The GCEPD driver enforces the GCE requirements for disk sizing and naming.
  Disks must be created with a minimum size of 10GB. Disk names must adhere to
  the regular expression of `[a-z]([-a-z0-9]*[a-z0-9])?`, which means the first
  character must be a lowercase letter, and all following characters must be a
  dash, lowercase letter, or digit, except the last character, which cannot be a
  dash.
* If the `zone` parameter is not specified in the driver configuration, and a
  request is received to list all volumes that does not specify a zone in the
  InstanceID header, volumes from all zones will be returned.
* By default, all disks will be created with type `pd-ssd`, which creates an SSD
  based disk. If you wish to create disks that are not SSD-based, change the
  default via the driver config, or the type can be changed at creation time by
  using the `Type` field of the create request.

### Activating the Driver
To activate the GCEPD driver please follow the instructions for
[activating storage drivers](../servers/libstorage.md#storage-drivers), using `gcepd` as the
driver name.

### Troubleshooting
* Make sure that the JSON credentials file as specified in the `keyfile`
  configuration parameter is present and accessible, or that you are running in
  a GCE instance created with a Service Account attached. Whether using a
  `keyfile` or the Service Account associated with the GCE instance, the Service
  Account must have the appropriate permissions as described in
  `Configuration Notes`

<a name="gce-persistent-disk-examples"></a>

### Examples
Below is a full `config.yml` that works with GCE

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: gcepd
  server:
    services:
      gcepd:
        driver: gcepd
        gcepd:
          keyfile: /etc/gcekey.json
          tag: rexray
```

### Caveats
* Snapshot and copy functionality is not yet implemented
* Most GCE instances can have up to 64 TB of total persistent disk space
  attached. Shared-core machine types or custom machine types with less than
  3.75 GB of memory are limited to 3 TB of total persistent disk space. Total
  persistent disk space for an instance includes the size of the root persistent
  disk. You can attach up to 16 independent persistent disks to most instances,
  but instances with shared-core machine types or custom machine types with less
  than 3.75 GB of memory are limited to a maximum of 4 persistent disks,
  including the root persistent disk. See
  [GCE Disks](https://cloud.google.com/compute/docs/disks/) docs for more
  details.
* If running libStorage server in a mode where volume mounts will not be
  performed on the same host where libStorage server is running, it should be
  possible to use a Service Account without the `Service Account Actor` role,
  but this has not been tested. Note that if persistent disk mounts are to be
  performed on *any* GCE instances that have a Service Account associated with
  the, the `Service Account Actor` role is required.
