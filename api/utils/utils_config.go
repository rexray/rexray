package utils

import (
	"fmt"
	"strings"

	gofig "github.com/akutz/gofig/types"
)

func isSet(
	config gofig.Config,
	key string,
	roots ...string) bool {

	return isSetPrefix(config, "libstorage.", key, roots...)
}

func isSetPrefix(
	config gofig.Config,
	prefix, key string,
	roots ...string) bool {

	for _, r := range roots {
		rk := strings.Replace(key, prefix, fmt.Sprintf("%s.", r), 1)
		if config.IsSet(rk) {
			return true
		}
	}

	if config.IsSet(key) {
		return true
	}

	return false
}

func getString(
	config gofig.Config,
	key string,
	roots ...string) string {

	return getStringPrefix(config, "libstorage.", key, roots...)
}

func getStringPrefix(
	config gofig.Config,
	prefix, key string,
	roots ...string) string {

	var val string

	for _, r := range roots {
		rk := strings.Replace(key, prefix, fmt.Sprintf("%s.", r), 1)
		if val = config.GetString(rk); val != "" {
			return val
		}
	}

	val = config.GetString(key)
	if val != "" {
		return val
	}

	return ""
}

func getBool(
	config gofig.Config,
	key string,
	roots ...string) bool {

	return getBoolPrefix(config, "libstorage.", key, roots...)
}

func getBoolPrefix(
	config gofig.Config,
	prefix, key string,
	roots ...string) bool {

	for _, r := range roots {
		rk := strings.Replace(key, prefix, fmt.Sprintf("%s.", r), 1)
		if config.IsSet(rk) {
			return config.GetBool(rk)
		}
	}

	if config.IsSet(key) {
		return config.GetBool(key)
	}

	return false
}

func getStringSlice(
	config gofig.Config,
	key string,
	roots ...string) []string {

	return getStringSlicePrefix(config, "libstorage.", key, roots...)
}

func getStringSlicePrefix(
	config gofig.Config,
	prefix, key string,
	roots ...string) []string {

	var val []string

	for _, r := range roots {
		rk := strings.Replace(key, prefix, fmt.Sprintf("%s.", r), 1)
		if val = config.GetStringSlice(rk); len(val) > 0 {
			return val
		}
	}

	return config.GetStringSlice(key)
}
