package cinder

import (
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
	
	"github.com/rexray/rexray/libstorage/api/types"

)

const (
	// Name is the provider's name.
	Name string = "cinder"

	// ConfigAuthURL is the config key for the Identity Auth URL
	ConfigAuthURL = Name + ".authURL"

	// ConfigUserID is the config key for the user ID
	ConfigUserID = Name + ".userID"

	// ConfigUserName is the config key for the user name
	ConfigUserName = Name + ".userName"

	// ConfigPassword is the config key for the user password
	ConfigPassword = Name + ".password"

	// ConfigTokenID is the config key for the token ID
	ConfigTokenID = Name + ".tokenID"

	// ConfigTenantID is the config key for the tenant ID
	ConfigTenantID = Name + ".tenantID"

	// ConfigTenantName is the config key for the tenant name
	ConfigTenantName = Name + ".tenantName"

	// ConfigDomainID is the config key for the domain ID
	ConfigDomainID = Name + ".domainID"

	// ConfigDomainName is the config key for the domain name
	ConfigDomainName = Name + ".domainName"

	// ConfigRegionName is the config key for the region name
	ConfigRegionName = Name + ".regionName"

	// ConfigAvailabilityZoneName is the config key for the availability
	// zone name
	ConfigAvailabilityZoneName = Name + ".availabilityZoneName"

	// ConfigTrustID is the config key for the trust ID
	ConfigTrustID = Name + ".trustID"

	// ConfigAttachTimeout is the config key for the attach timeout
	ConfigAttachTimeout = Name + ".attachTimeout"

	// ConfigDeleteTimeout is the config key for the delete timeout
	ConfigDeleteTimeout = Name + ".deleteTimeout"

	// ConfigCreateTimeout is the config key for the create timeout
	ConfigCreateTimeout = Name + ".createTimeout"

	// ConfigSnapshotTimeout is the config key for the snapshot timeout
	ConfigSnapshotTimeout = Name + ".snapshotTimeout"

	// ConfigCACert is the config key for custom CA certificate (usually for self signed use case)
	ConfigCACert = Name + ".CACert"

	// ConfigInsecure is the config key to disable TLS verification of the server identity
	ConfigInsecure = Name + ".insecure"

	// ConfigDevicePattern is the config key to specify the device name pattern returned by Cinder
	ConfigDevicePattern = Name + ".devicePattern" 
	
	// ConfigHostPattern is the config key to specify de the device name pattern used by the host
	ConfigHostPattern = Name + ".hostPattern" 

	// ConfigMappingType is the device mapping type: ebs or virtio 
	ConfigMappingType = Name + ".mappingType"
)

func init() {
	r := gofigCore.NewRegistration("Cinder")
	r.Key(gofig.String, "", "", "", ConfigAuthURL)
	r.Key(gofig.String, "", "", "", ConfigUserID)
	r.Key(gofig.String, "", "", "", ConfigUserName)
	r.Key(gofig.String, "", "", "", ConfigPassword)
	r.Key(gofig.String, "", "", "", ConfigTokenID)
	r.Key(gofig.String, "", "", "", ConfigTenantID)
	r.Key(gofig.String, "", "", "", ConfigTenantName)
	r.Key(gofig.String, "", "", "", ConfigDomainID)
	r.Key(gofig.String, "", "", "", ConfigDomainName)
	r.Key(gofig.String, "", "", "", ConfigRegionName)
	r.Key(gofig.String, "", "", "", ConfigAvailabilityZoneName)
	r.Key(gofig.String, "", "", "", ConfigTrustID)
	r.Key(gofig.String, "", "1m", "", ConfigAttachTimeout)
	r.Key(gofig.String, "", "10m", "", ConfigDeleteTimeout)
	r.Key(gofig.String, "", "10m", "", ConfigCreateTimeout)
	r.Key(gofig.String, "", "10m", "", ConfigSnapshotTimeout)
	r.Key(gofig.String, "", "", "", ConfigCACert)
	r.Key(gofig.Bool, "", false, "", ConfigInsecure)
	r.Key(gofig.String, "", "", "", ConfigMappingType)
	r.Key(gofig.String, "", "/dev/sd", "", ConfigDevicePattern)
	r.Key(gofig.String, "", "/dev/xvd", "", ConfigHostPattern)
	gofigCore.Register(r)
}


// Driver extend driver type so it possible to resolve the correct device name attached to an host
type Driver interface{
	ResolveDeviceName(ctx types.Context, device string, volumeID string) string
}




