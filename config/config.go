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

	"github.com/emccode/rexray/errors"
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

type secureConfig struct {
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
	AwsRegion    string

	RackspaceAuthUrl    string
	RackspaceUserId     string
	RackspaceUserName   string
	RackspaceTenantId   string
	RackspaceTenantName string
	RackspaceDomainId   string
	RackspaceDomainName string

	ScaleIoEndpoint             string
	ScaleIoInsecure             bool
	ScaleIoUseCerts             bool
	ScaleIoUserName             string
	ScaleIoSystemId             string
	ScaleIoSystemName           string
	ScaleIoProtectionDomainId   string
	ScaleIoProtectionDomainName string
	ScaleIoStoragePoolId        string
	ScaleIoStoragePoolName      string

	XtremIoEndpoint         string
	XtremIoUserName         string
	XtremIoInsecure         bool
	XtremIoDeviceMapper     string
	XtremIoMultipath        string
	XtremIoRemoteManagement bool
}

type JsonMarshalStrategy int

const (
	JsonMarshalSecure JsonMarshalStrategy = iota
	JsonMarshalPlainText
)

type plainTextConfig struct {
	secureConfig

	AwsSecretKey      string
	RackspacePassword string
	ScaleIoPassword   string
	XtremIoPassword   string
}

// Config contains the configuration information
type Config struct {
	plainTextConfig

	GlobalFlags     *flag.FlagSet `json:"-"`
	AdditionalFlags *flag.FlagSet `json:"-"`
	Viper           *viper.Viper  `json:"-"`
	Host            string        `json:"-"`

	jsonMarshalStrategy JsonMarshalStrategy
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
		plainTextConfig: plainTextConfig{
			secureConfig: secureConfig{},
		},
		Viper:               viper.New(),
		GlobalFlags:         &flag.FlagSet{},
		AdditionalFlags:     &flag.FlagSet{},
		jsonMarshalStrategy: JsonMarshalSecure,
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

func (c *Config) JsonMarshalStrategy() JsonMarshalStrategy {
	return c.jsonMarshalStrategy
}

func (c *Config) SetJsonMarshalStrategy(s JsonMarshalStrategy) {
	c.jsonMarshalStrategy = s
}

func (c *Config) Copy() (*Config, error) {
	newC := New()
	if err := c.Viper.Unmarshal(&newC.plainTextConfig); err != nil {
		return nil, err
	}
	if err := c.Viper.Unmarshal(&newC.secureConfig); err != nil {
		return nil, err
	}
	return newC, nil
}

func FromJson(from string) (*Config, error) {
	c := New()
	if err := json.Unmarshal([]byte(from), c); err != nil {
		return nil, err
	}
	c.Sync()
	return c, nil
}

func (c *Config) ToJson() (string, error) {
	buf, err := c.marshalJSON(JsonMarshalPlainText)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (c *Config) ToSecureJson() (string, error) {
	buf, err := c.marshalJSON(JsonMarshalSecure)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// The implementation of the encoding/json.Marshaller interface. It allows
// this type to provide its own marshalling routine.
func (c *Config) MarshalJSON() ([]byte, error) {
	return c.marshalJSON(c.jsonMarshalStrategy)
}

func (c *Config) marshalJSON(s JsonMarshalStrategy) ([]byte, error) {
	switch s {
	case JsonMarshalSecure:
		return json.MarshalIndent(c.secureConfig, "", "  ")
	case JsonMarshalPlainText:
		return json.MarshalIndent(c.plainTextConfig, "", "  ")
	}

	return nil, errors.WithField(
		"strategy", s, "unknown json marshalling strategy")
}

func (c *Config) ReadConfig(in io.Reader) error {

	if err := c.Viper.ReadConfigNoNil(in); err != nil {
		return err
	}

	if err := c.Viper.Unmarshal(&c.plainTextConfig); err != nil {
		return err
	}

	if err := c.Viper.Unmarshal(&c.secureConfig); err != nil {
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
	evArr := make([]string, len(keys))
	for k, v := range keys {
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
