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

 EnvVar | YAML | CLI  | Description
--------|------|------|-------------
`AWS_ACCESS_KEY` | `awsAccessKey` | `--awsAccessKey` | The AWS access key, the public part of the IAM credentials.
`AWS_SECRET_KEY` | `awsSecretKey` | `--awsSecretKey` | The AWS secret key, the private part of the IAM credentials |
`AWS_REGION` | `awsRegion` | `--awsRegion` | The AWS region |

## Activating the Driver
To activate the EC2 driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `ec2` as the driver name.

## Examples
