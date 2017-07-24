# FittedCloud EBS Driver Testing

The unit/integration tests are provided and executed via the standard Go test
pattern, in a file named `fittedcloud_test.go`.  These tests are designed to
test the storage driver's and executor's functions at a low-level, ensuring,
given the proper input, the expected output is received.

## Unit/Integration Tests
The unit/integration tests must be executed on an EC2 instance that has access
to EBS. It also requires the FittedCloud Agent to be installed in order to
optimize EBS storage by thin provisioning.

Download and install the latest FittedCloud Agent software.

```bash
curl -k 'https://customer.fittedcloud.com/downloadsoftware?ver=latest' -o fcagent.run
sudo bash ./fcagent.run -- -o S -m -d <User ID>
```

The `<User ID>` is your FittedCloud account ID, which can be found on the
settings page once you registered and logged in to the customer.fittedcoud.com
web site.

In order to execute the tests either compile the test binary locally or
on the instance. From the root of the libStorage project execute the following:

```bash
BUILD_TAGS="gofig pflag libstorage_integration_docker libstorage_storage_driver libstorage_storage_executor libstorage_storage_driver_fittedcloud libstorage_storage_executor_fittedcloud" make build-tests
```

Once the test binary is compiled, if it was built locally, copy it to the EC2
instance.

Using an SSH session to connect to the EC2 instance, please export the required
AWS credentials used by the EBS storage driver:

```bash
export EBS_ACCESSKEY=VAL
export EBS_SECRETKEY=VAL
```

The tests may now be executed with the following command:

```bash
./fittedcloud.test
```

An exit code of `0` means the tests completed successfully. If there are errors
then it may be useful to run the tests once more with increased logging:

```bash
LIBSTORAGE_LOGGING_LEVEL=debug ./fittedcloud.test -test.v
```
