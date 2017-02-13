# GCE Persistent Disk Driver Testing
This package includes two different kinds of tests for the GCE Persistent Disk
storage driver:

Test Type | Description
----------|------------
Unit/Integration | The unit/integration tests are provided and executed via the standard Go test pattern, in a file named `gce_test.go`. These tests are designed to test the storage driver's and executor's functions at a low-level, ensuring, given the proper input, the expected output is received.
Test Execution Plan | The test execution plan operates above the code-level, using a Vagrantfile to deploy a complete implementation of the GCE storage driver in order to run real-world, end-to-end test scenarios.

## Unit/Integration Tests
The unit/integration tests must be executed on a node that is hosted within GCE.
In order to execute the tests either compile the test binary locally or
on the instance. From the root of the libStorage project execute the following:

```bash
GOOS=linux make test-gcepd
```

Once the test binary is compiled, if it was built locally, copy it to the GCE
instance. You will also need to copy the JSON file with your service account
credentials.

Using an SSH session to connect to the GCE instance, please export the required
GCE credentials used by the GCE storage driver:

```bash
export GCE_KEYFILE=/etc/gcekey.json
```

The tests may now be executed with the following command:

```bash
sudo ./gcepd.test
```

An exit code of `0` means the tests completed successfully. If there are errors
then it may be useful to run the tests once more with increased logging:

```bash
sudo LIBSTORAGE_LOGGING_LEVEL=debug ./gcepd.test -test.v
```

## Test Execution Plan
In addition to the low-level unit/integration tests, the GCE storage driver
provides a test execution plan automated with Vagrant:

```
vagrant up --provider=google --no-parallel
```

The above command brings up a Vagrant environment using GCE instances in order
to test the GCE driver. If the command completes successfully then the
environment was brought online without issue and indicates that the test
execution plan succeeded as well.

The *--no-parallel* flag is important, as the tests are written such that tests
on one node are supposed to run and finished before the next set of tests.

The following sections outline dependencies, settings, and different execution
scenarios that may be required or useful for using the Vagrantfile.

### Test Plan Dependencies
The following dependencies are required in order to execute the included test
execution plan:

  * [Vagrant](https://www.vagrantup.com/) 1.8.4+
  * [vagrant-google](https://github.com/mitchellh/vagrant-google)

Once Vagrant is installed the required plug-ins may be installed with the
following commands:

```bash
vagrant plugin install vagrant-google
```

### Test Plan Settings
The following environment variables may be used to configure the `Vagrantfile`.

Environment Variable | Description | Required | Default
---------------------|-------------|:--------:|--------
`GCE_PROJECT_ID`     | The GCE Project ID | ✓ |
`GCE_CLIENT_EMAIL`   | The email address of the service account holder | ✓ |
`GCE_JSON_KEY`       | The location of the GCE credentials file on your local machine | ✓ |
`GCE_MACHINE_TYPE`   | The GCE machine type to use | | n1-standard-1
`GCE_IMAGE`          | The GCE disk image to boot from | | centos-7-v20170110
`GCE_ZONE`           | The GCE zone to launch instance within | | us-west1-b
`REMOTE_USER`        | The account name to SSH to the GCE node as | | *The local user* (.e.g. `whoami`)
`REMOTE_SSH_KEY`     | The location of the private SSH key to use for SSH into the GCE node | | ~/.ssh/id_rsa

### Test Plan Nodes
The `Vagrantfile` deploys two GCE/rexray clients with Docker named:

  * libstorage-gce-test0
  * libstorage-gce-test1


### Test Plan Scripts
This package includes test scripts that execute the test plan:

  * `client0-tests.sh`
  * `client1-tests.sh`

The above files are copied to their respective instances and executed
as soon as the instance is online.

### Test Plan Cleanup
Once the test plan has been executed, successfully or otherwise, it's important
to remember to clean up the GCE resources that may have been created along the
way. To do so simply execute the following command:

```bash
vagrant destroy -f
```
