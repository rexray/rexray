package utils

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// ParseAuthConfig returns a new AuthTokenConfig instance.
func ParseAuthConfig(
	ctx types.Context,
	config gofig.Config,
	fields log.Fields,
	roots ...string) (*types.AuthConfig, error) {

	const prefix = types.ConfigServer + "."

	if !isSetPrefix(config, prefix, types.ConfigServerAuthAllow, roots...) {
		ctx.Debug("server auth config not defined")
		return nil, nil
	}

	f := func(k string, v interface{}) {
		if fields == nil {
			return
		}
		fields[k] = v
		ctx.WithField(k, v).Debug("parsed server auth property")
	}

	authConfig := &types.AuthConfig{Alg: "HS256"}

	if isSetPrefix(config, prefix, types.ConfigServerAuthDisabled, roots...) {
		authConfig.Disabled = getBool(
			config, types.ConfigServerAuthDisabled, roots...)
		f(types.ConfigServerAuthDisabled, authConfig.Disabled)
	}

	if isSetPrefix(config, prefix, types.ConfigServerAuthKey, roots...) {
		szKey := getStringPrefix(
			config, prefix, types.ConfigServerAuthKey, roots...)
		f(types.ConfigServerAuthKey, szKey)
		if gotil.FileExists(szKey) {
			buf, err := ioutil.ReadFile(szKey)
			if err != nil {
				return nil, err
			}
			authConfig.Key = buf
		} else {
			authConfig.Key = []byte(szKey)
		}
	}

	if isSetPrefix(config, prefix, types.ConfigServerAuthAlg, roots...) {
		authConfig.Alg = getStringPrefix(
			config, prefix, types.ConfigServerAuthAlg, roots...)
	}
	f(types.ConfigServerAuthAlg, authConfig.Alg)

	if isSetPrefix(config, prefix, types.ConfigServerAuthAllow, roots...) {
		authConfig.Allow = getStringSlicePrefix(
			config, prefix, types.ConfigServerAuthAllow, roots...)
		f(types.ConfigServerAuthAllow, authConfig.Allow)
	}

	if isSetPrefix(config, prefix, types.ConfigServerAuthDeny, roots...) {
		authConfig.Deny = getStringSlicePrefix(
			config, prefix, types.ConfigServerAuthDeny, roots...)
		f(types.ConfigServerAuthDeny, authConfig.Deny)
	}

	return authConfig, nil
}
