package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/emccode/rexray/util"
)

var (
	envVarRx *regexp.Regexp
)

func init() {
	var envVarRxErr error
	envVarRx, envVarRxErr = regexp.Compile(`^\s*([^#=]+)=(.+)$`)
	if envVarRxErr != nil {
		panic(envVarRxErr)
	}
	loadEtcEnvironment()
	initConfigKeyMap()
}

type Config struct {
	GlobalFlags     *pflag.FlagSet `json:"-"`
	AdditionalFlags *pflag.FlagSet `json:"-"`
	Viper           *viper.Viper   `json:"-"`
	Keys            ConfigKeyMap   `json:"-"`

	Host             string `json:"-"`
	LogLevel         string
	StorageDrivers   []string
	VolumeDrivers    []string
	OsDrivers        []string
	MinVolSize       int
	RemoteManagement bool

	DockerVolumeType       string
	DockerIops             int
	DockerSize             int
	DockerAvailabilityZone string

	AwsAccessKey string
	AwsSecretKey string
	AwsRegion    string

	RackspaceAuthUrl    string
	RackspaceUserId     string
	RackspaceUserName   string
	RackspacePassword   string
	RackspaceTenantId   string
	RackspaceTenantName string
	RackspaceDomainId   string
	RackspaceDomainName string

	ScaleIoEndpoint             string
	ScaleIoInsecure             bool
	ScaleIoUseCerts             bool
	ScaleIoUserName             string
	ScaleIoPassword             string
	ScaleIoSystemId             string
	ScaleIoSystemName           string
	ScaleIoProtectionDomainId   string
	ScaleIoProtectionDomainName string
	ScaleIoStoragePoolId        string
	ScaleIoStoragePoolName      string

	XtremIoEndpoint         string
	XtremIoUserName         string
	XtremIoPassword         string
	XtremIoInsecure         bool
	XtremIoDeviceMapper     string
	XtremIoMultipath        string
	XtremIoRemoteManagement bool
}

func New() *Config {
	log.Debug("initializing configuration")

	c := &Config{
		Viper: viper.New(),
	}

	cfgName := "config"
	cfgType := "yml"
	etcRexRay := util.EtcDirPath()
	usrRexRay := fmt.Sprintf("%s/.rexray", util.HomeDir())
	etcRexRayFile := fmt.Sprintf("%s/%s.%s", etcRexRay, cfgName, cfgType)
	usrRexRayFile := fmt.Sprintf("%s/%s.%s", usrRexRay, cfgName, cfgType)

	c.Viper.SetConfigName(cfgName)
	c.Viper.SetConfigType(cfgType)
	c.Viper.SetTypeByDefaultValue(true)

	log.WithFields(log.Fields{
		"name": cfgName,
		"type": cfgType}).Debug("set config name and type")

	c.Viper.AddConfigPath(etcRexRay)
	c.Viper.AddConfigPath(usrRexRay)

	log.WithFields(log.Fields{
		"global": etcRexRay,
		"user":   usrRexRay}).Debug("added config paths")

	if util.FileExists(etcRexRayFile) || util.FileExists(usrRexRayFile) {

		log.WithFields(log.Fields{
			"global": etcRexRayFile,
			"user":   usrRexRayFile}).Debug(
			"reading configuration file(s) from default path(s)")

		if readCfgErr := c.Viper.ReadInConfig(); readCfgErr != nil {
			log.WithFields(log.Fields{
				"global": etcRexRayFile,
				"user":   usrRexRayFile,
				"error":  readCfgErr}).Error(
				"error reading configuration file(s) from default path(s)")
		}
	}

	c.initConfigKeys()

	return c
}

func (c *Config) Copy() (*Config, error) {
	newC := New()
	mErr := c.Viper.Marshal(newC)
	if mErr != nil {
		return nil, mErr
	}
	return newC, nil
}

func FromJson(from string) (*Config, error) {
	c := New()
	umErr := json.Unmarshal([]byte(from), c)
	if umErr != nil {
		return nil, umErr
	}
	c.Sync()
	return c, nil
}

