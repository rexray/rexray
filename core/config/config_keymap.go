package config

import (
	flag "github.com/spf13/pflag"
)

var (
	keys configKeyMap
)

type configKey struct {
	Shorthand    string
	EnvVar       string
	DefaultValue interface{}
	Description  string
}

type configKeyMap map[string]*configKey

type setVapFun func(name string, aof *interface{}) *flag.Flag

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
	c.assvp(OSDrivers, &c.OSDrivers)
	c.aivp(MinVolSize, &c.MinVolSize)
	c.abvp(RemoteManagement, &c.RemoteManagement)

	c.asvp(DockerVolumeType, &c.DockerVolumeType)
	c.aivp(DockerIOPS, &c.DockerIOPS)
	c.aivp(DockerSize, &c.DockerSize)
	c.asvp(DockerAvailabilityZone, &c.DockerAvailabilityZone)
	c.asvp(AwsAccessKey, &c.AwsAccessKey)
	c.asvp(AwsSecretKey, &c.AwsSecretKey)
	c.asvp(AwsRegion, &c.AwsRegion)

	c.asvp(RackspaceAuthURL, &c.RackspaceAuthURL)
	c.asvp(RackspaceUserID, &c.RackspaceUserID)
	c.asvp(RackspaceUserName, &c.RackspaceUserName)
	c.asvp(RackspacePassword, &c.RackspacePassword)
	c.asvp(RackspaceTenantID, &c.RackspaceTenantID)
	c.asvp(RackspaceTenantName, &c.RackspaceTenantName)
	c.asvp(RackspaceDomainID, &c.RackspaceDomainID)
	c.asvp(RackspaceDomainName, &c.RackspaceDomainName)

	c.asvp(OpenstackAuthURL, &c.OpenstackAuthURL)
	c.asvp(OpenstackUserID, &c.OpenstackUserID)
	c.asvp(OpenstackUserName, &c.OpenstackUserName)
	c.asvp(OpenstackPassword, &c.OpenstackPassword)
	c.asvp(OpenstackTenantID, &c.OpenstackTenantID)
	c.asvp(OpenstackTenantName, &c.OpenstackTenantName)
	c.asvp(OpenstackDomainID, &c.OpenstackDomainID)
	c.asvp(OpenstackDomainName, &c.OpenstackDomainName)
	c.asvp(OpenstackRegionName, &c.OpenstackRegionName)
	c.asvp(OpenstackAvailabilityZoneName, &c.OpenstackAvailabilityZoneName)

	c.asvp(ScaleIOEndpoint, &c.ScaleIOEndpoint)
	c.abvp(ScaleIOInsecure, &c.ScaleIOInsecure)
	c.abvp(ScaleIOUseCerts, &c.ScaleIOUseCerts)
	c.asvp(ScaleIOUserName, &c.ScaleIOUserName)
	c.asvp(ScaleIOPassword, &c.ScaleIoPassword)
	c.asvp(ScaleIOSystemID, &c.ScaleIOSystemID)
	c.asvp(ScaleIOSystemName, &c.ScaleIOSystemName)
	c.asvp(ScaleIOProtectionDomainID, &c.ScaleIOProtectionDomainID)
	c.asvp(ScaleIOProtectionDomainName, &c.ScaleIOProtectionDomainName)
	c.asvp(ScaleIOStoragePoolID, &c.ScaleIOStoragePoolID)
	c.asvp(ScaleIOStoragePoolName, &c.ScaleIOStoragePoolName)

	c.asvp(XtremIOEndpoint, &c.XtremIOEndpoint)
	c.asvp(XtremIOUserName, &c.XtremIOUserName)
	c.asvp(XtremIOPassword, &c.XtremIoPassword)
	c.abvp(XtremIOInsecure, &c.XtremIOInsecure)
	c.abvp(XtremIODeviceMapper, &c.XtremIODeviceMapper)
	c.abvp(XtremIOMultipath, &c.XtremIOMultipath)
	c.abvp(XtremIORemoteManagement, &c.XtremIORemoteManagement)
}

