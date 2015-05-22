# REX-Ray


## Overview
```REX-Ray``` is a Go package for guest storage introspection that is meant to provide visibility and management of external/underlying storage that is attached via methods specified in drivers.  This storage can be from a specific storage platform in addition to being provided by virtual infrastructure.

This can either be integrated at a package level to other Go based projects, or it can used in as a daemon.  Currently, when spinning up the daemon it would result in a ```Unix socket``` or a ```HTTP endpoint```.  The [REX-RayCLI](https://github.com/emccode/rexraycli) repo has a compiled CLI that can be found under the releases which can start the daemon easily.  The CLI should provide the majority of functionality for ```REX-Ray```.

There are three types of drivers.  The ```Volume Driver``` represents ```Volume Manager``` abstractions that should satisfy requirements from things that wish to manage storage.  For example, ```Docker``` would leverage this interface which matches the Docker storage API.  The ```Storage Driver``` is an abstraction for ```External Storage``` which can be virtual or from a storage platform.  Lastly, the ```OS Driver``` is provides an abstraction for the differences relating to mounting across operating systems.

The driver to be used is automatically detected or hints can be provided.  Drivers are then intialiazed as adapters which allow the retrieval of guest identifiers and further information from other platforms that are relevant to storage management.

## State
We have a first release available that support all of the following capabilities!  

## Examples
One of the best examples of this in action would be to review the [Dogged](https://github.com/emccode/dogged) repo.  This repo maintains efforts for the EMC {code} team relating to embedding REX-Ray inside of Container Engines such as Docker.  Here you will REX-Ray enabling Docker to manage it's own storage via Container Data Volumes.

## Features
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


## CLI
REX-Ray can be used independently as a CLI tool instead of embedding via Go packages.  See the [REX-RayCLI](https://github.com/emccode/rexraycli) repo.

## Environment Variables - General

    REXRAY_DEBUG (optional) - show debug messages
    REXRAY_STORAGEDRIVERS (optional) - only do checks using these drivers
    REXRAY_OSDRIVERS (optional) - only do checks using these drivers
    REXRAY_VOLUMEDRIVERS (optional) - only do checks using these drivers
    REXRAY_MINVOLSIZE - minimum volume size to create

## Environment Variables - Storage Drivers

    AWS_ACCESS_KEY - (AWS)
    AWS_SECRET_KEY - (AWS)
    AWS_REGION (AWS) - Override the detected region
    OS_AUTH_URL - (RACKSPACE)
    OS_USERNAME - (RACKSPACE)
    OS_PASSWORD - (RACKSPACE)
    GOSCALEIO_ENDPOINT - (SCALEIO)
    GOSCALEIO_INSECURE - (SCALEIO)
    GOSCALEIO_USERNAME - (SCALEIO)
    GOSCALEIO_PASSWORD - (SCALEIO)
    GOSCALEIO_SYSTEMID - (SCALEIO)
    GOSCALEIO_PROTECTIONDOMAINID - (SCALEIO)
    GOSCALEIO_STORAGEPOOLID - (SCALEIO)

## Environment Variables - Volume Drivers - Docker

    REXRAY_DOCKER_VOLUMETYPE - Specifies the type of volume, based on Storage Driver
    REXRAY_DOCKER_IOPS - Specifies the amount of IOPS, based on Storage Driver
    REXRAY_DOCKER_SIZE - Specifies the size of volumes created
    REXRAY_DOCKER_AVAILABILITYZONE - Specifies the availability zone, based on Storage Driver



## Storage Drivers - Examples
- AWS
- RackSpace
- ScaleIO
- ..more to come

## Storage Driver - Interface
These represent the methods that should be available from storage drivers.

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

### Get all block devices
The following examples assumes that you have passed proper environment variables based on the guest instance.

#### Get Local Instance
    instance, err := driver.GetInstance()
    if err != nil {
      log.Fatalf("Error: %s", err)
    }

#### Get All Block Devices
    allBlockDevices, err := rexray.GetBlockDeviceMapping()
    if err != nil {
      log.Fatalf("Error: %s", err)
    }

## OS Driver - Interface
These represent the methods that should be available from OS drivers.

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

## Volume Driver - Interface
These represent the methods that should be available from Volume drivers.

    type Driver interface {
    	//MountVolume will attach a Volume, prepare for mounting, and mount
    	Mount(string, string, bool, string) (string, error)

    	//UnmountVolume will unmount and detach a Volume
    	Unmount(string, string) error

    	//Path will return the mountpoint of a volume
    	Path(string, string) (string, error)

    	//Create will create a remote volume
    	Create(string) error

    	//Remove will remove a remote volume
    	Remove(string) error
    }

## Contributions
We are actively looking for contributors to this project.  This can involve any number of area.

  - Documentation
  - Storage Drivers
  - OS Drivers
  - A future distributed model


Licensing
---------
Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

Support
-------
Please file bugs and issues at the Github issues page. For more general discussions you can contact the EMC Code team at <a href="https://groups.google.com/forum/#!forum/emccode-users">Google Groups</a> or tagged with **EMC** on <a href="https://stackoverflow.com">Stackoverflow.com</a>. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.
