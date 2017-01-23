Functional tests for Azure driver.

It requires to be run either inside of Azure instance or instance ID should be defined via environment variable 'AZURE_INSTANCE_ID' that points to existing running Azure instance.

In order to run test the following environment variables should be defined (they should be filled with your data):
        AZURE_INSTANCE_ID=<instance_name>                       # it is required if run test outside Azure instance
        AZURE_SUBSCRIPTION_ID=<subscrption_id>                  # your subscription ID
        AZURE_RESOURCE_GROUP=<resource_group_name>              # your resource group name
        AZURE_TENANT_ID=<tenant_id>                             # your tenant ID
        AZURE_CLIENT_ID=<client_id>                             # id of your client (application)
        AZURE_CLIENT_SECRET=<put yout secret>                   # your client(application) secret key
        AZURE_CONTAINER=<container_name>                        # name of container for disk blob objects, e.g. 'vhds'
        AZURE_STORAGE_ACCOUNT=<storage_account_name>            # your storage account name
        AZURE_STORAGE_ACCESS_KEY=<storage_account_access_key>   # your storage account access key

The driver and tests do not create container, instance, etc, all entities should be created before to run tests / use of libstorage.

