# EC2 Vagrantfile
This directory includes a Vagrantfile that can be used to bring a libStorage
server/client installation online using REX-Ray and EC2 instances. For example:

```
vagrant up --provider=aws --no-parallel
```

The following sections outline dependencies, settings, and different execution
scenarios that may be required or useful for using the Vagrantfile.

### Dependencies
The following dependencies are required in order to bring the EC2 Vagrant
enviornment online:

  * [Vagrant](https://www.vagrantup.com/) 1.8.5+
  * [vagrant-aws](https://github.com/mitchellh/vagrant-aws)
  * [vagrant-hostmanager](https://github.com/devopsgroup-io/vagrant-hostmanager)

Once Vagrant is installed the required plug-ins may be installed with the
following commands:

```bash
vagrant plugin install vagrant-aws
vagrant plugin install vagrant-hostmanager
```

### Settings
The following environment variables may be used to configure the `Vagrantfile`.

Environment Variable | Description | Required | Default
---------------------|-------------|:--------:|--------
`AWS_AKEY`           | The AWS access key ID. | ✓ |
`AWS_SKEY`           | The AWS secret key. | ✓ |
`AWS_KPNM`           | The name of the AWS key pair for instances. | ✓ |
`AWS_AMI`            | The AMI to use. Defaults to Amazon Linux PV x64 ([AMIs by region](https://aws.amazon.com/amazon-linux-ami/)) | | ami-de347abe
`AWS_REGN`           | The AWS region. | | us-west-1
`AWS_ZONE`           | The AWS availability zone. | | a
`AWS_SSHK`           | Local SSH key used to access AWS instances. | ✓ |

### Nodes
The `Vagrantfile` which deploys a libStorage server and two clients onto EC2
instances named:

  * libstorage-server
  * libstorage-client0
  * libstorage-client1

### Test Plan Boot Order
The order in which Vagrant brings the nodes online can vary:

```bash
vagrant up --provider=aws
```

The above command will bring all three nodes -- the server and both clients --
online in parallel since the Vagrant AWS plug-in is configured to support
the parallel option. However, the Vagrantfile will fail if the server
is not brought online first. That is why when bringing the Vagrant nodes online
all at once, the following command should be used:

```bash
vagrant up --provider=aws --no-parallel
```

The above command will bring the nodes online in the following order:

  1. libstorage-server
  2. libstorage-client0
  3. libstorage-client1

However, another option for starting the environment is to bring up the server
first and then the clients in parallel. Doing so can duplicate a scenario
where two clients are contending for storage resources:

```bash
vagrant up --provider=aws libstorage-server
vagrant up --provider=aws '/.*client.*/'
```

### Scripts
This package includes several test scripts that act as simple ways to execute
random pieces of logic for a node after it is brought online:

  * `server.sh`
  * `client0.sh`
  * `client1.sh`

The above files are copied to their respective instances and executed
as soon as the instance is online. That means the `server-tests.sh` script is
executed before the client nodes are even online since the server node is
brought online first.

### Cleanup
It's important to remember to clean up any EC2, EBS, & EFS resources that may
have been created along the way. To do so simply execute the following command:

```bash
vagrant destroy -f
```
