# libStorage [![GoDoc](https://godoc.org/github.com/emccode/libstorage?status.svg)](http://godoc.org/github.com/emccode/libstorage) [![Build Status](http://travis-ci.org/emccode/libstorage.svg?branch=master)](https://travis-ci.org/emccode/libstorage) [![Go Report Card](https://goreportcard.com/badge/github.com/emccode/libstorage)](https://goreportcard.com/report/github.com/emccode/libstorage) [![codecov](https://codecov.io/gh/emccode/libstorage/branch/master/graph/badge.svg)](https://codecov.io/gh/emccode/libstorage)
`libStorage` provides a vendor agnostic storage orchestration model, API, and
reference client and server implementations.

## Overview
`libStorage` enables storage consumption by leveraging methods commonly
available, locally and/or externally, to an operating system (OS).

### The Past
The `libStorage` project and its architecture represents a culmination of
experience gained from the project authors' building of
[several](https://www.emc.com/cloud-virtualization/virtual-storage-integrator.htm)
different
[storage](https://www.emc.com/storage/storage-analytics.htm)
orchestration [tools](https://github.com/emccode/rexray). While created using
different languages and targeting disparate storage platforms, all the tools
were architecturally aligned and embedded functionality directly inside the
tools and affected storage platforms.

This shared design goal enabled tools that natively consumed storage, sans
external dependencies.

### The Present
Today `libStorage` focuses on adding value to container runtimes and storage
orchestration tools such as `Docker` and `Mesos`, however the `libStorage`
framework is available abstractly for more general usage across:

* Operating systems
* Storage platforms
* Hardware platforms
* Virtualization platforms

The client side implementation, focused on operating system activities,
has a minimal set of dependencies in order to avoid a large, runtime footprint.

## Storage Orchestration Tools Today
Today there are many storage orchestration and abstraction tools relevant to
to container runtimes. These tools often must be installed locally and run
alongside the container runtime.

![Storage Orchestration Tool Architecture Today](/.docs/.themes/yeti/img/architecture-today.png "Storage Orchestration Tool Architecture Today")

*The solid green lines represent active communication paths. The dotted black
lines represent passive paths. The orange volume represents a operating system
device and volume path available to the container runtime.*

## libStorage Embedded Architecture
Embedding `libStorage` client and server components enable container
runtimes to communicate directly with storage platforms, the ideal
architecture. This design requires minimal operational dependencies and is
still able to provide volume management for container runtimes.

![libStorage Embedded Architecture](/.docs/.themes/yeti/img/architecture-embeddedlibstorage.png "libStorage Embedded Architecture")

## libStorage Centralized Architecture
In a centralized architecture, `libStorage` is hosted as a service, acting as a
go-between for container runtimes and backend storage platforms.

The `libStorage` endpoint is advertised by a tool like [REX-Ray](https://github.com/emccode/rexray), run from anywhere, and is
responsible for all control plane operations to the storage platform along with
maintaining escalated credentials for these platforms. All client based
processes within the operating system are still embedded in the container
runtime.

![libStorage Centralized Architecture](/.docs/.themes/yeti/img/architecture-centralized.png "libStorage Centralized Architecture")

## libStorage Decentralized Architecture
Similar to the centralized architecture, this implementation design involves
running a separate `libStorage` process alongside each container runtime, across
one or several hosts.

![libStorage De-Centralized Architecture](/.docs/.themes/yeti/img/architecture-decentralized.png "libStorage De-Centralized Architecture")

## API
Central to `libStorage` is the `HTTP`/`JSON` API. It defines the control plane
calls that occur between the client and server. While the `libStorage` package
includes reference implementations of the client and server written using Go,
both the client and server could be written using any language as long as both
adhere to the published `libStorage` API.

## Client
The `libStorage` client is responsible for discovering a host's instance ID
and the next, available device name. The client's reference implementation is
written using Go and is compatible with C++.

The design goal of the client is to be lightweight, portable, and avoid
obsolescence by minimizing dependencies and focusing on deferring as much of
the logic as possible to the server.

## Server
The `libStorage` server implements the `libStorage` API and is responsible for
coordinating requests between clients and backend orchestration packages. The
server's reference implementation is also written using Go.

## Model
The `libStorage` [model](http://libstorage.rtfd.org/en/latest/user-guide/model/)
defines several data structures that are easily represented using Go structs or
a portable format such as JSON.

## Documentation [![Docs](https://readthedocs.org/projects/libstorage/badge/?version=latest)](http://libstorage.readthedocs.org)
The `libStorage` documentation is available at
[libstorage.rtfd.org](http://libstorage.rtfd.org).
