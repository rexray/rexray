# Microsoft

Azure storage

---

## Overview
Microsoft Azure support is included with REX-Ray as well.

<a name="azure-ud"></a>

## Azure Unmanaged Disk
The Azure Unmanaged Disk (Azure UD) plug-in can be installed with the following
command:

```bash
$ docker plugin install rexray/azureud \
  AZUREUD_CLIENTID=123def01-2345-6789-abcd-ef0123456789 \
  AZUREUD_CLIENTSECRET=XXXXXXXX \
  AZUREUD_RESOURCEGROUP=testgroup \
  AZUREUD_STORAGEACCESSKEY=XXXXXXXX \
  AZUREUD_STORAGEACCOUNT=username \
  AZUREUD_SUBSCRIPTIONID=abcdef01-2345-6789-abcd-ef0123456789 \
  AZUREUD_TENANTID=usernamehotmail.onmicrosoft.com
```

##### Requirements

See [Azure UD](../../../storage-providers/microsoft.md#azure-ud) driver documentation for detailed
requirements

##### Privileges
The Azure UD plug-in requires the following privileges:

Type | Value
-----|------
network | `host`
mount | `/dev`
allow-all-devices | `true`
capabilities | `CAP_SYS_ADMIN`

##### Configuration
The following environment variables can be used to configure the Azure UD
plug-in:

Environment Variable | Description | Default | Required
---------------------|-------------|---------|---------
`AZUREUD_CLIENTID` | UUID of your client, which was created as an App Registration within your Azure active directory account |  | ✓
`AZUREUD_CLIENTSECRET` | Valid access key associated with `AZUREUD_CLIENTID` | | ✓
`AZUREUD_CONTAINER` | The name of an existing container within `storageAccount`. This container must already exist and is not created automatically. | `vhds` |
`AZUREUD_RESOURCEGROUP` | Name of the resource group for your VMs and storage | | ✓
`AZUREUD_STORAGEACCESSKEY` | Valid access key associated with `AZUREUD_STORAGEACCOUNT` | | ✓
`AZUREUD_STORAGEACCOUNT` | Name of the storage account where your disks will be created | | ✓
`AZUREUD_SUBSCRIPTIONID` | UUID of your Azure subscription | | ✓
`AZUREUD_TENANTID` | Domain or UUID for your active directory account within Azure | | ✓
`AZUREUD_USEHTTPS` | Boolean value on whether to use HTTPS when communicating with the Azure storage endpoin | `true` |
`HTTP_PROXY` | Address of HTTP proxy server to gain access to API endpoint | |
