# Azure Unmanaged Disk driver Testing

## Unit/Integration Tests
The unit/integration tests must be executed on a node that is hosted within
Azure. In order to execute the tests either compile the test binary locally or
on the instance. From the root of the libStorage project execute the following:

```bash
GOOS=linux make test-azureud
```

Once the test binary is compiled, if it was built locally, copy it to the Azure
instance.

Using an SSH session to connect to the instance, please export the required
config options used by the Azure UD driver:

```bash
export AZUREUD_SUBSCRIPTIONID=<your subscription id>
export AZUREUD_RESOURCEGROUP=<your resource group name>
export AZUREUD_TENANTID=<your active directory tenant name>
export AZUREUD_STORAGEACCOUNT=<your storage account name>
export AZUREUD_STORAGEACCESSKEY=<your access key for storage account>
export AZUREUD_CLIENTID=<your client ID>
export AZUREUD_CLIENTSECRET=<your secret key for client>
```

The tests may now be executed with the following command:

```bash
./azureud.test
```

An exit code of `0` means the tests completed successfully. If there are errors
then it may be useful to run the tests once more with increased logging:

```bash
LIBSTORAGE_LOGGING_LEVEL=debug ./azureud.test -test.v
