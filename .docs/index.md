# libStorage

Opening up storage for all

---

## Overview
`libStorage` is an open source, platform agnostic, storage provisioning and
orchestration framework, model, and API.

## Features
The following features unique to this project make it a perfect choice for
adding value to upstream applications by centralizing storage management:

- A standardized storage orchestration [model and API](http://docs.libstorage.apiary.io)
- A lightweight, reference client implementation with a minimal dependency
  footprint
- The ability to embed both the libStorage client and server, creating native
  application integration opportunities

## Operations
`libStorage` supports the following operations:

Resource Type | Operation | Description
--------------|-----------|------------
Volume | List / Inspect | Get detailed information about one to many volumes
       | Create / Remote | Manage the volume lifecycle
       | Attach / Detach | Provision volumes to a client
       | Mount / Unmount | Make attached volumes ready-to-use, local file systems
Snapshot | | Coming soon
Storage Pool | | Coming soon

## Getting Started
Using libStorage can be broken down into several, distinct steps:

1. Configuring [libStorage](./user-guide/config.md)
2. Understanding the [API](http://docs.libstorage.apiary.io)
3. Identifying a production server and client implementation, such as
   [REX-Ray](https://rexray.rtfd.org)

## Getting Help
To get help with libStorage, please use the
[discussion group](https://groups.google.com/forum/#!forum/emccode-users),
[GitHub issues](https://github.com/emccode/libstorage/issues), or tagging
questions with **EMC** at [StackOverflow](https://stackoverflow.com).

The code and documentation are released with no warranties or SLAs and are
intended to be supported through a community driven process.
