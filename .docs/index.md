# libStorage

Opening up storage for all

---

## Overview
`libStorage` provides a vendor agnostic storage orchestration model,
API, and reference client and server implementations. It focuses on being a
portable storage driver framework that brings external storage functionality to
any platform or application.

## Features
The project has some very unique qualities that make it perfect for embedding
in upstream projects to centralize external storage functionality.

- Lightweight client package enable minimal dependencies to provide full
featured storage functionality to platforms
- Embedded and remotable modes for providing choice of centralized control of
storage operations
- Optionally enables storage platforms to serve as libStorage servers making
integration of application platforms native
- Dynamically downloaded executors run specific storage tasks without critical
long running plugins per host
- Includes Go client/server packages for simple integration to other platforms
and applications
- Flexible HTTP/JSON API for other deployment opportunities

## Operations
Today `libStorage` supports the following volume management features.

- List/Inspect for retrieving volumes and detailed information
- Create/Remove for managing volume lifecycle
- Attach/Detach for getting volumes to instaces to be used
- Mount/Unmount to comprehensively get volumes to instances, discover,
optionally format, and mount
- Path to review the existing mounted path of a volume
- Map to list the current attached volumes to an instance

The operations for `Snapshots` and `Storage Pools` is planned for future
releases.

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
