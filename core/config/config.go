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

	"github.com/emccode/rexray/core/errors"
	"github.com/emccode/rexray/util"
)

var (
	envVarRx *regexp.Regexp
)

func init() {
	envVarRx = regexp.MustCompile(`^\s*([^#=]+)=(.+)$`)
	loadEtcEnvironment()
	initConfigKeyMap()
}

// JSONMarshalStrategy is a JSON marshalling strategy
type JSONMarshalStrategy int

const (
	// JSONMarshalSecure indicates that the secure fields should be omitted
	JSONMarshalSecure JSONMarshalStrategy = iota

	// JSONMarshalPlainText indicates that all fields should be included
	JSONMarshalPlainText
)

type secureConfig struct {
	LogLevel         string
	StorageDrivers   []string
	VolumeDrivers    []string
	OSDrivers        []string
	MinVolSize       int
	RemoteManagement bool

	DockerVolumeType       string
	DockerIOPS             int
	DockerSize             int
	DockerAvailabilityZone string

	AwsAccessKey string
	AwsRegion    string

	RackspaceAuthURL    string
	RackspaceUserID     string
	RackspaceUserName   string
	RackspaceTenantID   string
	RackspaceTenantName string
	RackspaceDomainID   string
	RackspaceDomainName string

	OpenstackAuthURL              string
	OpenstackUserID               string
	OpenstackUserName             string
	OpenstackTenantID             string
	OpenstackTenantName           string
	OpenstackDomainID             string
	OpenstackDomainName           string
	OpenstackRegionName           string
	OpenstackAvailabilityZoneName string

	ScaleIOEndpoint             string
	ScaleIOInsecure             bool
	ScaleIOUseCerts             bool
	ScaleIOUserName             string
	ScaleIOSystemID             string
	ScaleIOSystemName           string
	ScaleIOProtectionDomainID   string
	ScaleIOProtectionDomainName string
	ScaleIOStoragePoolID        string
	ScaleIOStoragePoolName      string

	XtremIOEndpoint         string
	XtremIOUserName         string
	XtremIOInsecure         bool
	XtremIODeviceMapper     bool
	XtremIOMultipath        bool
	XtremIORemoteManagement bool
}

type plainTextConfig struct {
	AwsSecretKey      string
	RackspacePassword string
	OpenstackPassword string
	ScaleIoPassword   string
	XtremIoPassword   string
}

// Config contains the configuration information
type Config struct {
	secureConfig
	plainTextConfig

	GlobalFlags     *flag.FlagSet `json:"-"`
	AdditionalFlags *flag.FlagSet `json:"-"`
	Viper           *viper.Viper  `json:"-"`
	Host            string        `json:"-"`

	jsonMarshalStrategy JSONMarshalStrategy
}

// New initializes a new instance of a Config struct
func New() *Config {
	return NewConfig(true, true, "config", "yml")
}

