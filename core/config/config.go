package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/emccode/rexray/core/errors"
	"github.com/emccode/rexray/util"
)

var (
	envVarRx      *regexp.Regexp
	bracketRx     *regexp.Regexp
	registrations []*Registration
)

func init() {
	envVarRx = regexp.MustCompile(`^\s*([^#=]+)=(.+)$`)
	bracketRx = regexp.MustCompile(`^\[(.*)\]$`)
	loadEtcEnvironment()
	Register(globalRegistration())
	Register(driverRegistration())
}

// Config contains the configuration information
type Config struct {
	FlagSets map[string]*flag.FlagSet `json:"-"`
	v        *viper.Viper
}

// Register registers a new configuration with the config package.
func Register(r *Registration) {
	registrations = append(registrations, r)
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
		v:        viper.New(),
		FlagSets: map[string]*flag.FlagSet{},
	}
	c.v.SetTypeByDefaultValue(false)
	c.v.SetConfigName(configName)
	c.v.SetConfigType(configType)

	c.processRegistrations()

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

	return c
}

func (c *Config) processRegistrations() {
	for _, r := range registrations {

		fs := &flag.FlagSet{}

		for _, k := range r.keys {

			evn := k.envVarName

			if !strings.Contains(k.keyName, ".") {
				evn = fmt.Sprintf("REXRAY_%s", k.envVarName)
			}

			log.WithFields(log.Fields{
				"keyName":      k.keyName,
				"keyType":      k.keyType,
				"flagName":     k.flagName,
				"envVar":       evn,
				"defaultValue": k.defVal,
				"usage":        k.desc,
			}).Debug("adding flag")

			// bind the environment variable
			c.v.BindEnv(k.keyName, evn)

			if k.short == "" {
				switch k.keyType {
				case String:
					fs.String(k.flagName, k.defVal.(string), k.desc)
				case Int:
					fs.Int(k.flagName, k.defVal.(int), k.desc)
				case Bool:
					fs.Bool(k.flagName, k.defVal.(bool), k.desc)
				}
			} else {
				switch k.keyType {
				case String:
					fs.StringP(k.flagName, k.short, k.defVal.(string), k.desc)
				case Int:
					fs.IntP(k.flagName, k.short, k.defVal.(int), k.desc)
				case Bool:
					fs.BoolP(k.flagName, k.short, k.defVal.(bool), k.desc)
				}
			}

			c.v.BindPFlag(k.keyName, fs.Lookup(k.flagName))
		}

		c.FlagSets[r.name+" Flags"] = fs

		// read the config
		if r.yaml != "" {
			c.ReadConfig(bytes.NewReader([]byte(r.yaml)))
		}
	}
}

// Copy creates a copy of this Config instance
func (c *Config) Copy() (*Config, error) {
	newC := New()
	m := map[string]interface{}{}
	c.v.Unmarshal(&m)
	for k, v := range m {
		newC.v.Set(k, v)
	}
	return newC, nil
}

// FromJSON initializes a new Config instance from a JSON string
func FromJSON(from string) (*Config, error) {
	c := New()
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(from), &m); err != nil {
		return nil, err
	}
	for k, v := range m {
		c.v.Set(k, v)
	}
	return c, nil
}

// ToJSON exports this Config instance to a JSON string
func (c *Config) ToJSON() (string, error) {
	buf, _ := json.MarshalIndent(c, "", "  ")
	return string(buf), nil
}

// ToJSONCompact exports this Config instance to a compact JSON string
func (c *Config) ToJSONCompact() (string, error) {
	buf, _ := json.Marshal(c)
	return string(buf), nil
}

// MarshalJSON implements the encoding/json.Marshaller interface. It allows
// this type to provide its own marshalling routine.
func (c *Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.allSettings())
}

