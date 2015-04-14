# RexRay


## Overview
```RexRay``` is a Go package for guest storage introspection that is meant to provide visibility and management of external/underlying storage that is attached via methods specified in drivers.  This storage can be from a specific storage platform in addition to being provided by virtual infrastructure.

The driver to be used is automatically detected or hints can be provided.  Drivers are then intialized to retrieve guest identifiers and further information from other platforms that are relevant to storage management.

## State
Currently it has view only capabilities.  Working on more drivers, and actual management capabilities.

## Features
- Visibility
- Management
 - Disk Provision
 - Disk Snapshot/Unsnapshot
 - Disk Attach/Detach

## Environment Variables

    REXRAY_DEBUG - show debug messages
    REXRAY_STORAGEDRIVERS - only do checks using these drivers
    AWS_ACCESS_KEY - (AWS)
    AWS_SECRET_KEY - (AWS)
    OS_AUTH_URL - (RACKSPACE)
    OS_USERNAME - (RACKSPACE)
    OS_PASSWORD - (RACKSPACE)

## Storage Drivers - Examples
- AWS
- RackSpace
- ..more to come

### Get all block devices
The following example assumes that you have passed proper environment variables based on the guest instance.

    allBlockDevices, err := rexray.GetBlockDeviceMapping()
    if err != nil {
      log.Fatalf("Error: %s", err)
    }



Licensing
---------
Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

Support
-------
Please file bugs and issues at the Github issues page. For more general discussions you can contact the EMC Code team at <a href="https://groups.google.com/forum/#!forum/emccode-users">Google Groups</a> or tagged with **EMC** on <a href="https://stackoverflow.com">Stackoverflow.com</a>. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.