func (c *Config) ToJson() (string, error) {
	buf, bufErr := json.MarshalIndent(c, "", "  ")
	if bufErr != nil {
		return "", bufErr
	}
	return string(buf), nil
}

func (c *Config) ReadConfigFile(filePath string) error {

	buf, bufErr := ioutil.ReadFile(filePath)
	if bufErr != nil {
		log.WithFields(
			log.Fields{"filePath": filePath, "error": bufErr}).Error(
			"error reading config file")
		return bufErr
	}

	if cfgErr := c.Viper.ReadConfig(bytes.NewReader(buf)); cfgErr != nil {
		log.WithFields(
			log.Fields{"filePath": filePath, "error": cfgErr}).Error(
			"error reading config file")
		return cfgErr
	}

	log.WithField("filePath", filePath).Info("read config file")

	return nil
}

func (c *Config) EnvVars() []string {
	evArr := make([]string, len(c.Keys))
	for k, v := range c.Keys {
		evArr = append(evArr,
			fmt.Sprintf("%s=%s", v.EnvVar, c.Viper.GetString(k)))
	}
	return evArr
}

func (c *Config) Sync() {

	w := c.Viper.Set

	w(Host, c.Host)
	w(LogLevel, c.LogLevel)
	w(StorageDrivers, c.StorageDrivers)
	w(VolumeDrivers, c.VolumeDrivers)
	w(OsDrivers, c.OsDrivers)
	w(MinVolSize, c.MinVolSize)
	w(RemoteManagement, c.RemoteManagement)

	w(DockerVolumeType, c.DockerVolumeType)
	w(DockerIops, c.DockerIops)
	w(DockerSize, c.DockerSize)
	w(DockerAvailabilityZone, c.DockerAvailabilityZone)
	w(AwsAccessKey, c.AwsAccessKey)
	w(AwsSecretKey, c.AwsSecretKey)
	w(AwsRegion, c.AwsRegion)

	w(RackspaceAuthUrl, c.RackspaceAuthUrl)
	w(RackspaceUserId, c.RackspaceUserId)
	w(RackspaceUserName, c.RackspaceUserName)
	w(RackspacePassword, c.RackspacePassword)
	w(RackspaceTenantId, c.RackspaceTenantId)
	w(RackspaceTenantName, c.RackspaceTenantName)
	w(RackspaceDomainId, c.RackspaceDomainId)
	w(RackspaceDomainName, c.RackspaceDomainName)

	w(ScaleIoEndpoint, c.ScaleIoEndpoint)
	w(ScaleIoInsecure, c.ScaleIoInsecure)
	w(ScaleIoUseCerts, c.ScaleIoUseCerts)
	w(ScaleIoUserName, c.ScaleIoUserName)
	w(ScaleIoPassword, c.ScaleIoPassword)
	w(ScaleIoSystemId, c.ScaleIoSystemId)
	w(ScaleIoSystemName, c.ScaleIoSystemName)
	w(ScaleIoProtectionDomainId, c.ScaleIoProtectionDomainId)
	w(ScaleIoProtectionDomainName, c.ScaleIoProtectionDomainName)
	w(ScaleIoStoragePoolId, c.ScaleIoStoragePoolId)
	w(ScaleIoStoragePoolName, c.ScaleIoStoragePoolName)

	w(XtremIoEndpoint, c.XtremIoEndpoint)
	w(XtremIoUserName, c.XtremIoUserName)
	w(XtremIoPassword, c.XtremIoPassword)
	w(XtremIoInsecure, c.XtremIoInsecure)
	w(XtremIoDeviceMapper, c.XtremIoDeviceMapper)
	w(XtremIoMultipath, c.XtremIoMultipath)
	w(XtremIoRemoteManagement, c.XtremIoRemoteManagement)
}

func loadEtcEnvironment() {
	lr := util.LineReader("/etc/environment")
	if lr == nil {
		return
	}
	for l := range lr {
		m := envVarRx.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		os.Setenv(m[1], m[2])
	}
}
