# RexRay


## Overview
```RexRay``` is a cross-platform storage introspection application that is meant to provide visibility and management of external/underlying storage that is attached via methods specified in drivers.  This storage can be from a specific storage platform in addition to being provided by virtual infrastructure.

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

## Storage Drivers - Examples

### Azure

### Ceph

### AWS
    AWS_ACCESS_KEY=access_key AWS_SECRET_KEY="secret_key" go run /usr/src/go/src/github.com/emccode/rexray/rexray.go

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

### GCE

### RackSpace
    OS_AUTH_URL=https://identity.api.rackspacecloud.com/v2.0 OS_USERNAME=username OS_PASSWORD='password' go run /usr/src/go/src/github.com/emccode/rexray/rexray.go

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

### ScaleIO




Licensing
---------
Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

Support
-------
Please file bugs and issues at the Github issues page. For more general discussions you can contact the EMC Code team at <a href="https://groups.google.com/forum/#!forum/emccode-users">Google Groups</a> or tagged with **EMC** on <a href="https://stackoverflow.com">Stackoverflow.com</a>. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.
