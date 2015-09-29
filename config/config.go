package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	log "github.com/Sirupsen/logrus"
	flag "github.com/spf13/pflag"
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

// Config contains the configuration information
type Config struct {
	GlobalFlags     *flag.FlagSet `json:"-"`
	AdditionalFlags *flag.FlagSet `json:"-"`
	Viper           *viper.Viper  `json:"-"`
	Keys            ConfigKeyMap  `json:"-"`

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

// New initializes a new instance of a Config struct
func New() *Config {
	return NewConfig(true, true, "config", "yml")
}

func NewConfig(
	loadGlobalConfig, loadUserConfig bool, 
	configName, configType string) *Config {
	
	log.Debug("initializing configuration")

	c := &Config{
		Viper:           viper.New(),
		GlobalFlags:     &flag.FlagSet{},
		AdditionalFlags: &flag.FlagSet{},
	}
	c.Viper.SetTypeByDefaultValue(true)
	c.Viper.SetConfigName(configName)
	c.Viper.SetConfigType(configType)

	cfgFile := fmt.Sprintf("%s.%s", configName, configType)
	etcRexRayFile := fmt.Sprintf("%s/%s", util.EtcDirPath(), cfgFile)
	usrRexRayFile := fmt.Sprintf("%s/.rexray/%s", util.HomeDir(), cfgFile)

	if loadGlobalConfig && util.FileExists(etcRexRayFile) {
		log.WithField("path", etcRexRayFile).Debug("loading global config file")
		if err := c.ReadConfigFile(etcRexRayFile); err != nil {
			log.WithFields(log.Fields{
				"path":  etcRexRayFile,
				"error": err}).Error(
				"error reading global config file")
		}
	}

	if loadUserConfig && util.FileExists(usrRexRayFile) {
		log.WithField("path", usrRexRayFile).Debug("loading user config file")
		if err := c.ReadConfigFile(usrRexRayFile); err != nil {
			log.WithFields(log.Fields{
				"path":  usrRexRayFile,
				"error": err}).Error(
				"error reading user config file")
		}
	}

	c.initConfigKeys()

	return c
	
}

func (c *Config) Copy() (*Config, error) {
	newC := New()
	mErr := c.Viper.Unmarshal(newC)
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

func (c *Config) ReadConfig(in io.Reader) error {

	if err := c.Viper.ReadConfigNoNil(in); err != nil {
		return err
	}

	if err := c.Viper.Unmarshal(c); err != nil {
		return err
	}

	for key, _ := range keys {
		c.updateFlag(key, c.GlobalFlags)
		c.updateFlag(key, c.AdditionalFlags)
	}

	return nil
}

func (c *Config) ReadConfigFile(filePath string) error {
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return c.ReadConfig(bytes.NewBuffer(buf))
}

func (c *Config) updateFlag(name string, flags *flag.FlagSet) {
	if f := flags.Lookup(name); f != nil {
		val := c.Viper.Get(name)
		strVal := fmt.Sprintf("%v", val)
		f.DefValue = strVal
	}
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
		if m == nil || len(m) < 3 || os.Getenv(m[1]) != "" {
			continue
		}
		os.Setenv(m[1], m[2])
	}
}
