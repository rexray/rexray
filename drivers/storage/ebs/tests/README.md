# EBS Driver Testing
This package includes two different kinds of tests for the EBS storage driver:

Test Type | Description
----------|------------
Unit/Integration | The unit/integration tests are provided and executed via the standard Go test pattern, in a file named `ebs_test.go`. These tests are designed to test the storage driver's and executor's functions at a low-level, ensuring, given the proper input, the expected output is received.
Test Execution Plan | The test execution plan operates above the code-level, using a Vagrantfile to deploy a complete implementation of the EBS storage driver in order to run real-world, end-to-end test scenarios.

## Unit/Integration Tests
The unit/integration tests must be executed on an EC2 instance that has access
to EBS. In order to execute the tests either compile the test binary locally or
on the instance. From the root of the libStorage project execute the following:

```bash
make deps && make ./drivers/storage/ebs/tests/ebs.test
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
./ebs.test
```

An exit code of `0` means the tests completed successfully. If there are errors
then it may be useful to run the tests once more with increased logging:

```bash
LIBSTORAGE_LOGGING_LEVEL=debug ./ebs.test -test.v
```

## Test Execution Plan
In addition to the low-level unit/integration tests, the EBS storage driver
provides a test execution plan automated with Vagrant:

```
vagrant up --provider=aws --no-parallel
```

The above command brings up a Vagrant environment using EC2 instances in order
to test the EBS driver. If the command completes successfully then the
environment was brought online without issue and indicates that the test
execution plan succeeded as well.

The following sections outline dependencies, settings, and different execution
scenarios that may be required or useful for using the Vagrantfile.

### Test Plan Dependencies
The following dependencies are required in order to execute the included test
execution plan:

  * [Vagrant](https://www.vagrantup.com/) 1.8.5+
  * [vagrant-aws](https://github.com/mitchellh/vagrant-aws)
  * [vagrant-hostmanager](https://github.com/devopsgroup-io/vagrant-hostmanager)

Once Vagrant is installed the required plug-ins may be installed with the
following commands:

```bash
vagrant plugin install vagrant-aws
vagrant plugin install vagrant-hostmanager
```

### Test Plan Settings
The following environment variables may be used to configure the `Vagrantfile`.
Please note that use of the default AWS environment variables was avoided. This
is to refrain from affecting the unit/integration tests as the two may make
use of the default AWS environment variables.

Environment Variable | Description | Required | Default
---------------------|-------------|:--------:|--------
`AWS_AKEY`           | The AWS access key ID. | ✓ |
`AWS_SKEY`           | The AWS secret key. | ✓ |
`AWS_KPNM`           | The name of the AWS key pair for instances. | ✓ |
`AWS_AMI`            | The AMI to use. Defaults to Amazon Linux PV x64 ([AMIs by region](https://aws.amazon.com/amazon-linux-ami/)) | | ami-de347abe
`AWS_REGN`           | The AWS region. | | us-west-1
`AWS_ZONE`           | The AWS availability zone. | | a
`AWS_SSHK`           | Local SSH key used to access AWS instances. | ✓ |

### Test Plan Nodes
The `Vagrantfile` which deploys a libStorage server and two clients onto EC2
instances named:

  * libstorage-ebs-test-server
  * libstorage-ebs-test-client0
  * libstorage-ebs-test-client1

### Test Plan Boot Order
The order in which Vagrant brings the nodes online can vary:

```bash
vagrant up --provider=aws
```

The above command will bring all three nodes -- the server and both clients --
online in parallel since the Vagrant AWS plug-in is configured to support
the parallel option. However, the test execution plan will fail if the server
is not brought online first. That is why when bringing the Vagrant nodes online
all at once, the following command should be used:

```bash
vagrant up --provider=aws --no-parallel
```

The above command will bring the nodes online in the following order:

  1. libstorage-ebs-test-server
  2. libstorage-ebs-test-client0
  3. libstorage-ebs-test-client1

However, another option for executing the test plan is to bring up the server
first and then the clients in parallel. Doing so can duplicate a scenario
where two clients are contending for storage resources:

```bash
vagrant up --provider=aws libstorage-ebs-test-server
vagrant up --provider=aws '/.*client.*/'
```

### Test Plan Scripts
This package includes several test scripts that act as the primary executors
of the test plan:

  * `server-tests.sh`
  * `client0-tests.sh`
  * `client1-tests.sh`

The above files are copied to their respective instances and executed
as soon as the instance is online. That means the `server-tests.sh` script is
executed before the client nodes are even online since the server node is
brought online first.

### Test Plan Cleanup
Once the test plan has been executed, successfully or otherwise, it's important
to remember to clean up the EC2 and EBS resources that may have been created
along the way. To do so simply execute the following command:

```bash
vagrant destroy -f
```
