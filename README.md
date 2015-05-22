# REX-RayCLI
A CLI implementation of RexRay providing guest storage introspection and management.  The CLI should be distributed to the system that requires introspection and storage management.  It will discover proper drivers to use, and then with proper authorization, will get further details about those devices.

Once the introspection has occured, ```Rexray``` can then manage manage storage using initialized drivers in a common manner between storage providers.  The providers will attach devices via any method possible to get the device attached as the next available  ```/dev/xvd_``` or one that is automatically assigned via the REX-Ray driver.

```REX-RayCLI``` is an implementation of [REX-Ray](https://github.com/emccode/rexray).  It provides a working application, but as well a working example of using the ```Rexray``` Go package.

## Environment Variables

    REXRAY_DEBUG - show debug messages
    REXRAY_STORAGEDRIVERS - only do checks using these drivers
    AWS_ACCESS_KEY - (AWS)
    AWS_SECRET_KEY - (AWS)
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

## CLI Usage Examples
The CLI can be built, or you can retrieve pre-compiled executables from the Github releases.

    attach-volume - attach a remote volume to this instance
    copy-snapshot - copy a snapshot to another snapshot
    detach-volume - detach a remote volume from this instance
    format-device - format an attached device
    get-instance - get the local storage instance information
    get-mount - get the local mounts
    get-snapshot - get remote volume snapshots
    get-volume - get remote volumes
    get-volumemap - get volume mapping
    get-volumepath - get local mount path of a remote volume
    mount-device - mount a local device to a mount path
    mount-volume - mount a remote volume to a mount path
    new-snapshot - create a new snapshot
    new-volume - create a new volume
    remove-snapshot - remote a snapshot
    remove-volume - remove a remote volume
    unmount-device - unmount a local device
    unmount-volume - unmount a remote volume from this instance
    version


### Azure

### AWS
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


### Ceph

### CloudStack

### GCE

### KVM

### OpenStack

### RackSpace
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

### ScaleIO
    export REXRAY_STORAGEDRIVERS=scaleio
    export GOSCALEIO_ENDPOINT=https://mdm1.scaleio.local:443/api GOSCALEIO_INSECURE=true GOSCALEIO_USERNAME=admin GOSCALEIO_PASSWORD=Scaleio123 GOSCALEIO_SYSTEMID=1aa75ddc59b6a8f7 GOSCALEIO_PROTECTIONDOMAINID=ea81096700000000 GOSCALEIO_STORAGEPOOLID=1041757800000001
    ./rexray get-volume


### vSphere

### vCloud Director

### VIPR-C

## Downloading
See the releases area for downloadable binaries.

## Manually Building
This might currently require upstream additions for the Goamz package to github.com/clintonskitson/goamz at the snapcopy branch.

    docker run --rm -it -v $GOPATH:/go -w /go/src/github.com/emccode/rexraycli golang:1.4.2-cross make release


## Contributing
We are always looking for contributors!

  - Documentation
  - Upstream [REX-Ray](https://github.com/emccode/rexray) storage and infrastructure drivers
  - Additional enhancements to CLI commands

Licensing
---------
Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

Support
-------
Please file bugs and issues at the Github issues page. For more general discussions you can contact the EMC Code team at <a href="https://groups.google.com/forum/#!forum/emccode-users">Google Groups</a> or tagged with **EMC** on <a href="https://stackoverflow.com">Stackoverflow.com</a>. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.
