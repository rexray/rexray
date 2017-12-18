# DigitalOcean

Block Storage

---

## Overview
Thanks to the efforts of our tremendous community, libStorage also has built-in
support for DigitalOcean!

<a name="digitalocean-block-storage"></a>
<a name="dobs"></a>

## DO Block Storage
The DigitalOcean Block Storage (DOBS) driver registers a driver named `dobs`
with the libStorage service registry and is used to attach and mount
DigitalOcean block storage devices to DigitalOcean instances.

### Requirements
The DigitalOcean block storage driver has the following requirements:

* Valid DigitalOcean account
* Valid DigitalOcean [access token](https://goo.gl/iKoAec)

### Configuration
The following is an example with all possible fields configured. For a running
example see the [Examples](#dobs-examples) section.

```yaml
dobs:
  token:  123456
  region: nyc1
  statusMaxAttempts: 10
  statusInitialDelay: 100ms
  statusTimeout: 2m
  convertUnderscores: false
```

#### Configuration Notes
- The `token` contains your DigitalOcean [access token](https://goo.gl/iKoAec)
- `region` specifies the DigitalOcean region where volumes should be created
- `statusMaxAttempts` is the number of times the status of a volume will be
  queried before giving up when waiting on a status change
- `statusInitialDelay` specifies a time duration used to wait when polling
  volume status. This duration is used in exponential backoff, such that the
  first wait will be for this duration, the second for 2x, the third for 4x,
  etc. The units of the duration must be given (e.g. "100ms" or "1s").
- `statusTimeout` is a maximum length of time that polling for volume status can
  occur. This serves as a backstop against a stuck request of malfunctioning API
  that never returns.
- `convertUnderscores` is a boolean flag that controls whether the driver will
  automatically convert underscores to dashes during a volume create request.
  Digital Ocean does not allow underscores in the volume name, but some
  container orchestrators (e.g. Docker Swarm) automatically prefix volume names
  with a string containing a dash. This flag enables such requests to proceed,
  but with the volume name modified.

!!! note
    The DigitalOcean service currently only supports block storage volumes in
    specific regions. Make sure to use a [supported region](https://www.digitalocean.com/community/tutorials/how-to-use-block-storage-on-digitalocean#what-is-digitalocean-block-storage).

    The standard environment variable for the DigitalOcean access token is
    `DIGITALOCEAN_ACCESS_TOKEN`. However, the environment variable mapped to
    this driver's `dobs.token` property is `DOBS_TOKEN`. This choice was made
    to ensure that the driver must be explicitly configured for access instead
    of detecting a default token that may not be intended for the driver.

<a name="dobs-examples"></a>

### Examples
Below is a full `config.yml` that works with DOBS

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: dobs
  server:
    services:
      dobs:
        driver: dobs
        dobs:
          token: 123456
          region: nyc1
```
