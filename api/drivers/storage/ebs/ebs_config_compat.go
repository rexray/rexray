package types

import "github.com/akutz/gofig"
import log "github.com/Sirupsen/logrus"

const (
	//ConfigEBS is a config key.
	ConfigEBS = "ebs"

	//ConfigEBSAccessKey is a config key.
	ConfigEBSAccessKey = ConfigEBS + ".accessKey"

	//ConfigEBSSecretKey is a config key.
	ConfigEBSSecretKey = ConfigEBS + ".secretKey"

	//ConfigEBSRegion is a config key.
	ConfigEBSRegion = ConfigEBS + ".region"

	//ConfigEBSEndpoint is a config key.
	ConfigEBSEndpoint = ConfigEBS + ".endpoint"

	//ConfigEBSMaxRetries is a config key.
	ConfigEBSMaxRetries = ConfigEBS + ".maxRetries"

	//ConfigEBSTag is a config key.
	ConfigEBSTag = ConfigEBS + ".tag"

	//ConfigEBSRexrayTag is a config key.
	ConfigEBSRexrayTag = ConfigEBS + ".rexrayTag"

	//ConfigOldEBS is a config key.
	ConfigOldEBS = "ec2"

	//ConfigOldEBSAccessKey is a config key.
	ConfigOldEBSAccessKey = ConfigOldEBS + ".accessKey"

	//ConfigOldEBSSecretKey is a config key.
	ConfigOldEBSSecretKey = ConfigOldEBS + ".secretKey"

	//ConfigOldEBSRegion is a config key.
	ConfigOldEBSRegion = ConfigOldEBS + ".region"

	//ConfigOldEBSEndpoint is a config key.
	ConfigOldEBSEndpoint = ConfigOldEBS + ".endpoint"

	//ConfigOldEBSMaxRetries is a config key.
	ConfigOldEBSMaxRetries = ConfigOldEBS + ".maxRetries"

	//ConfigOldEBSTag is a config key.
	ConfigOldEBSTag = ConfigOldEBS + ".tag"

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
