# Summary
Please enter a summary of the issue here.

# New Feature
This section is for issues that relate to suggested enhancements or other
ideas that may improve REX-Ray. Please provide as much detail as possible
regarding the idea. This issue will then serve as the means to have a
discussion about your idea!

# Bug Reports
This section is for issues that relate to discovered problems or bugs.

## Version
Please paste the output of `rexray version`. For example:

```shell
$ rexray version
REX-Ray
-------
Binary: /usr/bin/rexray
Flavor: client+agent+controller
SemVer: 0.7.0
OsArch: Linux-x86_64
Branch: v0.7.0
Commit: a20a838ca70838a914b632637398824fcb10d0db
Formed: Mon, 23 Jan 2017 10:14:32 EST

libStorage
----------
SemVer: 0.4.0
OsArch: Linux-x86_64
Branch: v0.7.0
Commit: a1103d3f215117f7b9f51dae2b24f852c9c54995
Formed: Mon, 23 Jan 2017 10:14:12 EST
```

## Expected Behavior
Please describe in detail the expected behavior.

## Actual Behavior
Please describe in detail the actual behavior.

## Steps To Reproduce
Please list the steps to reproduce the issue in this section.

1. The first step should always be enabling `debug` logging.
  * Open the file `/etc/rexray/config.yml`
  * Set the log level for both REX-Ray and libStorage to `debug` and enable
HTTP request and response tracing for libStorage:
```yaml
rexray:
  logLevel:        debug
libstorage:
  logging:
    level:         debug
    httpRequests:  true
    httpResponses: true
```
* Please list each step with as much detail as possible.
* The more information gathered up front, the easier it is to solve
the problem.
* Thank you!

## Configuration Files
Please paste any related configuration files, such as `/etc/rexray/config.yml`
in this section. Please use the appropriate formatting when pasting YAML.
content. For example:

```yaml
rexray:
  logLevel:        debug
libstorage:
  logging:
    level:         debug
    httpRequests:  true
    httpResponses: true
  service:         ebs
ebs:
  accessKey:       123456
  secretKey:       abcdef
```

Proper formatting of pasted content is very important as structured data can
sometimes be accidentally recorded incorrectly, affecting the desired outcome.

## Logs
It is very important when filing an issue to include associated logs. There are
two different logs about which to be concerned: the service log (if REX-Ray is
running as a service) and the client log.

### Service Log
The REX-Ray service log file is stored at `/var/log/rexray/rexray.log`. Instead
of pasting the entire log file into this issue, please create a new
[gist](https://gist.github.com/) and paste the log file's contents there.
Please name the file `rexray-service.log` in the gist. The proper extension will
indicate how to format the contents.

### Client Log
The REX-Ray client will emit all of its logs to the console when operating with
debug logging enabled. Simply copy the contents of the console and paste them
into the same gist as above naming the file `rexray-client.log.sh`. The `sh`
extension will cause the contents to be formatted as if they were emitted
to the shell, which they were.
