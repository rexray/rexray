package types

const (
	//ConfigIntegration is a config key.
	ConfigIntegration = ConfigRoot + ".integration"

	// ConfigIntegrationVolMountPreempt is a config key.
	ConfigIntegrationVolMountPreempt = ConfigIntegration + ".volume.mount.preempt"

	// ConfigIntegrationVolCreateDisable is a config key.
	ConfigIntegrationVolCreateDisable = ConfigIntegration + ".volume.create.disable"

	// ConfigIntegrationVolRemoveDisable is a config key.
	ConfigIntegrationVolRemoveDisable = ConfigIntegration + ".volume.remove.disable"

	// ConfigIntegrationVolUnmountIgnoreUsed is a config key.
	ConfigIntegrationVolUnmountIgnoreUsed = ConfigIntegration + ".volume.unmount.ignoreusedcount"

	// ConfigIntegrationVolPathCache is a config key.
	ConfigIntegrationVolPathCache = ConfigIntegration + ".volume.path.cache"

	//ConfigDockerVolumeCreateImplicit is a config key.
	ConfigIntegrationVolCreateImplicit = ConfigIntegration + ".volume.create.implicit"
)
