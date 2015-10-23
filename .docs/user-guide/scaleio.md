#ScaleIO

Scale-out with simplified storage management

---

## Overview
The ScaleIO registers a storage driver named `scaleio` with the `REX-Ray`
driver manager and is used to connect and manage ScaleIO storage.

## Configuration Options
The following are the configuration options for the `scaleio` storage driver.

 EnvVar | YAML | CLI  
--------|------|------
`GOSCALEIO_ENDPOINT` | `scaleIoEndpoint` | `--scaleIoEndpoint`
`GOSCALEIO_INSECURE` | `scaleIoInsecure` | `--scaleIoInsecure`
`GOSCALEIO_USECERTS` | `scaleIoUseCerts` | `--scaleIoUseCerts`
`GOSCALEIO_USERNAME` | `scaleIoUserName` | `--scaleIoUserName`
`GOSCALEIO_PASSWORD` | `scaleIoPassword` | `--scaleIoPassword`
`GOSCALEIO_SYSTEMID` | `scaleIoSystemId` | `--scaleIoSystemId`
`GOSCALEIO_SYSTEMNAME` | `scaleIoSystemName` | `--scaleIoSystemName`
`GOSCALEIO_PROTECTIONDOMAINID` | `scaleIoProtectionDomainId` | `--scaleIoProtectionDomainId`
`GOSCALEIO_PROTECTIONDOMAIN` | `scaleIoProtectionDomainName` | `--scaleIoProtectionDomainName`
`GOSCALEIO_STORAGEPOOLID` | `scaleIoStoragePoolId` | `--scaleIoStoragePoolId`
`GOSCALEIO_STORAGEPOOL` | `scaleIoStoragePoolName` | `--scaleIoStoragePoolName`

## Activating the Driver
To activate the ScaleIO driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `scaleio` as the driver name.

## Examples
