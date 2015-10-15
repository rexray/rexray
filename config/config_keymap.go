package config

import (
	flag "github.com/spf13/pflag"
)

var (
	keys ConfigKeyMap
)

type ConfigKey struct {
	Shorthand    string
	EnvVar       string
	DefaultValue interface{}
	Description  string
}

type ConfigKeyMap map[string]*ConfigKey

type SetVapFun func(name string, aof *interface{}) *flag.Flag

func (c *Config) gbvp(name string, aof *bool) *flag.Flag {
	return c.boolVarP(c.GlobalFlags, name, aof)
}
func (c *Config) abvp(name string, aof *bool) *flag.Flag {
	return c.boolVarP(c.AdditionalFlags, name, aof)
}
func (c *Config) boolVarP(fset *flag.FlagSet, name string, aof *bool) *flag.Flag {
	i := keys[name]
	fset.BoolVarP(aof, name, i.Shorthand, c.Viper.GetBool(name), i.Description)
	return fset.Lookup(name)
}

func (c *Config) givp(name string, aof *int) *flag.Flag {
	return c.intVarP(c.GlobalFlags, name, aof)
}
func (c *Config) aivp(name string, aof *int) *flag.Flag {
	return c.intVarP(c.AdditionalFlags, name, aof)
}
func (c *Config) intVarP(fset *flag.FlagSet, name string, aof *int) *flag.Flag {
	i := keys[name]
	fset.IntVarP(aof, name, i.Shorthand, c.Viper.GetInt(name), i.Description)
	return fset.Lookup(name)
}

func (c *Config) gsvp(name string, aof *string) *flag.Flag {
	return c.stringVarP(c.GlobalFlags, name, aof)
}
func (c *Config) asvp(name string, aof *string) *flag.Flag {
	return c.stringVarP(c.AdditionalFlags, name, aof)
}
func (c *Config) stringVarP(fset *flag.FlagSet, name string, aof *string) *flag.Flag {
	i := keys[name]
	fset.StringVarP(aof, name, i.Shorthand, c.Viper.GetString(name), i.Description)
	return fset.Lookup(name)
}

func (c *Config) gssvp(name string, aof *[]string) *flag.Flag {
	return c.stringSliceVarP(c.GlobalFlags, name, aof)
}
func (c *Config) assvp(name string, aof *[]string) *flag.Flag {
	return c.stringSliceVarP(c.AdditionalFlags, name, aof)
}
func (c *Config) stringSliceVarP(fset *flag.FlagSet, name string, aof *[]string) *flag.Flag {
	i := keys[name]
	fset.StringSliceVarP(
		aof, name, i.Shorthand, c.Viper.GetStringSlice(name), i.Description)
	return fset.Lookup(name)
}

