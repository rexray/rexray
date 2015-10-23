# Amazon EC2

Simplifying storage with scalable compute capacity in the cloud

---

## Overview
The Amazon EC2 driver registers a storage driver named `ec2` with the `REX-Ray`
driver manager and is used to connect and manage storage on EC2 instances. The
EC2 driver is made possible by the
[goamz project](https://github.com/mitchellh/goamz).

## Configuration Options
The following are the configuration options for the `ec2` storage driver.

 EnvVar | YAML | CLI  
--------|------|------
`AWS_ACCESS_KEY` | `awsAccessKey` | `--awsAccessKey`
`AWS_SECRET_KEY` | `awsSecretKey` | `--awsSecretKey`
`AWS_REGION` | `awsRegion` | `--awsRegion`

## Activating the Driver
To activate the EC2 driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `ec2` as the driver name.

## Examples
