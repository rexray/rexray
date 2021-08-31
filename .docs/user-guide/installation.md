# Installation

Getting the bits, bit by bit

---

## Overview
There are several different methods available for installing REX-Ray. It
is written in Go, so there are typically no dependencies that must be installed
alongside its single binary file. The manual methods can be extremely simple
through tools like `curl`. You also have the opportunity to perform install
steps individually. Following the manual installs, [configuration](./config.md)
must take place.

Great examples of automation tools, such as `Ansible` and `Puppet`, are also
provided. These approaches automate the entire configuration process.

## Manual Installs
Manual installations are in contrast to batch, automated installations.

Make sure that before installing REX-Ray that you have uninstalled any previous
versions. A `rexray uninstall` can assist with this where appropriate.

Following an installation and configuration, you can use REX-Ray interactively
through commands like `rexray volume`. Noticeably different from this is having
REX-Ray integrate with Container Engines such as Docker. This requires that
you run `rexray start` or relevant service start command like
`systemctl start rexray`.


### Install via curl
The following command will download the most recent, stable build of REX-Ray
and install it to `/usr/bin/rexray` or `/opt/bin/rexray`. On Linux systems
REX-Ray will also be registered as either a SystemD or SystemV service.

```bash
$ curl -sSL https://rexray.io/install | sh -s -- stable
```

### Install a specific version
The following command will emit a list of the available REX-Ray packages:
The curl command can also be used to install a specific version of REX-Ray:

```bash
$ curl -sSL https://rexray.io/install | sh -s -- list
curl -sSL https://rexray.io/install | sh -s -- list unstable
curl -sSL https://rexray.io/install | sh -s -- list staged
curl -sSL https://rexray.io/install | sh -s -- list stable
```

To install the stable `0.9.1` release, use the following command:
```bash
$ curl -sSL https://rexray.io/install | sh -s -- stable 0.9.1
```

### Install a pre-built binary
There are a handful of necessary manual steps to properly install REX-Ray
from pre-built binaries.

