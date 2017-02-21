Functional tests for Azure Unmanaged Disk driver.

It requires to be run inside of Azure instance.

In order to run test the following environment variables should be defined
(they should be filled with your data):
        AZUREUD_SUBSCRIPTION_ID=<subscrption_id>                  # your subscription ID
        AZUREUD_RESOURCE_GROUP=<resource_group_name>              # your resource group name
        AZUREUD_TENANT_ID=<tenant_id>                             # your tenant ID
        AZUREUD_CLIENT_ID=<client_id>                             # id of your client (application)
        AZUREUD_CLIENT_SECRET=<put yout secret>                   # your client(application) secret key
        AZUREUD_STORAGE_ACCOUNT=<storage_account_name>            # your storage account name
        AZUREUD_STORAGE_ACCESS_KEY=<storage_account_access_key>   # your storage account access key

The driver and tests do not create container, instance, etc, all entities should
be created before to run tests / use of libstorage.
