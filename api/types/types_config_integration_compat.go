package types

import "github.com/akutz/gofig"
import log "github.com/Sirupsen/logrus"

const (
	//ConfigOldRoot is a config key.
	ConfigOldRoot = "volume"

	// ConfigOldIntegrationVolMountPreempt is a config key.
	ConfigOldIntegrationVolMountPreempt = ConfigOldRoot + ".mount.preempt"

	// ConfigOldIntegrationVolCreateDisable is a config key.
	ConfigOldIntegrationVolCreateDisable = ConfigOldRoot + ".create.disable"

	// ConfigOldIntegrationVolRemoveDisable is a config key.
	ConfigOldIntegrationVolRemoveDisable = ConfigOldRoot + ".remove.disable"

	// ConfigOldIntegrationVolUnmountIgnoreUsed is a config key.
	ConfigOldIntegrationVolUnmountIgnoreUsed = ConfigOldRoot + ".unmount.ignoreusedcount"

	// ConfigOldIntegrationVolPathCache is a config key.
	ConfigOldIntegrationVolPathCache = ConfigOldRoot + ".path.cache"
)

// BackCompat ensures keys can be used from old configurations.
func BackCompat(config gofig.Config) {
	checks := [][]string{
		{ConfigIntegrationVolMountPreempt, ConfigOldIntegrationVolMountPreempt},
		{ConfigIntegrationVolCreateDisable, ConfigOldIntegrationVolCreateDisable},
		{ConfigIntegrationVolRemoveDisable, ConfigOldIntegrationVolRemoveDisable},
		{ConfigIntegrationVolUnmountIgnoreUsed, ConfigOldIntegrationVolUnmountIgnoreUsed},
		{ConfigIntegrationVolPathCache, ConfigOldIntegrationVolPathCache},
	}
	for _, check := range checks {
		if !config.IsSet(check[0]) && config.IsSet(check[1]) {
			log.Debug(config.Get(check[1]))
			config.Set(check[0], config.Get(check[1]))
		}
	}
}