func (c *Config) initConfigKeys() {

	for k, v := range keys {
		c.Viper.BindEnv(k, v.EnvVar)
		c.Viper.SetDefault(k, v.DefaultValue)
	}

	c.gsvp(Host, &c.Host)
	c.gsvp(LogLevel, &c.LogLevel)
	c.assvp(StorageDrivers, &c.StorageDrivers)
	c.assvp(VolumeDrivers, &c.VolumeDrivers)
	c.assvp(OsDrivers, &c.OsDrivers)
	c.aivp(MinVolSize, &c.MinVolSize)
	c.abvp(RemoteManagement, &c.RemoteManagement)

	c.asvp(DockerVolumeType, &c.DockerVolumeType)
	c.aivp(DockerIops, &c.DockerIops)
	c.aivp(DockerSize, &c.DockerSize)
	c.asvp(DockerAvailabilityZone, &c.DockerAvailabilityZone)
	c.asvp(AwsAccessKey, &c.AwsAccessKey)
	c.asvp(AwsSecretKey, &c.AwsSecretKey)
	c.asvp(AwsRegion, &c.AwsRegion)

	c.asvp(RackspaceAuthUrl, &c.RackspaceAuthUrl)
	c.asvp(RackspaceUserId, &c.RackspaceUserId)
	c.asvp(RackspaceUserName, &c.RackspaceUserName)
	c.asvp(RackspacePassword, &c.RackspacePassword)
	c.asvp(RackspaceTenantId, &c.RackspaceTenantId)
	c.asvp(RackspaceTenantName, &c.RackspaceTenantName)
	c.asvp(RackspaceDomainId, &c.RackspaceDomainId)
	c.asvp(RackspaceDomainName, &c.RackspaceDomainName)

	c.asvp(OpenstackAuthUrl, &c.OpenstackAuthUrl)
	c.asvp(OpenstackUserId, &c.OpenstackUserId)
	c.asvp(OpenstackUserName, &c.OpenstackUserName)
	c.asvp(OpenstackPassword, &c.OpenstackPassword)
	c.asvp(OpenstackTenantId, &c.OpenstackTenantId)
	c.asvp(OpenstackTenantName, &c.OpenstackTenantName)
	c.asvp(OpenstackDomainId, &c.OpenstackDomainId)
	c.asvp(OpenstackDomainName, &c.OpenstackDomainName)
	c.asvp(OpenstackRegionName, &c.OpenstackRegionName)
	c.asvp(OpenstackAvailabilityZoneName, &c.OpenstackAvailabilityZoneName)

	c.asvp(ScaleIoEndpoint, &c.ScaleIoEndpoint)
	c.abvp(ScaleIoInsecure, &c.ScaleIoInsecure)
	c.abvp(ScaleIoUseCerts, &c.ScaleIoUseCerts)
	c.asvp(ScaleIoUserName, &c.ScaleIoUserName)
	c.asvp(ScaleIoPassword, &c.ScaleIoPassword)
	c.asvp(ScaleIoSystemId, &c.ScaleIoSystemId)
	c.asvp(ScaleIoSystemName, &c.ScaleIoSystemName)
	c.asvp(ScaleIoProtectionDomainId, &c.ScaleIoProtectionDomainId)
	c.asvp(ScaleIoProtectionDomainName, &c.ScaleIoProtectionDomainName)
	c.asvp(ScaleIoStoragePoolId, &c.ScaleIoStoragePoolId)
	c.asvp(ScaleIoStoragePoolName, &c.ScaleIoStoragePoolName)

	c.asvp(XtremIoEndpoint, &c.XtremIoEndpoint)
	c.asvp(XtremIoUserName, &c.XtremIoUserName)
	c.asvp(XtremIoPassword, &c.XtremIoPassword)
	c.abvp(XtremIoInsecure, &c.XtremIoInsecure)
	c.asvp(XtremIoDeviceMapper, &c.XtremIoDeviceMapper)
	c.asvp(XtremIoMultipath, &c.XtremIoMultipath)
	c.abvp(XtremIoRemoteManagement, &c.XtremIoRemoteManagement)
}

