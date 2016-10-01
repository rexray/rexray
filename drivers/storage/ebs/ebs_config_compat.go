package ebs

import (
	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
)

const (
	//ConfigEBS is a config key.
	ConfigEBS = "ebs"

	//ConfigEBSAccessKey is a config key.
	ConfigEBSAccessKey = ConfigEBS + "." + AccessKey

	//ConfigEBSSecretKey is a config key.
	ConfigEBSSecretKey = ConfigEBS + "." + SecretKey

	//ConfigEBSRegion is a config key.
	ConfigEBSRegion = ConfigEBS + "." + Region

	//ConfigEBSEndpoint is a config key.
	ConfigEBSEndpoint = ConfigEBS + "." + Endpoint

	//ConfigEBSMaxRetries is a config key.
	ConfigEBSMaxRetries = ConfigEBS + "." + MaxRetries

	//ConfigEBSTag is a config key.
	ConfigEBSTag = ConfigEBS + "." + Tag

	//ConfigEBSRexrayTag is a config key.
	ConfigEBSRexrayTag = ConfigEBS + ".rexrayTag"

	//ConfigOldEBS is a config key.
	ConfigOldEBS = "ec2"

	//ConfigOldEBSAccessKey is a config key.
	ConfigOldEBSAccessKey = ConfigOldEBS + "." + AccessKey

	//ConfigOldEBSSecretKey is a config key.
	ConfigOldEBSSecretKey = ConfigOldEBS + "." + SecretKey

	//ConfigOldEBSRegion is a config key.
	ConfigOldEBSRegion = ConfigOldEBS + "." + Region

	//ConfigOldEBSEndpoint is a config key.
	ConfigOldEBSEndpoint = ConfigOldEBS + "." + Endpoint

	//ConfigOldEBSMaxRetries is a config key.
	ConfigOldEBSMaxRetries = ConfigOldEBS + "." + MaxRetries

	//ConfigOldEBSTag is a config key.
	ConfigOldEBSTag = ConfigOldEBS + "." + Tag

	//ConfigOldEBSRexrayTag is a config key.
	ConfigOldEBSRexrayTag = ConfigOldEBS + ".rexrayTag"
)

// BackCompat ensures keys can be used from old configurations.
func BackCompat(config gofig.Config) {
	checks := [][]string{
		{ConfigEBSAccessKey, ConfigOldEBSAccessKey},
		{ConfigEBSSecretKey, ConfigOldEBSSecretKey},
		{ConfigEBSRegion, ConfigOldEBSRegion},
		{ConfigEBSEndpoint, ConfigOldEBSEndpoint},
		{ConfigEBSMaxRetries, ConfigOldEBSMaxRetries},
		{ConfigEBSTag, ConfigOldEBSTag},
		{ConfigEBSRexrayTag, ConfigOldEBSRexrayTag},
	}
	for _, check := range checks {
		if !config.IsSet(check[0]) && config.IsSet(check[1]) {
			log.Debug(config.Get(check[1]))
			config.Set(check[0], config.Get(check[1]))
		}
	}
}
