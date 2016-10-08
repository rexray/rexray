# Storage Providers

Connecting storage and platforms...

---

## Overview
The list of storage providers supported by REX-Ray now mirrors the validated
storage platform table from the libStorage project.

!!! note "note"

    The initial REX-Ray 0.4.x release omits support for several,
    previously verified storage platforms. These providers will be
    reintroduced incrementally, beginning with 0.4.1. If an absent driver
    prevents the use of REX-Ray, please continue to use 0.3.3 until such time
    the storage platform is introduced in REX-Ray 0.4.x. Instructions on how
    to install REX-Ray 0.3.3 may be found
    [here](./installation.md#rex-ray-033).

## Supported Providers
The following storage providers and platforms are supported by REX-Ray.

Provider              | Storage Platform(s)
----------------------|--------------------
EMC | [ScaleIO](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/#scaleio), [Isilon](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/#isilon)
[Oracle VirtualBox](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/#virtualbox) | Virtual Media
Amazon EC2 | [EBS](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#aws-ebs), [EFS](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#aws-efs)

## Coming Soon
Support for the following storage providers will be reintroduced in upcoming
releases:

Provider              | Storage Platform(s)
----------------------|--------------------
Google Compute Engine (GCE) | Disk
Open Stack | Cinder
Rackspace | Cinder
EMC | XtremIO, VMAX
