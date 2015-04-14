# rexraycli
A CLI implementation of RexRay providing guest storage introspection and management.  The CLI should be distributed to the system that requires introspection and storage management.  It will discover proper drivers to use, and then with proper authorization, will get further details about those devices.  In the long term, managenement capabilities for volumes, and snapshots should be present.

```RexrayCLI``` is an implementation of [Rexray](https://github.com/emccode/rexray).  It provides a working application, but as well a working example of using the ```Rexray``` Go package.


## Manually Building

    docker run --rm -it -v $GOPATH:/go -w /go/src/github.com/emccode/rexraycli golang:1.4.2-cross make release


## Environment Variables

    REXRAY_DEBUG - show debug messages
    REXRAY_STORAGEDRIVERS - only do checks using these drivers
    AWS_ACCESS_KEY - (AWS)
    AWS_SECRET_KEY - (AWS)
    OS_AUTH_URL - (RACKSPACE)
    OS_USERNAME - (RACKSPACE)
    OS_PASSWORD - (RACKSPACE)

## CLI Usage Examples
The CLI can be built, or you can retrieve pre-compiled executables from the Github releases.

    get-storage - retrieve storage details for guest instance
    version - show Rexray version

### Azure

### AWS
    AWS_ACCESS_KEY=access_key AWS_SECRET_KEY="secret_key" ./rexray get-storage

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


### Ceph

### CloudStack

### GCE

### KVM

### OpenStack

### RackSpace
    OS_AUTH_URL=https://identity.api.rackspacecloud.com/v2.0 OS_USERNAME=username OS_PASSWORD='password' ./rexray

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

### vSphere

### vCloud Director

### VIPR-C


Licensing
---------
Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

Support
-------
Please file bugs and issues at the Github issues page. For more general discussions you can contact the EMC Code team at <a href="https://groups.google.com/forum/#!forum/emccode-users">Google Groups</a> or tagged with **EMC** on <a href="https://stackoverflow.com">Stackoverflow.com</a>. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.