func initConfigKeyMap() {

	ck := func(envVar string, defVal interface{}, desc string) *ConfigKey {
		return &ConfigKey{
			EnvVar:       envVar,
			DefaultValue: defVal,
			Description:  desc,
		}
	}

	keys = ConfigKeyMap{

		//Host: ck("REXRAY_HOST", "unix:///var/run/rexray.sock",
		//	"The REX-Ray service address"),
		Host: ck("REXRAY_HOST", "tcp://:7979",
			"The REX-Ray service address"),

		LogLevel: ck("REXRAY_LOGLEVEL", "info",
			"The log level (panic, fatal, error, warn, info, debug)"),

		StorageDrivers: ck("REXRAY_STORAGEDRIVERS", []string{},
			"A space-delimited list of storage drivers to consider"),

		VolumeDrivers: ck("REXRAY_VOLUMEDRIVERS", []string{"docker"},
			"A space-delimited list of volume drivers to consider"),

		OsDrivers: ck("REXRAY_OSDRIVERS", []string{"linux"},
			"A space-delimited list of OS drivers to consider"),

		MinVolSize: ck("REXRAY_MINVOLSIZE", 0,
			"The minimum volume size REX-Ray is allowed to create"),

		RemoteManagement: ck("REXRAY_REMOTEMANAGEMENT", false,
			"A flag indicating whether or not to enable remote management"),

		DockerVolumeType:       ck("REXRAY_DOCKER_VOLUMETYPE", "", ""),
		DockerIops:             ck("REXRAY_DOCKER_IOPS", 0, ""),
		DockerSize:             ck("REXRAY_DOCKER_SIZE", 0, ""),
		DockerAvailabilityZone: ck("REXRAY_DOCKER_AVAILABILITYZONE", "", ""),

		AwsAccessKey: ck("AWS_ACCESS_KEY", "", ""),
		AwsSecretKey: ck("AWS_SECRET_KEY", "", ""),
		AwsRegion:    ck("AWS_REGION", "", ""),

		RackspaceAuthUrl:    ck("OS_AUTH_URL", "", ""),
		RackspaceUserId:     ck("OS_USERID", "", ""),
		RackspaceUserName:   ck("OS_USERNAME", "", ""),
		RackspacePassword:   ck("OS_PASSWORD", "", ""),
		RackspaceTenantId:   ck("OS_TENANT_ID", "", ""),
		RackspaceTenantName: ck("OS_TENANT_NAME", "", ""),
		RackspaceDomainId:   ck("OS_DOMAIN_ID", "", ""),
		RackspaceDomainName: ck("OS_DOMAIN_NAME", "", ""),

		OpenstackAuthUrl:              ck("OS_AUTH_URL", "", ""),
		OpenstackUserId:               ck("OS_USERID", "", ""),
		OpenstackUserName:             ck("OS_USERNAME", "", ""),
		OpenstackPassword:             ck("OS_PASSWORD", "", ""),
		OpenstackTenantId:             ck("OS_TENANT_ID", "", ""),
		OpenstackTenantName:           ck("OS_TENANT_NAME", "", ""),
		OpenstackDomainId:             ck("OS_DOMAIN_ID", "", ""),
		OpenstackDomainName:           ck("OS_DOMAIN_NAME", "", ""),
		OpenstackRegionName:           ck("OS_REGION_NAME", "", ""),
		OpenstackAvailabilityZoneName: ck("OS_AVAILABILITY_ZONE_NAME", "", ""),

		ScaleIoEndpoint:             ck("GOSCALEIO_ENDPOINT", "", ""),
		ScaleIoInsecure:             ck("GOSCALEIO_INSECURE", false, ""),
		ScaleIoUseCerts:             ck("GOSCALEIO_USECERTS", true, ""),
		ScaleIoUserName:             ck("GOSCALEIO_USERNAME", "", ""),
		ScaleIoPassword:             ck("GOSCALEIO_PASSWORD", "", ""),
		ScaleIoSystemId:             ck("GOSCALEIO_SYSTEMID", "", ""),
		ScaleIoSystemName:           ck("GOSCALEIO_SYSTEMNAME", "", ""),
		ScaleIoProtectionDomainId:   ck("GOSCALEIO_PROTECTIONDOMAINID", "", ""),
		ScaleIoProtectionDomainName: ck("GOSCALEIO_PROTECTIONDOMAIN", "", ""),
		ScaleIoStoragePoolId:        ck("GOSCALEIO_STORAGEPOOLID", "", ""),
		ScaleIoStoragePoolName:      ck("GOSCALEIO_STORAGEPOOL", "", ""),

		XtremIoEndpoint:         ck("GOXTREMIO_ENDPOINT", "", ""),
		XtremIoUserName:         ck("GOXTREMIO_USERNAME", "", ""),
		XtremIoPassword:         ck("GOXTREMIO_PASSWORD", "", ""),
		XtremIoInsecure:         ck("GOXTREMIO_INSECURE", false, ""),
		XtremIoDeviceMapper:     ck("GOXTREMIO_DM", "", ""),
		XtremIoMultipath:        ck("GOXTREMIO_MULTIPATH", "", ""),
		XtremIoRemoteManagement: ck("GOXTREMIO_REMOTEMANAGEMENT", false, ""),
	}

	keys[Host].Shorthand = "h"
	keys[LogLevel].Shorthand = "l"
}
