# Summary
Please enter a summary of the issue here.

# New Feature
This section is for issues that relate to suggested enhancements or other
ideas that may improve libStorage. Please provide as much detail as possible
regarding the idea. This issue will then serve as the means to have a
discussion about your idea!

# Bug Reports
This section is for issues that relate to discovered problems or bugs.

## Expected Behavior
Please describe in detail the expected behavior.

## Actual Behavior
Please describe in detail the actual behavior.

## Steps To Reproduce
Please list the steps to reproduce the issue in this section.

1. The first step should always be enabling `debug` logging.
  * Open the file `/etc/libstorage/config.yml` (or the config file for the
application using libStorage)
  * Set the log level for libStorage to `debug` and enable HTTP request and
response tracing for libStorage:
```yaml
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
Please paste any related configuration files, such as
`/etc/libstorage/config.yml` (or the config file for the application using
libStorage in this section. Please use the appropriate formatting when pasting
YAML content. For example:

```yaml
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
two different logs about which to be concerned: the service log (if libStorage
is running as a service) and the client log.

### Service Log
A service that embeds the libStorage service may have its own service log.
Instead of pasting the entire log file into this issue, please create a new
[gist](https://gist.github.com/) and paste the log file's contents there.
Please name the file `libstorage-service.log` in the gist. The proper extension
will indicate how to format the contents.

### Client Log
libStorage clients may emit their logs to the console or there may be an
associated client log file. Please copy the contents of the console and paste
them into the same gist as above naming the file `libstorage-client.log.sh`.
The `sh` extension will cause the contents to be formatted as if they were
emitted to the shell, which they were.