// ReadConfig reads a configuration stream into the current config instance
func (c *Config) ReadConfig(in io.Reader) error {

	if in == nil {
		return errors.New("config reader is nil")
	}

	c.v.MergeConfig(in)

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

// EnvVars returns an array of the initialized configuration keys as key=value
// strings where the key is configuration key's environment variable key and
// the value is the current value for that key.
func (c *Config) EnvVars() []string {
	keyVals := c.allSettings()
	envVars := make(map[string]string)
	c.flattenEnvVars("", keyVals, envVars)
	var evArr []string
	for k, v := range envVars {
		kk := k
		if !strings.Contains(k, "_") {
			kk = "REXRAY_" + k
		}
		evArr = append(evArr, fmt.Sprintf("%s=%v", kk, v))
	}
	return evArr
}

// flattenEnvVars returns a map of configuration keys coming from a config
// which may have been nested.
func (c *Config) flattenEnvVars(
	prefix string, keys map[string]interface{}, envVars map[string]string) {

	for k, v := range keys {

		var kk string
		if prefix == "" {
			kk = k
		} else {
			kk = fmt.Sprintf("%s.%s", prefix, k)
		}
		ek := strings.ToUpper(strings.Replace(kk, ".", "_", -1))

		log.WithFields(log.Fields{
			"key":   kk,
			"value": v,
		}).Debug("flattening env vars")

		switch vt := v.(type) {
		case string:
			envVars[ek] = vt
		case []interface{}:
			var vArr []string
			for _, iv := range vt {
				vArr = append(vArr, iv.(string))
			}
			envVars[ek] = strings.Join(vArr, " ")
		case map[string]interface{}:
			c.flattenEnvVars(kk, vt, envVars)
		case bool:
			envVars[ek] = fmt.Sprintf("%v", vt)
		case int, int32, int64:
			envVars[ek] = fmt.Sprintf("%v", vt)
		}
	}
	return
}

func (c *Config) allSettings() map[string]interface{} {
	as := map[string]interface{}{}
	ms := map[string]map[string]interface{}{}

	for k, v := range c.v.AllSettings() {
		switch tv := v.(type) {
		case nil:
			continue
		case map[string]interface{}:
			ms[k] = tv
		default:
			as[k] = tv
		}
	}

	for msk, msv := range ms {
		flat := map[string]interface{}{}
		flattenMapKeys(msk, msv, flat)
		for fk, fv := range flat {
			if asv, ok := as[fk]; ok && asv == fv {
				delete(as, fk)
			}
		}
		as[msk] = msv
	}

	return as
}

func flattenMapKeys(
	prefix string, m map[string]interface{}, flat map[string]interface{}) {
	for k, v := range m {
		kk := fmt.Sprintf("%s.%s", prefix, k)
		switch vt := v.(type) {
		case map[string]interface{}:
			flattenMapKeys(kk, vt, flat)
		default:
			flat[strings.ToLower(kk)] = v
		}
	}
}

// GetString returns the value associated with the key as a string
func (c *Config) GetString(k string) string {
	return c.v.GetString(k)
}

// GetBool returns the value associated with the key as a bool
func (c *Config) GetBool(k string) bool {
	return c.v.GetBool(k)
}

// GetStringSlice returns the value associated with the key as a string slice
func (c *Config) GetStringSlice(k string) []string {
	return c.v.GetStringSlice(k)
}

// GetInt returns the value associated with the key as an int
func (c *Config) GetInt(k string) int {
	return c.v.GetInt(k)
}

// Get returns the value associated with the key
func (c *Config) Get(k string) interface{} {
	return c.v.Get(k)
}

// Set sets an override value
func (c *Config) Set(k string, v interface{}) {
	c.v.Set(k, v)
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

func globalRegistration() *Registration {
	r := NewRegistration("Global")
	r.Yaml(`host: tcp://:7979
logLevel: warn
`)
	r.Key(String, "h", "tcp://:7979", "The REX-Ray host", "host")
	r.Key(String,
		"l", "warn", "The log level (error, warn, info, debug)", "logLevel")
	return r
}

func driverRegistration() *Registration {
	r := NewRegistration("Driver")
	r.Yaml(`osDrivers:
- linux
volumeDrivers:
- docker
`)
	r.Key(String, "", "linux", "The OS drivers to consider", "osDrivers")
	r.Key(String, "", "", "The storage drivers to consider", "storageDrivers")
	r.Key(String,
		"", "docker", "The volume drivers to consider", "volumeDrivers")
	return r
}
