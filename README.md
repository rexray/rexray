# REX-Ray


## Overview
```REX-Ray``` is a Go package for guest storage introspection that is meant to provide visibility and management of external/underlying storage that is attached via methods specified in drivers.  This storage can be from a specific storage platform in addition to being provided by virtual infrastructure.

The driver to be used is automatically detected or hints can be provided.  Drivers are then intialized to retrieve guest identifiers and further information from other platforms that are relevant to storage management.

## State
We have a first release available that support all of the following capabilities!

## Examples
One of the best examples of this in action would be to review the [REX-Ray CLI](https://github.com/emccode/rexraycli) tool.

## Features
- Visibility
 - Local Instance
 - Volumes
- Management
 - Volume Create/Remove
 - Volume Snapshot/Unsnapshot
 - Volume Attach/Detach
 - Replicate Snapshot

## CLI
See the [REX-RayCLI](https://github.com/emccode/rexraycli) repo.

## Environment Variables

    REXRAY_DEBUG - show debug messages
    REXRAY_STORAGEDRIVERS - only do checks using these drivers
    REXRAY_MINVOLSIZE - minimum volume size to create
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


## Storage Drivers - Examples
- AWS
- RackSpace
- ScaleIO
- ..more to come

## Storage Driver - Interface
These represent the methods that should be available from storage drivers.

      type Driver interface {
      	// Lists the block devices that are attached to the instance
      	GetBlockDeviceMapping() (interface{}, error)
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
