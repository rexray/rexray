#Automation

Push Button, Receive REX-Ray...

---

## Overview
Because REX-Ray is simple to install using the `curl` script, installing using
configuration management tools is relatively easy as well. There are a few things
that can be tricky though - most notably writing out the configuration file.

We have provided some examples with common configuration management and
orchestration tools below.  Optionally, Docker is also listed in some examples.

<br>
## Ansible
With Ansible, installing the latest REX-Ray binaries can be accomplished by
including the `codenrhoden.rexray` role from Ansible Galaxy.  The role accepts
all the necessary variables to properly fill out your `config.yml` file.

Install the role from Galaxy:

```shell
ansible-galaxy install codenrhoden.rexray
```

Example playbook for installing REX-Ray on GCE Docker hosts:

```yaml
- hosts: gce_docker_hosts
  roles:
  - { role: codenrhoden.rexray,
      rexray_service: true,
      rexray_storage_drivers: [gce],
      rexray_gce_keyfile: "/opt/gce_keyfile" }
```

Run the playbook:

```shell
ansible-playbook -i <inventory> playbook.yml
```

<br>
## AWS CloudFormation
With CloudFormation, the installation of the latest Docker and REX-Ray binaries
can be passed to the orchestrator using the 'UserData' property in a
CloudFormation template. Obviously this could also be passed in as raw userdata
from the AWS GUI... but that wouldn't really be automating things, now would it?

```json
      "Properties": {
        "UserData"       : { "Fn::Base64" : { "Fn::Join" : ["", [
             "#!/bin/bash -xe\n",
             "apt-get update\n",
             "apt-get -y install python-setuptools\n",
             "easy_install https://s3.amazonaws.com/cloudformation-examples/aws-cfn-bootstrap-latest.tar.gz\n",
             "ln -s /usr/local/lib/python2.7/dist-packages/aws_cfn_bootstrap-1.4-py2.7.egg/init/ubuntu/cfn-hup /etc/init.d/cfn-hup\n",
             "chmod +x /etc/init.d/cfn-hup\n",
             "update-rc.d cfn-hup defaults\n ",
             "service cfn-hup start\n",
             "/usr/local/bin/cfn-init --stack ",{ "Ref":"AWS::StackName" }," --resource RexrayInstance "," --configsets InstallAndRun --region ",{"Ref":"AWS::Region"},"\n",

             "# Install the latest Docker..\n",
             "/usr/bin/curl -o /tmp/install-docker.sh https://get.docker.com/\n",
             "chmod +x /tmp/install-docker.sh\n",
             "/tmp/install-docker.sh\n",

             "# add the ubuntu user to the docker group..\n",
             "/usr/sbin/usermod -G docker ubuntu\n",

             "# Install the latest REX-ray\n",
             "/usr/bin/curl -ssL -o /tmp/install-rexray.sh https://dl.bintray.com/emccode/rexray/install\n",
             "chmod +x /tmp/install-rexray.sh\n",
             "/tmp/install-rexray.sh\n",
             "chgrp docker /etc/rexray/config.yml\n",
             "reboot\n"
        ]]}}        
      }
    },
```

<br>
## CFEngine
ToDo

<br>
## Chef
ToDo

<br>
## OpenStack Heat
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
                /usr/bin/curl -ssL -o /tmp/install-rexray.sh https://dl.bintray.com/emccode/rexray/install
                chmod +x /tmp/install-rexray.sh
                /tmp/install-rexray.sh
                chgrp docker /etc/rexray/config.yml
              params:
                dummy: ""
```

<br>
## Puppet
ToDo

<br>
## Salt
ToDo

<br>
## Vagrant
Using Vagrant is a great option to deploy pre-configured REX-Ray nodes including
Docker using the VirtualBox driver.  All volume requests are handled using
VirtualBox's Virtual Media.

[here](https://github.com/emccode/vagrant/tree/master/rexray).
