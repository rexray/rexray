package gofig

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/akutz/gofig/types"
)

// config contains the configuration information
type config struct {
	v                         *viper.Viper
	flagSets                  map[string]*pflag.FlagSet
	disableEnvVarSubstitution bool
}

func newConfigObj() *config {
	return &config{
		v:                         viper.New(),
		flagSets:                  map[string]*pflag.FlagSet{},
		disableEnvVarSubstitution: DisableEnvVarSubstitution,
	}
}

func (c *config) FlagSets() map[string]*pflag.FlagSet {
	return c.flagSets
}

func (c *config) processRegKeys(r types.ConfigRegistration) {
	fsn := fmt.Sprintf("%s Flags", r.Name())
	fs, ok := c.flagSets[fsn]
	if !ok {
		fs = &pflag.FlagSet{}
		c.flagSets[fsn] = fs
	}

	for k := range r.Keys() {

		if fs.Lookup(k.FlagName()) != nil {
			continue
		}

		evn := k.EnvVarName()

		if LogRegKey {
			log.WithFields(log.Fields{
				"keyName":      k.KeyName(),
				"keyType":      k.KeyType(),
				"flagName":     k.FlagName(),
				"envVar":       evn,
				"defaultValue": k.DefaultValue(),
				"usage":        k.Description(),
			}).Debug("adding flag")
		}

		// bind the environment variable
		c.v.BindEnv(k.KeyName(), evn)

		if k.Short() == "" {
			switch k.KeyType() {
			case types.String, types.SecureString:
				fs.String(k.FlagName(), k.DefaultValue().(string), k.Description())
			case types.Int:
				fs.Int(k.FlagName(), k.DefaultValue().(int), k.Description())
			case types.Bool:
				fs.Bool(k.FlagName(), k.DefaultValue().(bool), k.Description())
			}
		} else {
			switch k.KeyType() {
			case types.String, types.SecureString:
				fs.StringP(k.FlagName(), k.Short(), k.DefaultValue().(string), k.Description())
			case types.Int:
				fs.IntP(k.FlagName(), k.Short(), k.DefaultValue().(int), k.Description())
			case types.Bool:
				fs.BoolP(k.FlagName(), k.Short(), k.DefaultValue().(bool), k.Description())
			}
		}

		c.v.BindPFlag(k.KeyName(), fs.Lookup(k.FlagName()))
	}
}
