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
Dell EMC | [ScaleIO](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#dell-emc-scaleio), [Isilon](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#dell-emc-isilon)
[Oracle VirtualBox](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#virtualbox) | Virtual Media
Amazon EC2 | [EBS](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#aws-ebs), [EFS](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#aws-efs), [S3FS](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#aws-s3fs)
Ceph | [RBD](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#ceph-rbd)
GCE | [Persistent Disk](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#gce-persistent-disk)
Microsoft Azure | [Unmanaged Disk](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers#azure-ud)
