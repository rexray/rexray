# RBD Driver Testing
This package includes two different kinds of tests for the RBD storage driver:

Test Type | Description
----------|------------
Unit/Integration | The unit/integration tests are provided and executed via the standard Go test pattern, in a file named `rbd_test.go`. These tests are designed to test the storage driver's and executor's functions at a low-level, ensuring, given the proper input, the expected output is received.
Test Execution Plan | The test execution plan operates above the code-level, using a Vagrantfile to deploy a complete implementation of the RBD storage driver in order to run real-world, end-to-end test scenarios.

## Unit/Integration Tests
The unit/integration tests must be executed on a node that has access to a Ceph
cluster. In order to execute the tests either compile the test binary locally or
on the instance. From the root of the libStorage project execute the following:

```bash
GOOS=linux make test-rbd
```

Once the test binary is compiled, if it was built locally, copy it to the node
where testing can occur

Using an SSH session to connect to the node, the tests may now be executed as
root with the following command:

```bash
sudo ./rbd.test
```

An exit code of `0` means the tests completed successfully. If there are errors
then it may be useful to run the tests once more with increased logging:

```bash
sudo LIBSTORAGE_LOGGING_LEVEL=debug ./rbd.test -test.v
```

## Test Execution Plan
In addition to the low-level unit/integration tests, the RBD storage driver
provides a test execution plan automated with Vagrant:

```
vagrant up --provider=virtualbox
```

The above command brings up a Vagrant environment using VirtualBox virtual
machines in order to test the RBD driver. If the command completes successfully
then the environment was brought online without issue and indicates that the
test execution plan succeeded as well.

The following sections outline dependencies, settings, and different execution
scenarios that may be required or useful for using the Vagrantfile.

### Test Plan Dependencies
The following dependencies are required in order to execute the included test
execution plan:

  * [Vagrant](https://www.vagrantup.com/) 1.8.4+
  * [vagrant-hostmanager](https://github.com/devopsgroup-io/vagrant-hostmanager)

Once Vagrant is installed the required plug-ins may be installed with the
following commands:

```bash
vagrant plugin install vagrant-hostmanager
```

**NOTE**: The VMs contain the default Vagrant insecure SSH public key, such that
`vagrant ssh` works by default. However, the `ceph-admin` VM needs to be able to
SSH to the other VMs in order to configure Ceph via `ceph-deploy`. In order to
do this, the Vagrant SSH private key must be in your local SSH agent. The most
typical way to accomplish this on a nix-like machine is by running the command:

```
ssh-add ~/.vagrant.d/insecure_private_key
```

Configuration of the Ceph cluster will not work without this step.

### Test Plan Nodes
The `Vagrantfile` deploys a Ceph cluster and two RBD/rexray clients named:

  * libstorage-rbd-test-server1
  * libstorage-rbd-test-server2
  * libstorage-ebs-test-admin
  * libstorage-ebs-test-client

The "admin" node is identical to client, except it is also use to configure
the Ceph cluster using `ceph-deploy`.

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
to remember to clean up the VirtualBox resources that may have been created
along the way. To do so simply execute the following command:

```bash
vagrant destroy -f
```
