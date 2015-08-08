# Overview
```REX-Ray``` is a Go package for guest storage introspection that is meant to provide visibility and management of external/underlying storage that is attached via methods specified in drivers.  This storage can be from a specific storage platform in addition to being provided by virtual infrastructure.

This can either be integrated at a package level to other Go based projects, or it can used in as a daemon.  Currently, when spinning up the daemon it would result in a ```Unix socket``` or a ```HTTP endpoint```.  The ```REX-Ray``` CLI can be used to start the daemon. In fact, the CLI should provide the majority of functionality for ```REX-Ray```.

There are three types of drivers.  The ```Volume Driver``` represents ```Volume Manager``` abstractions that should satisfy requirements from things that wish to manage storage.  For example, ```Docker``` would leverage this interface which matches the Docker storage API.  The ```Storage Driver``` is an abstraction for ```External Storage``` which can be virtual or from a storage platform.  Lastly, the ```OS Driver``` is provides an abstraction for the differences relating to mounting across operating systems.

The driver to be used is automatically detected or hints can be provided.  Drivers are then initialized as adapters which allow the retrieval of guest identifiers and further information from other platforms that are relevant to storage management.

The following example shows how easy it is to use REX-ray to get a list of volumes from a storage platform such as Amazon Web Services (AWS):

```bash
[0]akutz@pax:~$ export REXRAY_STORAGEDRIVERS=ec2
[0]akutz@pax:~$ export AWS_ACCESS_KEY=access_key
[0]akutz@pax:~$ export AWS_SECRET_KEY=secret_key
[0]akutz@pax:~$ rexray get-volume

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

See below for more examples of using the ```REX-ray``` CLI, features like Docker integration, and more.

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
# start the docker daemon and active the rex-ray driver
sudo rexray --daemon --host=unix:///run/docker/plugins/rexray.sock

# create a new container with a volume that leverages the rex-ray driver
docker run --volume-driver=rexray -v volumename:/pathtomount container
```

Additionally, the [Dogged](https://github.com/emccode/dogged) repo maintains efforts for the EMC {code} team relating to embedding ```REX-Ray``` inside of Container Engines such as Docker. Here you will find ```REX-Ray``` enabling Docker to manage its own storage via Container Data Volumes.

# Downloading
See the releases area for downloadable binaries.

# Building
// TODO

This might currently require upstream additions for the Goamz package to github.com/clintonskitson/goamz at the snapcopy branch.

```bash
docker run --rm -it -v $GOPATH:/go -w /go/src/github.com/emccode/rexray/rexray golang:1.4.2-cross make release
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

## Volume Driver Interface
These represent the methods that should be available from Volume drivers.

```go
type Driver interface {
    // MountVolume will attach a Volume, prepare for mounting, and mount
    Mount(string, string, bool, string) (string, error)

    // UnmountVolume will unmount and detach a Volume
    Unmount(string, string) error

    // Path will return the mountpoint of a volume
    Path(string, string) (string, error)

    // Create will create a remote volume
    Create(string) error

    // Remove will remove a remote volume
    Remove(string) error
}
```

## Storage Driver Interface
These represent the methods that should be available from storage drivers.

```go
type Driver interface {
    // Lists the block devices that are attached to the instance
    GetVolumeMapping() (interface{}, error)

    // Get the local instance
    GetInstance() (interface{}, error)

    // Get all Volumes available from infrastructure and storage platform
    GetVolume(string, string) (interface{}, error)

    // Get the currently attached Volumes
    GetVolumeAttach(string, string) (interface{}, error)

    // Create a snpashot of a Volume
    CreateSnapshot(bool, string, string, string) (interface{}, error)

    // Get all Snapshots or specific Snapshots
    GetSnapshot(string, string, string) (interface{}, error)

    // Remove Snapshot
    RemoveSnapshot(string) error

    // Create a Volume from scratch, from a Snaphot, or from another Volume
    CreateVolume(bool, string, string, string, string, int64, int64, string) (interface{}, error)

    // Remove Volume
    RemoveVolume(string) error

    // Get the next available Linux device for attaching external storage
    GetDeviceNextAvailable() (string, error)

    // Attach a Volume to an Instance
    AttachVolume(bool, string, string) (interface{}, error)

    // Detach a Volume from an Instance
    DetachVolume(bool, string, string) error

    // Copy a Snapshot to another region
    CopySnapshot(bool, string, string, string, string, string) (interface{}, error)
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

## Commands
Name | Description
-----|------------
```attach-volume``` | Attach a remote volume to this instance
```copy-snapshot``` | Copy a snapshot to another snapshot
```detach-volume``` | Detach a remote volume from this instance
```format-device``` | Format an attached device
```get-instance``` | Get the local storage instance information
```get-mount``` | Get the local mounts
```get-snapshot``` | Get remote volume snapshots
```get-volume``` | Get remote volumes
```get-volumemap``` | Get volume mapping
```get-volumepath``` | Get local mount path of a remote volume
```mount-device``` | Mount a local device to a mount path
```mount-volume``` | Mount a remote volume to a mount path
```new-snapshot``` | Create a new snapshot
```new-volume``` | Create a new volume
```remove-snapshot``` | Remote a snapshot
```remove-volume``` | Remove a remote volume
```unmount-device``` | Unmount a local device
```unmount-volume``` | Unmount a remote volume from this instance
```version``` | Print the ```REX-ray``` CLI version

## Examples
The follow examples demonstrate how to configure storage platforms and use the ```REX-ray``` CLI to interact with them.

### Azure
// TODO

### AWS
```bash
export REXRAY_STORAGEDRIVERS=ec2
export AWS_ACCESS_KEY=access_key AWS_SECRET_KEY="secret_key"

./rexray get-volume

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
// TODO

### RackSpace
```bash
export REXRAY_STORAGEDRIVERS=rackspace
export OS_AUTH_URL=https://identity.api.rackspacecloud.com/v2.0 OS_USERNAME=username OS_PASSWORD='password'

./rexray get-volume

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

./rexray get-volume
```

### XtremIO (iSCSI)
```bash
export GOXTREMIO_ENDPOINT="https://10.5.132.140/api/json"
export GOXTREMIO_INSECURE="true"
export GOXTREMIO_USERNAME="admin"
export GOXTREMIO_PASSWORD="Xtrem10"
export REXRAY_XTREMIO_MULTIPATH=true

./rexray get-volume
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
The ```REX-ray``` daemon comes with out-of-the-box support for SysV init scripts and systemd services. For example, in order to configure ```REX-ray``` as a systemd service using the included ```rexray.service``` unit file, please follow the commands below:

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
{"Mountpoint": "/var/lib/docker/volumes/test22"}
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