1. Download the proper binary. There are also pre-built binaries available for
the various release types.

    Version  | Description
    ---------|------------
    [Unstable](https://dl.bintray.com/rexray/rexray/unstable/latest/) | The most up-to-date, bleeding-edge, and often unstable REX-Ray binaries.
    [Staged](https://dl.bintray.com/rexray/rexray/staged/latest/) | The most up-to-date, release candidate REX-Ray binaries.
    [Stable](https://dl.bintray.com/rexray/rexray/stable/latest/) | The most up-to-date, stable REX-Ray binaries.

2. Uncompress and move the binary to the proper location. Preferably `/usr/bin`
should be where REX-Ray is moved, but this path is not required.
3. Install as a service with `rexray install`. This will register itself
with SystemD or SystemV for proper initialization.

### Build and install from source
It is also easy to build REX-Ray from source:

```bash
$ go get github.com/AVENTER-UG/rexray
```

For more information on how to build REX-Ray please see the
[Build Reference](/dev-guide/build-reference.md).

### Uninstall
Depending on how it was installed, REX-Ray can be installed one of a few ways:

#### RPM
If REX-Ray was installed on a system that uses the RPM package management
system, such as RedHat, CentOS, the following command can be used to uninstall
REX-Ray:

```sh
$ sudo rpm -e rexray
```

#### DEB
If REX-Ray was installed on a system that uses the DEB package management
system, such as Debian, Ubuntu, the following command can be used to uninstall
REX-Ray:

```sh
$ sudo dpkg --remove rexray
```

#### Default
No matter how REX-Ray was installed, the following command will always attempt
to perform an uninstallation using the OS-recommended method:

```sh
$ sudo rexray uninstall
```

## Automated Installs
Because REX-Ray is simple to install using the `curl` script, installation
using configuration management tools is relatively easy as well. However,
there are a few areas that may prove to be tricky, such as writing the
configuration file.

This section provides examples of automated installations using common
configuration management and orchestration tools.

### Ansible
With Ansible, installing the latest REX-Ray binaries can be accomplished by
including the `rexray.rexray` role from Ansible Galaxy.  The role accepts
all the necessary variables to properly fill out your `config.yml` file.

Install the role from Galaxy:

```sh
$ ansible-galaxy install rexray.rexray
```

Example playbook for installing REX-Ray on GCE Docker hosts:

```yaml
- hosts: gce_docker_hosts
  roles:
  - { role: rexray.rexray,
      rexray_service: true,
      rexray_storage_drivers: [gce],
      rexray_gce_keyfile: "/opt/gce_keyfile" }
```

Run the playbook:

```sh
$ ansible-playbook -i <inventory> playbook.yml
```

### AWS CloudFormation
With CloudFormation, the installation of the latest Docker and REX-Ray binaries
can be passed to the orchestrator using the 'UserData' property in a
CloudFormation template. While the payload could also be provided as raw user
data via the AWS GUI, it would not sustain scalable automation.

```json
"Properties": {
  "UserData": {
    "Fn::Base64": {
      "Fn::Join": ["", [
        "#!/bin/bash -xe\n",
        "apt-get update\n",
        "apt-get -y install python-setuptools\n",
        "easy_install https://s3.amazonaws.com/cloudformation-examples/aws-cfn-bootstrap-latest.tar.gz\n",
        "ln -s /usr/local/lib/python2.7/dist-packages/aws_cfn_bootstrap-1.4-py2.7.egg/init/ubuntu/cfn-hup /etc/init.d/cfn-hup\n",
        "chmod +x /etc/init.d/cfn-hup\n",
        "update-rc.d cfn-hup defaults\n ",
        "service cfn-hup start\n",
        "/usr/local/bin/cfn-init --stack ", {
          "Ref": "AWS::StackName"
        }, " --resource RexrayInstance ", " --configsets InstallAndRun --region ", {
          "Ref": "AWS::Region"
        }, "\n",

        "# Install the latest Docker..\n",
        "/usr/bin/curl -o /tmp/install-docker.sh https://get.docker.com/\n",
        "chmod +x /tmp/install-docker.sh\n",
        "/tmp/install-docker.sh\n",

        "# add the ubuntu user to the docker group..\n",
        "/usr/sbin/usermod -G docker ubuntu\n",

        "# Install the latest REX-ray\n",
        "/usr/bin/curl -ssL -o /tmp/install-rexray.sh https://rexray.io/install\n",
        "chmod +x /tmp/install-rexray.sh\n",
        "/tmp/install-rexray.sh\n",
        "chgrp docker /etc/rexray/config.yml\n",
        "reboot\n"
      ]]
    }
  }
}
```

### Docker Machine (VirtualBox)
SSH can be used to remotely deploy REX-Ray to a Docker Machine. While the
following example used VirtualBox as the underlying storage platform, the
provided `config.yml` file *could* be modified to use any of the supported
drivers.

1. SSH into the Docker machine and install REX-Ray.

        $ docker-machine ssh testing1 \
        "curl -sSL https://rexray.io/install | sh"

2. Install the udev extras package. This step is only required for versions of
   boot2docker older than 1.10.

        $ docker-machine ssh testing1 \
        "wget http://tinycorelinux.net/6.x/x86_64/tcz/udev-extra.tcz \
        && tce-load -i udev-extra.tcz && sudo udevadm trigger"

3. Create a basic REX-Ray configuration file inside the Docker machine.

    **Note**: It is recommended to replace the `volumePath` parameter with the
    local path VirtualBox uses to store its virtual media disk files.

        $ docker-machine ssh testing1 \
            "sudo tee -a /etc/rexray/config.yml << EOF
            libstorage:
              integration:
                volume:
                  operations:
                    mount:
                      preempt: false
            virtualbox:
              volumePath: $HOME/VirtualBox/Volumes
            "

4. Finally, start the REX-Ray service inside the Docker machine.

        $ docker-machine ssh testing1 "sudo rexray start"

### OpenStack Heat
Using OpenStack Heat, in the HOT template format (yaml):

```yaml
resources:
  my_server:
    type: OS::Nova::Server
    properties:
      user_data_format: RAW
      user_data:
        str_replace:
          template: |
            #!/bin/bash -v
            /usr/bin/curl -o /tmp/install-docker.sh https://get.docker.com
            chmod +x /tmp/install-docker.sh
            /tmp/install-docker.sh
            /usr/sbin/usermod -G docker ubuntu
            /usr/bin/curl -ssL -o /tmp/install-rexray.sh https://rexray.io/install
            chmod +x /tmp/install-rexray.sh
            /tmp/install-rexray.sh
            chgrp docker /etc/rexray/config.yml
          params:
            dummy: ""
```

### Vagrant
Using Vagrant is a great option to deploy pre-configured REX-Ray nodes,
including Docker, using the VirtualBox driver. All volume requests are handled
using VirtualBox's Virtual Media.

A Vagrant environment and instructions using it are provided
[here](https://github.com/codedellemc/vagrant/tree/master/rexray).
