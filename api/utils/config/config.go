package config

import (
	"fmt"

	"github.com/akutz/gofig"
	"github.com/akutz/gotil"
	"github.com/emccode/libstorage/api/utils/paths"
)

// NewConfig returns a new configuration instance.
func NewConfig() (gofig.Config, error) {
	config := gofig.New()

	etcYML := fmt.Sprintf("%s/config.yml", paths.EtcDirPath())
	usrYML := fmt.Sprintf("%s/config.yml", paths.UsrDirPath)

	etcYAML := fmt.Sprintf("%s/config.yaml", paths.EtcDirPath())
	usrYAML := fmt.Sprintf("%s/config.yaml", paths.UsrDirPath)

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

func readConfigFile(config gofig.Config, path string) error {
	if !gotil.FileExists(path) {
		return nil
	}
	return config.ReadConfigFile(path)
}
