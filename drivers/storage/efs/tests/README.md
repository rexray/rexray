# EFS driver testing

EFS driver requires valid AWS environment in which tests can run. The
requirements are:

* **VPC**: Private network where EC2 instance and EFS instance can be launched

* **Subnet**: In VPC it requires at least one valid Subnet for EC2 instance and
  EFS

* **EC2**: Any EC2 instance that has permission to run EFS plugin. For list of
  IAM permissions see user documentation.

For automated testing there are couple script available that will spin up
while AWS environment by using CloudFormation service.

* `./test-env-up.sh [key-name]` - should be used to launch whole AWS environment
  required for EFS storage driver testing. `[key-name]` is a valid AWS key that
  will be used to launch EC2 instance.

* `./test-run.sh [path-to-binary]` - to copy test binary and run it on EC2 instance.

* `./test-env-down.sh` - to tear down AWS environment infrastructure and clean
  up all resources.

**NOTE**: For configuration details see libstorage user guide.