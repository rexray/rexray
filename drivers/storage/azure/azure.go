// +build !libstorage_storage_driver libstorage_storage_driver_azure

package azure

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

const (
	// Name is the provider's name.
	Name = "azure"

	// TagDelimiter separates tags from volume or snapshot names
	TagDelimiter = "/"

	// DefaultUseHTTPS - Use https prefix by default
	// or not for Azure URI's
	DefaultUseHTTPS = true

	// TenantIDKey is a Directory ID from Azure
	TenantIDKey = "tenantID"
	// ClientIDKey is an Application ID from Azure
	ClientIDKey = "clientID"
	// ClientSecretKey is a secret of the application
	ClientSecretKey = "clientSecret"
	// CertPathKey is a path to application certificate in case of
	// authorization via certificate
	CertPathKey = "certPath"

	// StorageAccountKey is a name of storage account
	StorageAccountKey = "storageAccount"
	// StorageAccessKey is an access key of storage account
	StorageAccessKey = "storageAccessKey"
	// TODO: add option to pass StorageURI

	// SubscriptionIDKey is an ID of subscription
	SubscriptionIDKey = "subscriptionID"
	// ResourceGroupKey is a name of resource group
	ResourceGroupKey = "resourceGroup"
	// ContainerKey is a name of container in the storage account
	// ('vhds' by default)
	ContainerKey = "container"
	// UseHTTPSKey is a flag about use https or not for making Azure URI's
	UseHTTPSKey = "useHTTPS"
	// TagKey is a tag key
	TagKey = "tag"
)

const (
	// ConfigAzure is a config key
	ConfigAzure = Name

	// ConfigAzureSubscriptionIDKey is a config key
	ConfigAzureSubscriptionIDKey = ConfigAzure + "." + SubscriptionIDKey

	// ConfigAzureResourceGroupKey is a config key
	ConfigAzureResourceGroupKey = ConfigAzure + "." + ResourceGroupKey

	// ConfigAzureTenantIDKey is a config key
	ConfigAzureTenantIDKey = ConfigAzure + "." + TenantIDKey

	// ConfigAzureStorageAccountKey is a config key
	ConfigAzureStorageAccountKey = ConfigAzure + "." + StorageAccountKey

	// ConfigAzureStorageAccessKeyKey is a config key
	ConfigAzureStorageAccessKeyKey = ConfigAzure + "." + StorageAccessKey

	// ConfigAzureContainerKey is a config key
	ConfigAzureContainerKey = ConfigAzure + "." + ContainerKey

	// ConfigAzureClientIDKey is a config key
	ConfigAzureClientIDKey = ConfigAzure + "." + ClientIDKey

	// ConfigAzureClientSecretKey is a config key
	ConfigAzureClientSecretKey = ConfigAzure + "." + ClientSecretKey

	// ConfigAzureCertPathKey is a config key
	ConfigAzureCertPathKey = ConfigAzure + "." + CertPathKey

	// ConfigAzureUseHTTPSKey is a config key
	ConfigAzureUseHTTPSKey = ConfigAzure + "." + UseHTTPSKey

	// ConfigAzureTagKey is a config key
	ConfigAzureTagKey = ConfigAzure + "." + TagKey
)

func init() {
	r := gofigCore.NewRegistration("Azure")
	r.Key(gofig.String, "", "", "", ConfigAzureSubscriptionIDKey)
	r.Key(gofig.String, "", "", "", ConfigAzureResourceGroupKey)
	r.Key(gofig.String, "", "", "", ConfigAzureTenantIDKey)
	r.Key(gofig.String, "", "", "", ConfigAzureStorageAccountKey)
	r.Key(gofig.String, "", "", "", ConfigAzureContainerKey)
	r.Key(gofig.String, "", "", "", ConfigAzureClientIDKey)
	r.Key(gofig.String, "", "", "", ConfigAzureClientSecretKey)
	r.Key(gofig.String, "", "", "", ConfigAzureCertPathKey)
	r.Key(gofig.Bool, "", DefaultUseHTTPS, "", ConfigAzureUseHTTPSKey)
	r.Key(gofig.String, "", "",
		"Tag prefix for Azure naming", ConfigAzureTagKey)

	gofigCore.Register(r)
}
