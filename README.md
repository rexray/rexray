# libStorage
`libStorage` provides a vendor agnostic storage orchestration model,
API, and reference client and server implementations.

The design represents the knowledge we have gained from building storage orchestration tools in the past with a design goal architecturally aligned towards embedding functionality in appropriate `tools` and `storage platforms`.  **This will allow for tools to natively consume storage without extra dependencies and enable common features**.

## Summary
`libStorage` will enable common methods for consuming storage capabilities that may be present in a `local` or `external` fashion to an operating system.  These capabilities will be focused on container runtimes and storage orchestration tools including and relevant to `Docker` and `Mesos`, but additionally available abstractly for more general usage.

Additionally the common capabilities should be available across:
- Operating Systems
- Storage Platforms
- Hardware Platforms
- Virtualization Platforms

The client side implementation will be focused on `operating system` activities and include minimal dependencies to avoid unnecessarily bloating runtimes and tools.

## Storage Orchestration Tool Architecture Today
Today there are a lot of storage orchestration and abstraction tools that are present and relevant to container runtimes.  These tools tend to represent a model where the tool must be installed and running as a process locally within an operating system alongside a container runtime to function.

![Storage Orchestration Tool Architecture Today](/images/architecture-today.png "Storage Orchestration Tool Architecture Today")

*The solid green lines represent active communication paths.  The dotted black lines represent passive paths.  The orange volume represents a operating system device and volume path available to the container runtime.*

## libStorage Embedded Architecture
Embedding `libStorage` client and server components will enable `container runtimes` to communicate directly with `storage platforms`.  This represents the ideal architecture with minimal operational dependencies to support volumes for containers.

![libStorage Embedded Architecture](/images/architecture-embeddedlibstorage.png "libStorage Embedded Architecture")

## libStorage Centralized Architecture
In a centralized architecture, the `libStorage` server can be ran as a service.  In this case, the storage platform does not have the capability to communicate using the `libStorage API`, or it cannot advertise the `libStorage server`.  The `libStorage` endpoint is advertised by a tool like [REX-Ray](https://github.com/emccode/rexray), ran from anywhere, who is responsible for all control plane operations to the storage platform along with maintaing escalated credentials for these platforms.  All client based processes within the operating system are still embedded in the container runtime.

![libStorage Centralized Architecture](/images/architecture-centralized.png "libStorage Centralized Architecture")

## libStorage De-Centralized Architecture
This architecture is similar to the centralized, except the `libStorage` server is ran as a process on each operating system alongside the container runtime.

![libStorage De-Centralized Architecture](/images/architecture-decentralized.png "libStorage De-Centralized Architecture")


## API
Central to `libStorage` is the `HTTP/JSON` API.  It defines the control plane calls that occur between the `client` and `server` which can be written in any language.


## Client
The `libStorage client` initially will be written in Go and compatible with C++.  It will be focused on:
- Implementing client API of libStorage
- Operating System
  - Device
    - Discovery
    - Format
    - Mount
  - Layered Filesystems
   - Create/Remove

The design goal focuses on characteristics of:
 - Being lightweight
 - Minimal dependencies
 - Minimize obsolescense

## Server
The `libStorage server` initially will also be written in Go.  It will be focused on:
- Implementing server API of libStorage
- Returning storage platform requests to storage orchestration package


## Model
The LibStorage model defines several data structures that are easily
represented using Go structs or a portable format such as JSON.

## Documentation for LibStorage
`WORK IN PROGRESS`

[![Docs](https://readthedocs.org/projects/libstorage/badge/?version=latest)](http://libstorage.readthedocs.org)
You will find complete documentation for `LibStorage` at [libstorage.readthedocs.org](http://libstorage.readthedocs.org).


## Licensing
Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## Support
If you have questions relating to the project, please either post [Github Issues](https://github.com/emccode/libstorage/issues), join our Slack channel available by signup through [community.emc.com](https://community.emccode.com) and post questions into `#projects` or `#support`, or reach out to the maintainers directly.  The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.
