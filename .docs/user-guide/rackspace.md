#Rackspace

They manage your services, we manage your storage

---

## Overview
The Rackspace driver registers a storage driver named `rackspace` with the
`REX-Ray` driver manager and is used to connect and manage storage on Rackspace
instances.

## Configuration Options
The following are the configuration options for the `rackspace` storage driver.

 EnvVar | YAML | CLI  
--------|------|------
`OS_AUTH_URL` | `rackspaceAuthUrl` | `--rackspaceAuthUrl`
`OS_USERID` | `rackspaceUserId` | `--rackspaceUserId`
`OS_USERNAME` | `rackspaceUserName` | `--rackspaceUserName`
`OS_PASSWORD` | `rackspacePassword` | `--rackspacePassword`
`OS_TENANT_ID` | `rackspaceTenantId` | `--rackspaceTenantId`
`OS_TENANT_NAME` | `rackspaceTenantName` | `--rackspaceTenantName`
`OS_DOMAIN_ID` | `rackspaceDomainId` | `--rackspaceDomainId`
`OS_DOMAIN_NAME` | `rackspaceDomainName` | `--rackspaceDomainName` 

## Activating the Driver
To activate the Rackspace driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `rackspace` as the driver name.

## Examples
