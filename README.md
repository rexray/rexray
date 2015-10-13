# REX-Ray [![Build Status](http://travis-ci.org/emccode/rexray.svg?branch=master)](https://travis-ci.org/emccode/rexray) [![Coverage Status](http://coveralls.io/repos/emccode/rexray/badge.svg?branch=master&service=github)](https://coveralls.io/github/emccode/rexray?branch=master) [ ![Download](http://api.bintray.com/packages/emccode/rexray/stable/images/download.svg) ](https://dl.bintray.com/emccode/rexray/stable/latest/)
```REX-Ray``` provides visibility and management of external/underlying storage via guest storage introspection. Available as a Go package, CLI tool, and Linux service, and with built-in third-party support for tools such as ```Docker```, ```REX-Ray``` is easily integrated into any workflow. For example, here's how to list storage for a guest hosted on Amazon Web Services (AWS) with ```REX-Ray```:

```bash
[0]akutz@pax:~$ export REXRAY_STORAGEDRIVERS=ec2
[0]akutz@pax:~$ export AWS_ACCESS_KEY=access_key
[0]akutz@pax:~$ export AWS_SECRET_KEY=secret_key
[0]akutz@pax:~$ rexray volume get

- providername: ec2
  instanceid: i-695bb6ab
  volumeid: vol-dedbadc3
  devicename: /dev/sda1
  region: us-west-1
  status: attached
- providername: ec2
  instanceid: i-695bb6ab
  volumeid: vol-04c4b219
  devicename: /dev/xvdb
  region: us-west-1
  status: attached

[0]akutz@pax:~$
```

# Installation
Installing `REX-Ray` couldn't be easier.

```bash
curl -sSL https://dl.bintray.com/emccode/rexray/install | sh -
```

On Linux systems the above command will download `REX-Ray` and install it at `/usr/bin/rexray` and register it as a SystemV or SystemD service depending on what the Linux distribution supports. On Darwin (OS X) systems the binary is installed at `/usr/bin/rexray` sans service registration.

# Downloading
There are also pre-built binaries at the following locations:

Repository | Version | Description
---------- | ------- | -----------
[unstable](https://dl.bintray.com/emccode/rexray/unstable) | [ ![Download](https://api.bintray.com/packages/emccode/rexray/unstable/images/download.svg) ](https://dl.bintray.com/emccode/rexray/unstable/latest/) | The most up-to-date, bleeding-edge, and often unstable REX-Ray binaries.
[staged](https://dl.bintray.com/emccode/rexray/staged)   | [ ![Download](https://api.bintray.com/packages/emccode/rexray/staged/images/download.svg) ](https://dl.bintray.com/emccode/rexray/staged/latest/) | The most up-to-date, release candidate REX-Ray binaries.
[stable](https://dl.bintray.com/emccode/rexray/stable)   | [ ![Download](https://api.bintray.com/packages/emccode/rexray/stable/images/download.svg) ](https://dl.bintray.com/emccode/rexray/stable/latest/) | The most up-to-date, stable REX-Ray binaries.

# Building
`REX-Ray` is also fairly simple to build from source, especially if you have `Docker` installed:

```bash
SRC=$(mktemp -d 2> /dev/null || mktemp -d -t rexray 2> /dev/null) && cd $SRC && docker run --rm -it -v $SRC:/usr/src/rexray -w /usr/src/rexray golang:1.5.1 bash -c "git clone https://github.com/emccode/rexray.git . && make build-all”
```

If you'd prefer to not use `Docker` to build `REX-Ray` then all you need is Go 1.5:

```bash
# clone the rexray repo
git clone https://github.com/emccode/rexray.git

# change directories into the freshly-cloned repo
cd rexray

# build rexray
make build-all
```

After either of the above methods for building `REX-Ray` there should be a `.bin` directory in the
current directory, and inside `.bin` will be binaries for Linux-i386, Linux-x86-64,
and Darwin-x86-64.

```bash
[0]akutz@poppy:tmp.SJxsykQwp7$ ls .bin/*/rexray
-rwxr-xr-x. 1 root 14M Sep 17 10:36 .bin/Darwin-x86_64/rexray*
-rwxr-xr-x. 1 root 12M Sep 17 10:36 .bin/Linux-i386/rexray*
-rwxr-xr-x. 1 root 14M Sep 17 10:36 .bin/Linux-x86_64/rexray*
```

# State
We have a first release available that support all of the following capabilities!  

## Current Storage Drivers
- Block
    - Cloud infrastructure
        - EC2
        - OpenStack
    - Software-Defined with Kernel Module
        - ScaleIO
    - iSCSI
        - XtremIO (with/without Multipath and Device-Mapper)
- NAS

## Docker Integration
```REX-Ray``` can be leveraged by ```Docker``` (1.7+) as a ```VolumeDriver```. ```Docker``` can connect to one or more of these drivers by specifying different ```--host``` flags of ```unix:///run/docker/plugins/name.sock``` or ```tcp://127.0.0.1:port``` when executed. That means getting ```Docker``` to leverage ```REX-Ray``` is as easy as:

```bash
# Start the REX-Ray service. The Docker volume driver is enabled
# by default.
sudo rexray service start

# Create a new container with a volume that leverages the REX-Ray driver
docker run --volume-driver=rexray -v volumename:/pathtomount container
```

Additionally, the [Dogged](https://github.com/emccode/dogged) repo maintains efforts for the EMC {code} team relating to embedding ```REX-Ray``` inside of Container Engines such as Docker. Here you will find ```REX-Ray``` enabling Docker to manage its own storage via Container Data Volumes.

# Building
// TODO

This might currently require upstream additions for the Goamz package to github.com/clintonskitson/goamz at the snapcopy branch.

```bash
docker run --rm -it -e GO15VENDOREXPERIMENT=1 -v $GOPATH:/go -w /go/src/github.com/emccode/rexray/ golang:1.5 make install
```

# Environment Variables
The primary method used to configure the ```REX-Ray``` library and influence its behavior is through he use of environment variables.

## General
Name | Description | Optional
-----|-------------|-----------
```REXRAY_DEBUG``` | Set to ```true``` or ```TRUE``` to enable debug messages | yes
```REXRAY_STORAGEDRIVERS``` | A comma-separated list of storage driver names which instructs ```REX-ray``` to only do checks using the specified drivers | yes
```REXRAY_OSDRIVERS``` | A comma-separated list of OS driver names which instructs ```REX-ray``` to only do checks using the specified drivers | yes
```REXRAY_VOLUMEDRIVERS``` | A comma-separated list of volume driver names which instructs ```REX-ray``` to only do checks using the specified drivers | yes
```REXRAY_DAEMONDRIVERS``` | The daemon REST endpoint to run (defaults to ```dockervolumedriver```) | yes
```REXRAY_MINVOLSIZE``` | The minimum volume size to create | yes
```REXRAY_REMOTEMANAGEMENT``` | Set to ```true``` or ```TRUE``` to skip introspection during discovery and local instance lookups to enable remote managaement (XtremIO) | yes

## Volume Driver (Docker)
Name | Description
-----|-------------
```REXRAY_DOCKER_VOLUMETYPE``` | Specifies the type of volume, based on Storage Driver
```REXRAY_DOCKER_IOPS``` | Specifies the amount of IOPS, based on Storage Driver
```REXRAY_DOCKER_SIZE``` | Specifies the size of volumes created
```REXRAY_DOCKER_AVAILABILITYZONE``` | Specifies the availability zone, based on Storage Driver

## Storage Drivers

### Amazon Web Services (AWS)
Name | Description
-----|-------------
```AWS_ACCESS_KEY``` | |
```AWS_SECRET_KEY``` | |
```AWS_REGION``` | Override the detected region |

### Rackspace
Name | Description
-----|-------------
```OS_AUTH_URL``` | |
```OS_USERNAME``` | |
```OS_PASSWORD``` | |

### ScaleIO
Name | Description
-----|-------------
```GOSCALEIO_ENDPOINT``` | |
```GOSCALEIO_INSECURE``` | |
```GOSCALEIO_USERNAME``` | |
```GOSCALEIO_PASSWORD``` | |
```GOSCALEIO_SYSTEMID``` | |
```GOSCALEIO_SYSTEMNAME``` | |
```GOSCALEIO_PROTECTIONDOMAINID``` | |
```GOSCALEIO_PROTECTIONDOMAIN``` | |
```GOSCALEIO_STORAGEPOOLID``` | |
```GOSCALEIO_STORAGEPOOL``` | |

### XtremIO
Name | Description
-----|-------------
```GOXTREMIO_ENDPOINT``` | The API endpoint, ex. ```https://10.5.132.140/api/json```
```GOXTREMIO_USERNAME``` | The username
```GOXTREMIO_PASSWORD``` | The password
```GOXTREMIO_INSECURE``` | Set to ```true``` or ```TRUE``` to disable SSL certificate validation
```REXRAY_XTREMIO_DM``` | Set to ```true``` or ```TRUE``` to indicate that the device-mapper is installed and claiming devices
```REXRAY_XTREMIO_MULTIPATH``` | Set to ```true``` or ```TRUE``` to indicate that multipath is installed and claiming devices, overrides DM setting

# ```REX-Ray``` Library
This section outlines the primary types in the ```REX-Ray``` library as well as providing some code examples.

## Type/Method Overview
- Volume Driver (Volume Manager Interface Inward)
    - Mount
    - Unmount
    - Path
    - Create
    - Remove
- Storage Driver (External Storage Interface Outward)
    - Get Volume Mappings
    - Get Instance
    - Get Volume
    - Get Volumes Attached
    - Create Snapshot
    - Get Snapshot
    - Remove Snapshot
    - Create Volume
    - Remove Volume
    - Get Next Available Device
    - Attach Volume
    - Detach Volume
    - Copy Snapshot
- OS Driver (Local Management)
    - Get mounts
    - Mounted
    - Mount
    - Unmount
    - Format


## Storage Driver Interface
These represent the methods that should be available from storage drivers.

```go
type Driver interface {
  // GetVolumeMapping lists the block devices that are attached to the instance.
	GetVolumeMapping() ([]*BlockDevice, error)

	// GetInstance retrieves the local instance.
	GetInstance() (*Instance, error)

	// GetVolume returns all volumes for the instance based on either volumeID or volumeName
	// that are available to the instance.
	GetVolume(volumeID, volumeName string) ([]*Volume, error)

	// GetVolumeAttach returns the attachment details based on volumeID or volumeName
	// where the volume is currently attached.
	GetVolumeAttach(volumeID, instanceID string) ([]*VolumeAttachment, error)

	// CreateSnapshot is a synch/async operation that returns snapshots that have been
	// performed based on supplying a snapshotName, source volumeID, and optional description.
	CreateSnapshot(runAsync bool, snapshotName, volumeID, description string) ([]*Snapshot, error)

	// GetSnapshot returns a list of snapshots for a volume based on volumeID, snapshotID, or snapshotName.
	GetSnapshot(volumeID, snapshotID, snapshotName string) ([]*Snapshot, error)

	// RemoveSnapshot will remove a snapshot based on the snapshotID.
	RemoveSnapshot(snapshotID string) error

	// CreateVolume is sync/async and will create an return a new/existing Volume based on volumeID/snapshotID with
	// a name of volumeName and a size in GB.  Optionally based on the storage driver, a volumeType, IOPS, and availabilityZone
	// could be defined.
	CreateVolume(runAsync bool, volumeName string, volumeID string, snapshotID string, volumeType string, IOPS int64, size int64, availabilityZone string) (*Volume, error)

	// RemoveVolume will remove a volume based on volumeID.
	RemoveVolume(volumeID string) error

	// GetDeviceNextAvailable return a device path that will retrieve the next available disk device that can be used.
	GetDeviceNextAvailable() (string, error)

	// AttachVolume returns a list of VolumeAttachments is sync/async that will attach a volume to an instance based on volumeID and instanceID.
	AttachVolume(runAsync bool, volumeID, instanceID string) ([]*VolumeAttachment, error)

	// DetachVolume is sync/async that will detach the volumeID from the local instance or the instanceID.
	DetachVolume(runAsync bool, volumeID string, instanceID string) error

	// CopySnapshot is a sync/async and returns a snapshot that will copy a snapshot based on volumeID/snapshotID/snapshotName and
	// create a new snapshot of desinationSnapshotName in the destinationRegion location.
	CopySnapshot(runAsync bool, volumeID, snapshotID, snapshotName, destinationSnapshotName, destinationRegion string) (*Snapshot, error)}
```

## Volume Driver Interface
These represent the methods that should be available from volume drivers.  These drivers allow an abstracted way to
access storage drivers.

```go
type Driver interface {
	// Mount will return a mount point path when specifying either a volumeName or volumeID.  If a overwriteFs boolean
	// is specified it will overwrite the FS based on newFsType if it is detected that there is no FS present.
	Mount(volumeName, volumeID string, overwriteFs bool, newFsType string) (string, error)

	// Unmount will unmount the specified volume by volumeName or volumeID.
	Unmount(volumeName, volumeID string) error

	// Path will return the mounted path of the volumeName or volumeID.
	Path(volumeName, volumeID string) (string, error)

	// Create will create a new volume with the volumeName.
	Create(volumeName string) error

	// Remove will remove a volume of volumeName.
	Remove(volumeName string) error

	// Attach will attach a volume based on volumeName to the instance of instanceID.
	Attach(volumeName, instanceID string) (string, error)

	// Detach will detach a volume based on volumeName to the instance of instanceID.
	Detach(volumeName, instanceID string) error

	// NetworkName will return an identifier of a volume that is relevant when corelating a
	// local device to a device that is the volumeName to the local instanceID.
	NetworkName(volumeName, instanceID string) (string, error)
}
```

### Get all block devices
The following examples assumes that you have passed proper environment variables based on the guest instance.

#### Get Local Instance
```go
instance, err := driver.GetInstance()
if err != nil {
    log.Fatalf("Error: %s", err)
}
```

#### Get All Block Devices
```go
allBlockDevices, err := rexray.GetBlockDeviceMapping()
if err != nil {
    log.Fatalf("Error: %s", err)
}
```

## OS Driver Interface
These represent the methods that should be available from OS drivers.

```go
type Driver interface {
    // Shows the existing mount points
    GetMounts(string, string) ([]*mount.MountInfo, error)

    // Check whether path is mounted or not
    Mounted(string) (bool, error)

    // Unmount based on a path
    Unmount(string) error

    // Mount based on a device, target, options, label
    Mount(string, string, string, string) error

    // Format a device with a FS type
    Format(string, string, bool) error
}
```

# REX-Ray CLI
```REX-Ray``` can be used independently as a CLI tool that provides guest storage introspection and management.  The CLI should be distributed to the system that requires introspection and storage management.  It will discover proper drivers to use, and then with proper authorization, will get further details about those devices.

Once the introspection has occurred, ```REX-Ray``` can then manage manage storage using initialized drivers in a common manner between storage providers.  The providers will attach devices via any method possible to get the device attached as the next available  ```/dev/xvd_``` or one that is automatically assigned via the ```REX-Ray``` driver.

The ```REX-Ray``` CLI has a set of top-level commands that each represent logical groupings of
common categorizations. Simply execute them to find out more about them!

```bash
[0]akutz@pax:rexray$ rexray
REX-Ray:
  A guest-based storage introspection tool that enables local
  visibility and management from cloud and storage platforms.

Usage:
  rexray [flags]
  rexray [command]

Available Commands:
  volume      The volume manager
  snapshot    The snapshot manager
  device      The device manager
  driver      The driver manager
  service     The service controller
  version     Print the version
  help        Help about any command

Flags:
  -c, --config="$HOME/.rexray/config.yml": The REX-Ray configuration file
  -d, --debug=false: Enables verbose output
  -?, --help=false: Help for rexray
  -h, --host="tcp://127.0.0.1:7979": The REX-Ray service address


Use "rexray [command] --help" for more information about a command.
```

## Examples
The follow examples demonstrate how to configure storage platforms and use the ```REX-Ray``` CLI to interact with them.

### Azure
// TODO

### AWS
```bash
export REXRAY_STORAGEDRIVERS=ec2
export AWS_ACCESS_KEY=access_key AWS_SECRET_KEY="secret_key"

./rexray volume get

- providername: ec2
  instanceid: i-695bb6ab
  volumeid: vol-dedbadc3
  devicename: /dev/sda1
  region: us-west-1
  status: attached
- providername: ec2
  instanceid: i-695bb6ab
  volumeid: vol-04c4b219
  devicename: /dev/xvdb
  region: us-west-1
  status: attached
```

### Ceph
// TODO

### CloudStack
// TODO

### GCE
// TODO

### KVM
// TODO

### OpenStack
```bash
export REXRAY_STORAGEDRIVERS=openstack
export OS_AUTH_URL=https://region-a.geo-1.identity.hpcloudsvc.com:35357/v2.0/
export OS_TENANT_NAME=10677883784544-Project OS_REGION_NAME=region-a.geo-1
export OS_USERNAME=username OS_PASSWORD='password'

./rexray volume get

- name: testing4bc
  volumeid: a6aa61a1-2b2b-4d30-974b-6809fcd7cbff
  availabilityzone: az1
  status: available
  volumetype: standard
  iops: 0
  size: "1"
  networkname: ""
  attachments: []
```

### RackSpace
```bash
export REXRAY_STORAGEDRIVERS=rackspace
export OS_AUTH_URL=https://identity.api.rackspacecloud.com/v2.0 OS_USERNAME=username OS_PASSWORD='password'

./rexray volume get

- providername: RackSpace
  instanceid: 5ad7727c-aa5a-43e4-8ab7-a499295032d7
  volumeid: 738ea6b9-8c49-416c-97b7-a5264a799eb6
  devicename: /dev/xvdb
  region: DFW
  status: ""
- providername: RackSpace
  instanceid: 5ad7727c-aa5a-43e4-8ab7-a499295032d7
  volumeid: 43de157d-3dfb-441f-b832-4d2d8cf457cc
  devicename: /dev/xvdd
  region: DFW
  status: ""
```

### ScaleIO
```bash
export REXRAY_STORAGEDRIVERS=scaleio
export GOSCALEIO_ENDPOINT=https://mdm1.scaleio.local:443/api GOSCALEIO_INSECURE=true GOSCALEIO_USERNAME=admin GOSCALEIO_PASSWORD=Scaleio123 GOSCALEIO_SYSTEMID=1aa75ddc59b6a8f7 GOSCALEIO_PROTECTIONDOMAINID=ea81096700000000 GOSCALEIO_STORAGEPOOLID=1041757800000001

./rexray volume get
```

### XtremIO (iSCSI)
```bash
export GOXTREMIO_ENDPOINT="https://10.5.132.140/api/json"
export GOXTREMIO_INSECURE="true"
export GOXTREMIO_USERNAME="admin"
export GOXTREMIO_PASSWORD="Xtrem10"
export REXRAY_XTREMIO_MULTIPATH=true

./rexray volume get
```

### vSphere
// TODO

### vCloud Director
// TODO

### VIPR-C
// TODO

# REX-Ray Daemon
```REX-Ray``` can be run as a CLI for interactive usage, but it can also be executed with the ```--daemon``` flag to spawn a background process that hosts an HTTP server with a RESTful API.

## Installation
The ```REX-ray``` daemon comes with out-of-the-box support for SysV init scripts and systemd services. For example, in order to configure ```REX-Ray``` as a systemd service using the included ```rexray.service``` unit file, please follow the commands below:

```bash
# copy the service file to the systemd service unit file directory
sudo cp rexray.service /usr/lib/systemd/system/

# notify systemd about the new service & enable it to start on boot
sudo systemctl enable rexray

# start the rexray service
sudo systemctl start rexray
```

## Testing
In the case of doing local tests, since it passes HTTP via Unix socket, you can use tools like ```socat``` and others like ```curl-unix-socket``` to talk with the API.

### socat
This can be used for as a simple test of the messages that do not have bodies since it is not HTTP aware.  The following will test a basic activation message.

```bash
echo -e "GET /Plugin.Activate HTTP/1.1\r\n" | socat unix-connect:/run/docker/plugins/rexray.sock STDIO
```

### curl-unix-socket (go get github.com/Soulou/curl-unix-socket)
This utility is HTTP and Unix socket aware so can do POST messages in a HTTP friendly manner which allows us to specify a body.  

```bash
/usr/src/go/bin/curl-unix-socket -v -X POST -d '{"Name":"test22"}\r\n' unix:///run/docker/plugins/rexray.sock:/VolumeDriver.Mount

> POST /VolumeDriver.Mount HTTP/1.1
> Socket: /run/docker/plugins/rexray.sock
> Content-Length: 21
>
< HTTP/1.1 200 OK
< Content-Type: appplication/vnd.docker.plugins.v1+json
< Date: Fri, 22 May 2015 15:52:21 GMT
< Content-Length: 49
{"Mountpoint": "/var/lib/rexray/volumes/test22"}
```

# Contributions
We are actively looking for contributors to this project.  This can involve any number of area.

- Documentation
- Storage Drivers
- OS Drivers
- A future distributed model

# Licensing
---------
Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

# Support
-------
Please file bugs and issues at the Github issues page. For more general discussions you can contact the EMC Code team at <a href="https://groups.google.com/forum/#!forum/emccode-users">Google Groups</a> or tagged with **EMC** on <a href="https://stackoverflow.com">Stackoverflow.com</a>. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.
