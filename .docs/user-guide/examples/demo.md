# Demo

Easy as 1, 2, 3...

---

## Two-Node Client/Server

This demo consists of a two-node deployment with the first node configured as a
REX-Ray/libStorage server and the second node as merely a client. Both nodes
have Docker installed and configured to leverage REX-Ray for persistent storage.

The below example does have a few requirements:

 * VirtualBox 5.0+
 * Vagrant 1.8+
 * Ruby 2.0+

### Start REX-Ray Vagrant Environment
Before bringing the Vagrant environment online, please ensure it is accomplished
in a clean directory

```sh
$ cd $(mktemp -d)
```

Inside the newly created, temporary directory, download the REX-Ray
[Vagrantfile](https://github.com/AVENTER-UG/rexray/master/Vagrantfile):

```sh
$ curl -fsSLO https://raw.githubusercontent.com/rexray/rexray/master/Vagrantfile
```

Now it is time to bring the REX-Ray environment online:

!!! note "note"

    The next step could potentially open up the system on which the command is
    executed to security vulnerabilities. The Vagrantfile brings the VirtualBox
    web service online if it is not already running. However, in the name of
    simplicity the Vagrantfile also disables the web server's authentication
    module. Please do not disable authentication for the VirtualBox web server
    if this example is being executed on an open network or without some type of
    firewall in place.

```sh
$ vagrant up
```

The above command should result in output similar to [this
Gist](https://gist.github.com/akutz/13fc3b2237ea2c295a25c2e367e6bd8f).

Once the command has been completed successfully there will be two VMs online
named `node0` and `node1`. Both nodes are running Docker and REX-Ray with
`node0` configured to act as a libStorage server.

Now that the environment is online it is time to showcase Docker leveraging REX-
Ray to create persistent storage as well as illustrating REX-Ray's distributed
deployment capabilities.

### Node 0
First, SSH into `node0`

```sh
$ vagrant ssh node0
```

From `node0` use Docker with REX-Ray to create a new volume named
`hellopersistence`:

```sh
vagrant@node0:~$ docker volume create --driver rexray --opt size=1 \
                 --name hellopersistence
```

After the volume is created, mount it to the host and container using the
`--volume-driver` and `-v` flag in the `docker run` command:

```sh
vagrant@node0:~$ docker run -tid --volume-driver=rexray \
                 -v hellopersistence:/mystore \
                 --name temp01 busybox
```

Create a new file named `myfile` on the file system backed by the persistent
volume using `docker exec`:

```sh
vagrant@node0:~$ docker exec temp01 touch /mystore/myfile
```

Verify the file was successfully created by listing the contents of the
persistent volume:

```sh
vagrant@node0:~$ docker exec temp01 ls /mystore
```

Remove the container that was used to write the data to the persistent volume:

```sh
vagrant@node0:~$ docker rm -f temp01
```

Finally, exit the SSH session to `node0`:

```sh
vagrant@node0:~$ exit
```

### Node 1
It's time to connect to `node1` and use the volume `hellopersistence` that was
created in the previous section from `node0`.

!!! note "note"

    While `node1` runs both the Docker and REX-Ray services like `node0`, the
    REX-Ray service on `node1` in no way understands or is configured for the
    VirtualBox storage driver. All interactions with the VirtualBox web service
    occurs via `node0`'s libStorage server with which `node1` communicates.

Use the vagrant command to SSH into `node1`:

```sh
$ vagrant ssh node1
```

Next, create a new container that mounts the existing volume,
`hellopersistence`:

```sh
vagrant@node1:~$ docker run -tid --volume-driver=rexray \
                 -v hellopersistence:/mystore \
                 --name temp01 busybox
```

The next command validates the file `myfile` created from `node0` in the
previous section has persisted inside the volume across machines:

```sh
vagrant@node1:~$ docker exec temp01 ls /mystore
```

Finally, exit the SSH session to `node1`:

```sh
vagrant@node1:~$ exit
```

### Cleaning Up
Be sure to kill the VirtualBox web server with a quick `killall vboxwebsrv` and
to tear down the Vagrant environment with `vagrant destroy`. Omitting these
commands will leave the web service and REX-Ray Vagrant nodes online and consume
additional system resources.

### Congratulations
REX-Ray has been used to provide persistence for stateless containers! Examples
using MongoDB, Postgres, and more with persistent storage can be found at
[Application Examples](./apps.md) or within the [{code} Labs
repo](https://github.com/codedellemc/labs).
