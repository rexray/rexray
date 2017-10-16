# Microsoft

Azure storage

---

## Overview
Microsoft Azure support is included with REX-Ray as well.

<a name="azure-ud"></a>

## Azure Unmanaged Disk
The Microsoft Azure Unmanaged Disk (Azure UD) driver registers a driver
named `azureud` with the libStorage service registry and is used to connect and
mount Azure unmanaged disks from Azure page blob storage with Azure virtual
machines.

### Requirements
* An Azure account
* An Azure subscription
* An Azure storage account
* An Azure resource group
* Any virtual machine where disks are going to be attached must have the
  `lsscsi` utility installed. You can install this with `yum install lsscsi` on
  Red Hat based distributions, or with `apt-get install lsscsi` on Debian based
  distributions.

### Configuration
The following is an example with all possible fields configured. For a running
example see the [Examples](#azure-ud-examples) section.

```yaml
azureud:
  subscriptionID: abcdef01-2345-6789-abcd-ef0123456789
  resourceGroup: testgroup
  tenantID: usernamehotmail.onmicrosoft.com
  storageAccount: username
  storageAccessKey: XXXXXXXX
  clientID: 123def01-2345-6789-abcd-ef0123456789
  clientSecret: XXXXXXXX
  certPath:
  container: vhds
  useHTTPS: true
```

#### Configuration Notes
* `subscriptionID` is required, and is the UUID of your Azure subscription
* `resourceGroup` is required, and is the name of the resource group for your
  VMs and storage.
* `tenantID` is required, and is either the domain or UUID for your active
  directory account within Azure.
* `storageAccount` is required, and is the name of the storage account where
  your disks will be created.
* `storageAccessKey` is required, and is a valid access key associated with the
  `storageAccount`.
* `clientID` is required, and is the UUID of your client, which was created as
  an App Registration within your Azure active directory account. When creating
  an App Registration, this ID is shown as the Application ID.
* `clientSecret` is required if `certPath` is not provided instead. It is a
  valid access key associated with `clientID`, and is managed as part of the App
  Registration.
* `certPath` is an alternative to `clientSecret`, contains the location of a
  PKCS encoded RSA private key associated with `clientID`.
* `container` is optional, and specifies the name of an existing container
  within `storageAccount`. This container must already exist and is not created
  automatically.
* `useHTTPS` is optional, and is a boolean value on whether to use HTTPS when
  communicating with the Azure storage endpoint.

### Runtime Behavior
* The `container` config option defaults to `vhds`, and this container is
  present by default in Azure. Changing this option is only necessary if you
  want to use a different container within your storage account.
* Volume Attach/Detach operations in Azure take a long time, sometimes greater
  than 60 seconds, which is libStorage's default task timeout. When the timeout
  is hit, libStorage returns information to the caller about a queued task, and
  that the task is still running. This may cause issues for upstream callers.
  It is *highly* recommended to adjust this default timeout to 120 seconds by
  setting the `libstorage.server.tasks.exeTimeout` property. This is done in
  the `Examples` section below.

### Activating the Driver
To activate the Azure UD driver please follow the instructions for
[activating storage drivers](../servers/libstorage.md#storage-drivers), using `azureud` as
the driver name.

### Troubleshooting
* For help creating App Registrations, the steps in
  [this guide](https://www.terraform.io/docs/providers/azurerm/index.html#creating-credentials)
  cover creating a new Registration using the Azure portal and CLI.
* After creating your app registration, you must go into the
  `Required Permissions` tab and grant access to "Windows Azure Service
  Management API". Choose the delegated permission for accessing as organization
  users.
* You must also grant your app registration access to your subscription, by
  going to Subscriptions->Your `subscriptionID`->Access Control (IAM). From
  there, add your app registration as a user, which you will have to search for
  by name. Grant the role of "Owner".
* You should carefully check that your VM is compatible with the storage account
  you want to use. For example, if you need Azure Premium storage your machine
  should be of a compatible size (e.g. DS_V2, FS). For more details see the
  available VM [sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/virtual-machines-windows-sizes).  

<a name="azure-ud-examples"></a>

### Examples
Below is a full `config.yml` that works with Azure UD

```yaml
libstorage:
  # The libstorage.service property directs a libStorage client to direct its
  # requests to the given service by default. It is not used by the server.
  service: azureud
  server:
    tasks:
      exeTimeout: 120s
    services:
      azure:
        driver: azureud
        azureud:
          subscriptionID: abcdef01-2345-6789-abcd-ef0123456789
          resourceGroup: testgroup
          tenantID: usernamehotmail.onmicrosoft.com
          storageAccount: username
          storageAccessKey: XXXXXXXX
          clientID: 123def01-2345-6789-abcd-ef0123456789
          clientSecret: XXXXXXXX
```

### Caveats
* Snapshot and Copy functionality is not yet implemented
* The number of disks you can attach to a Virtual Machine depends on its type.
* Good resources for reading about disks in Azure are
  [here](https://docs.microsoft.com/en-us/azure/storage/storage-standard-storage)
  and [here](https://docs.microsoft.com/en-us/azure/storage/storage-about-disks-and-vhds-linux).
