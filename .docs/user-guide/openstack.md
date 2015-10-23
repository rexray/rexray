#OpenStack

Making storage management as transparent as the stack

---

## Overview
The OpenStack driver registers a storage driver named `openstack` with the
`REX-Ray` driver manager and is used to connect and manage storage on OpenStack
instances.

## Configuration Options
The following are the configuration options for the `openstack` storage driver.

 EnvVar | YAML | CLI
--------|------|------
`OS_AUTH_URL` | `openstackAuthUrl` | `--openstackAuthUrl`
`OS_USERID` | `openstackUserId` | `--openstackUserId`
`OS_USERNAME` | `openstackUserName` | `--openstackUserName`
`OS_PASSWORD` | `openstackPassword` | `--openstackPassword`
`OS_TENANT_ID` | `openstackTenantId` | `--openstackTenantId`
`OS_TENANT_NAME` | `openstackTenantName` | `--openstackTenantName`
`OS_DOMAIN_ID` | `openstackDomainId` | `--openstackDomainId`
`OS_DOMAIN_NAME` | `openstackDomainName` | `--openstackDomainName`
`OS_REGION_NAME` | `openstackRegionName` | `--openstackRegionName`
`OS_AVAILABILITY_ZONE_NAME` | `openstackAvailabilityZoneName` | `--openstackAvailabilityZoneName`

## Activating the Driver
To activate the OpenStack driver please follow the instructions for
[activating storage drivers](/user-guide/config#activating-storage-drivers),
using `openstack` as the driver name.

## Examples