// NewConfig initialies a new instance of a Config object with the specified
// options.
func NewConfig(
	loadGlobalConfig, loadUserConfig bool,
	configName, configType string) *Config {

	log.Debug("initializing configuration")

	c := &Config{
		secureConfig:        secureConfig{},
		plainTextConfig:     plainTextConfig{},
		Viper:               viper.New(),
		GlobalFlags:         &flag.FlagSet{},
		AdditionalFlags:     &flag.FlagSet{},
		jsonMarshalStrategy: JSONMarshalSecure,
	}
	c.Viper.SetTypeByDefaultValue(true)
	c.Viper.SetConfigName(configName)
	c.Viper.SetConfigType(configType)

	cfgFile := fmt.Sprintf("%s.%s", configName, configType)
	etcRexRayFile := util.EtcFilePath(cfgFile)
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

// JSONMarshalStrategy gets the JSON marshalling strategy
func (c *Config) JSONMarshalStrategy() JSONMarshalStrategy {
	return c.jsonMarshalStrategy
}

// SetJSONMarshalStrategy sets the JSON marshalling strategy
func (c *Config) SetJSONMarshalStrategy(s JSONMarshalStrategy) {
	c.jsonMarshalStrategy = s
}

// Copy creates a copy of this Config instance
func (c *Config) Copy() (*Config, error) {
	newC := New()
	c.Viper.Unmarshal(&newC.plainTextConfig)
	c.Viper.Unmarshal(&newC.secureConfig)
	return newC, nil
}

// FromJSON initializes a new Config instance from a JSON string
func FromJSON(from string) (*Config, error) {
	c := New()
	if err := json.Unmarshal([]byte(from), c); err != nil {
		return nil, err
	}
	c.sync()
	return c, nil
}

// ToJSON exports this Config instance to a JSON string
func (c *Config) ToJSON() (string, error) {
	buf, _ := c.marshalJSON(JSONMarshalPlainText)
	return string(buf), nil
}

// ToSecureJSON exports this Config instance to a JSON string omitting any of
// the secure fields
func (c *Config) ToSecureJSON() (string, error) {
	buf, _ := c.marshalJSON(JSONMarshalSecure)
	return string(buf), nil
}

// MarshalJSON implements the encoding/json.Marshaller interface. It allows
// this type to provide its own marshalling routine.
func (c *Config) MarshalJSON() ([]byte, error) {
	return c.marshalJSON(c.jsonMarshalStrategy)
}

func (c *Config) marshalJSON(s JSONMarshalStrategy) ([]byte, error) {
	switch s {
	case JSONMarshalPlainText:
		s := struct {
			plainTextConfig
			secureConfig
		}{
			c.plainTextConfig,
			c.secureConfig,
		}
		return json.MarshalIndent(s, "", "  ")
	default:
		return json.MarshalIndent(c.secureConfig, "", "  ")
	}
}

// ReadConfig reads a configuration stream into the current config instance
func (c *Config) ReadConfig(in io.Reader) error {

	if in == nil {
		return errors.New("config reader is nil")
	}

	c.Viper.ReadConfigNoNil(in)
	c.Viper.Unmarshal(&c.secureConfig)
	c.Viper.Unmarshal(&c.plainTextConfig)

	for key := range keys {
		c.updateFlag(key, c.GlobalFlags)
		c.updateFlag(key, c.AdditionalFlags)
	}

	return nil
}

// ReadConfigFile reads a configuration files into the current config instance
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

// EnvVars returns an array of the initialized configuration keys as key=value
// strings where the key is configuration key's environment variable key and
// the value is the current value for that key.
func (c *Config) EnvVars() []string {
	evArr := make([]string, len(keys))
	for k, v := range keys {
		evArr = append(evArr,
			fmt.Sprintf("%s=%s", v.EnvVar, c.Viper.GetString(k)))
	}
	return evArr
}

func (c *Config) sync() {

	w := c.Viper.Set

	w(Host, c.Host)
	w(LogLevel, c.LogLevel)
	w(StorageDrivers, c.StorageDrivers)
	w(VolumeDrivers, c.VolumeDrivers)
	w(OSDrivers, c.OSDrivers)
	w(MinVolSize, c.MinVolSize)
	w(RemoteManagement, c.RemoteManagement)

	w(DockerVolumeType, c.DockerVolumeType)
	w(DockerIOPS, c.DockerIOPS)
	w(DockerSize, c.DockerSize)
	w(DockerAvailabilityZone, c.DockerAvailabilityZone)
	w(AwsAccessKey, c.AwsAccessKey)
	w(AwsSecretKey, c.AwsSecretKey)
	w(AwsRegion, c.AwsRegion)

	w(RackspaceAuthURL, c.RackspaceAuthURL)
	w(RackspaceUserID, c.RackspaceUserID)
	w(RackspaceUserName, c.RackspaceUserName)
	w(RackspacePassword, c.RackspacePassword)
	w(RackspaceTenantID, c.RackspaceTenantID)
	w(RackspaceTenantName, c.RackspaceTenantName)
	w(RackspaceDomainID, c.RackspaceDomainID)
	w(RackspaceDomainName, c.RackspaceDomainName)

	w(OpenstackAuthURL, c.OpenstackAuthURL)
	w(OpenstackUserID, c.OpenstackUserID)
	w(OpenstackUserName, c.OpenstackUserName)
	w(OpenstackPassword, c.OpenstackPassword)
	w(OpenstackTenantID, c.OpenstackTenantID)
	w(OpenstackTenantName, c.OpenstackTenantName)
	w(OpenstackDomainID, c.OpenstackDomainID)
	w(OpenstackDomainName, c.OpenstackDomainName)
	w(OpenstackRegionName, c.OpenstackRegionName)
	w(OpenstackAvailabilityZoneName, c.OpenstackAvailabilityZoneName)

	w(ScaleIOEndpoint, c.ScaleIOEndpoint)
	w(ScaleIOInsecure, c.ScaleIOInsecure)
	w(ScaleIOUseCerts, c.ScaleIOUseCerts)
	w(ScaleIOUserName, c.ScaleIOUserName)
	w(ScaleIOPassword, c.ScaleIoPassword)
	w(ScaleIOSystemID, c.ScaleIOSystemID)
	w(ScaleIOSystemName, c.ScaleIOSystemName)
	w(ScaleIOProtectionDomainID, c.ScaleIOProtectionDomainID)
	w(ScaleIOProtectionDomainName, c.ScaleIOProtectionDomainName)
	w(ScaleIOStoragePoolID, c.ScaleIOStoragePoolID)
	w(ScaleIOStoragePoolName, c.ScaleIOStoragePoolName)

	w(XtremIOEndpoint, c.XtremIOEndpoint)
	w(XtremIOUserName, c.XtremIOUserName)
	w(XtremIOPassword, c.XtremIoPassword)
	w(XtremIOInsecure, c.XtremIOInsecure)
	w(XtremIODeviceMapper, c.XtremIODeviceMapper)
	w(XtremIOMultipath, c.XtremIOMultipath)
	w(XtremIORemoteManagement, c.XtremIORemoteManagement)
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