func initConfigKeyMap() {

	ck := func(envVar string, defVal interface{}, desc string) *configKey {
		return &configKey{
			EnvVar:       envVar,
			DefaultValue: defVal,
			Description:  desc,
		}
	}

	keys = configKeyMap{

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

		OSDrivers: ck("REXRAY_OSDRIVERS", []string{"linux"},
			"A space-delimited list of OS drivers to consider"),

		MinVolSize: ck("REXRAY_MINVOLSIZE", 0,
			"The minimum volume size REX-Ray is allowed to create"),

		RemoteManagement: ck("REXRAY_REMOTEMANAGEMENT", false,
			"A flag indicating whether or not to enable remote management"),

		DockerVolumeType:       ck("REXRAY_DOCKER_VOLUMETYPE", "", ""),
		DockerIOPS:             ck("REXRAY_DOCKER_IOPS", 0, ""),
		DockerSize:             ck("REXRAY_DOCKER_SIZE", 16, ""),
		DockerAvailabilityZone: ck("REXRAY_DOCKER_AVAILABILITYZONE", "", ""),

		AwsAccessKey: ck("AWS_ACCESS_KEY", "", ""),
		AwsSecretKey: ck("AWS_SECRET_KEY", "", ""),
		AwsRegion:    ck("AWS_REGION", "", ""),

		RackspaceAuthURL:    ck("OS_AUTH_URL", "", ""),
		RackspaceUserID:     ck("OS_USERID", "", ""),
		RackspaceUserName:   ck("OS_USERNAME", "", ""),
		RackspacePassword:   ck("OS_PASSWORD", "", ""),
		RackspaceTenantID:   ck("OS_TENANT_ID", "", ""),
		RackspaceTenantName: ck("OS_TENANT_NAME", "", ""),
		RackspaceDomainID:   ck("OS_DOMAIN_ID", "", ""),
		RackspaceDomainName: ck("OS_DOMAIN_NAME", "", ""),

		OpenstackAuthURL:              ck("OS_AUTH_URL", "", ""),
		OpenstackUserID:               ck("OS_USERID", "", ""),
		OpenstackUserName:             ck("OS_USERNAME", "", ""),
		OpenstackPassword:             ck("OS_PASSWORD", "", ""),
		OpenstackTenantID:             ck("OS_TENANT_ID", "", ""),
		OpenstackTenantName:           ck("OS_TENANT_NAME", "", ""),
		OpenstackDomainID:             ck("OS_DOMAIN_ID", "", ""),
		OpenstackDomainName:           ck("OS_DOMAIN_NAME", "", ""),
		OpenstackRegionName:           ck("OS_REGION_NAME", "", ""),
		OpenstackAvailabilityZoneName: ck("OS_AVAILABILITY_ZONE_NAME", "", ""),

		ScaleIOEndpoint:             ck("GOSCALEIO_ENDPOINT", "", ""),
		ScaleIOInsecure:             ck("GOSCALEIO_INSECURE", false, ""),
		ScaleIOUseCerts:             ck("GOSCALEIO_USECERTS", true, ""),
		ScaleIOUserName:             ck("GOSCALEIO_USERNAME", "", ""),
		ScaleIOPassword:             ck("GOSCALEIO_PASSWORD", "", ""),
		ScaleIOSystemID:             ck("GOSCALEIO_SYSTEMID", "", ""),
		ScaleIOSystemName:           ck("GOSCALEIO_SYSTEMNAME", "", ""),
		ScaleIOProtectionDomainID:   ck("GOSCALEIO_PROTECTIONDOMAINID", "", ""),
		ScaleIOProtectionDomainName: ck("GOSCALEIO_PROTECTIONDOMAIN", "", ""),
		ScaleIOStoragePoolID:        ck("GOSCALEIO_STORAGEPOOLID", "", ""),
		ScaleIOStoragePoolName:      ck("GOSCALEIO_STORAGEPOOL", "", ""),

		XtremIOEndpoint:         ck("GOXTREMIO_ENDPOINT", "", ""),
		XtremIOUserName:         ck("GOXTREMIO_USERNAME", "", ""),
		XtremIOPassword:         ck("GOXTREMIO_PASSWORD", "", ""),
		XtremIOInsecure:         ck("GOXTREMIO_INSECURE", false, ""),
		XtremIODeviceMapper:     ck("GOXTREMIO_DM", "", ""),
		XtremIOMultipath:        ck("GOXTREMIO_MULTIPATH", "", ""),
		XtremIORemoteManagement: ck("GOXTREMIO_REMOTEMANAGEMENT", false, ""),
	}

	keys[Host].Shorthand = "h"
	keys[LogLevel].Shorthand = "l"
}
