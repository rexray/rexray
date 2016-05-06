package config

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/api/utils/paths"
)

// NewConfig returns a new configuration instance.
func NewConfig() (gofig.Config, error) {
	config := gofig.New()

	etcYML := paths.Etc.Join("config.yml")
	etcYAML := paths.Etc.Join("config.yaml")

	userHomeDir := gotil.HomeDir()
	usrYML := paths.Join(userHomeDir, "config.yml")
	usrYAML := paths.Join(userHomeDir, "config.yaml")

	if err := readConfigFile(config, etcYML); err != nil {
		return nil, err
	}
	if err := readConfigFile(config, etcYAML); err != nil {
		return nil, err
	}
	if err := readConfigFile(config, usrYML); err != nil {
		return nil, err
	}
	if err := readConfigFile(config, usrYAML); err != nil {
		return nil, err
	}

	return config, nil
}

// UpdateLogLevel updates the log level based on the config.
func UpdateLogLevel(config gofig.Config) {
	ll, err := log.ParseLevel(config.GetString(types.ConfigLogLevel))
	if err != nil {
		return
	}
	log.SetLevel(ll)
}

func readConfigFile(config gofig.Config, path string) error {
	if !gotil.FileExists(path) {
		return nil
	}
	return config.ReadConfigFile(path)
}

// DeviceAttachTimeout gets the configured device attach timeout.
func DeviceAttachTimeout(config gofig.Config) time.Duration {
	return utils.DeviceAttachTimeout(
		config.GetString(types.ConfigDeviceAttachTimeout))
}

// DeviceScanType gets the configured device scan type.
func DeviceScanType(config gofig.Config) types.DeviceScanType {
	return types.ParseDeviceScanType(config.GetInt(types.ConfigDeviceScanType))
}
