package utils

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/akutz/gotil"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

// NewPathConfig returns a new path configuration object.
//
// The args parameter can take zero to three arguments:
//
//   1. The first argument is treated as a custom root, data
//      directory path. This defaults root path is "/".
//
//   2. The second argument is treated as a custom application
//      token. This is used to brand libStorage's paths, files,
//      and environment variables. The default application
//      token is "libstorage".
//
//   3. The third argument is treated as a custom home directory
//      for the executing process's user. The default value
//      is the user's home directory.
func NewPathConfig(args ...string) *types.PathConfig {

	var (
		home     string
		token    string
		userHome string
	)

	if len(args) > 0 {
		home = args[0]
	}
	if len(args) > 1 {
		token = args[1]
	}
	if len(args) > 2 {
		userHome = args[2]
	}

	if token == "" {
		if v := os.Getenv("LIBSTORAGE_APPTOKEN"); v != "" {
			token = v
		} else {
			token = "libstorage"
		}
	}
	if userHome == "" {
		userHome = gotil.HomeDir()
	}

	pathConfig := &types.PathConfig{
		Token:    token,
		UserHome: path.Join(userHome, fmt.Sprintf(".%s", token)),
	}
	pathConfig.UserDefaultTLSKnownHosts = path.Join(
		pathConfig.UserHome, "known_hosts")

	var (
		ucTok            = strings.ToUpper(token)
		envVarHome       = fmt.Sprintf("%s_HOME", ucTok)
		envVarHomeEtc    = fmt.Sprintf("%s_HOME_ETC", ucTok)
		envVarHomeEtcTLS = fmt.Sprintf("%s_HOME_ETC_TLS", ucTok)
		envVarHomeLib    = fmt.Sprintf("%s_HOME_LIB", ucTok)
		envVarHomeLog    = fmt.Sprintf("%s_HOME_LOG", ucTok)
		envVarHomeRun    = fmt.Sprintf("%s_HOME_RUN", ucTok)
	)

	// init the home dir
	if home != "" {
		pathConfig.Home = home
	} else if pathConfig.Home = os.Getenv(envVarHome); pathConfig.Home == "" {
		pathConfig.Home = "/"
	}
	if pathConfig.Home == "/" && os.Geteuid() != 0 {
		pathConfig.Home = path.Join(gotil.HomeDir(), fmt.Sprintf(".%s", token))
	}
	os.MkdirAll(pathConfig.Home, 0755)

	// init the other paths
	initPathConfigFieldWithEnvVar(envVarHomeEtc, &pathConfig.Etc)
	initPathConfigFieldWithEnvVar(envVarHomeEtcTLS, &pathConfig.TLS)
	initPathConfigFieldWithEnvVar(envVarHomeLib, &pathConfig.Lib)
	initPathConfigFieldWithEnvVar(envVarHomeLog, &pathConfig.Log)
	initPathConfigFieldWithEnvVar(envVarHomeRun, &pathConfig.Run)

	root := pathConfig.Home == "/"

	initPathConfigFieldWithPath(
		root, true, token, pathConfig.Home, "etc", &pathConfig.Etc)
	initPathConfigFieldWithPath(
		false, true, token, pathConfig.Etc, "tls", &pathConfig.TLS)
	initPathConfigFieldWithPath(
		root, true, token, pathConfig.Home, "var/lib", &pathConfig.Lib)
	initPathConfigFieldWithPath(
		root, true, token, pathConfig.Home, "var/log", &pathConfig.Log)
	initPathConfigFieldWithPath(
		root, true, token, pathConfig.Home, "var/run", &pathConfig.Run)
	initPathConfigFieldWithPath(
		false, false, token, pathConfig.TLS,
		fmt.Sprintf("%s.crt", token), &pathConfig.DefaultTLSCertFile)
	initPathConfigFieldWithPath(
		false, false, token, pathConfig.TLS,
		fmt.Sprintf("%s.key", token), &pathConfig.DefaultTLSKeyFile)
	initPathConfigFieldWithPath(
		false, false, token, pathConfig.TLS,
		"known_hosts", &pathConfig.DefaultTLSKnownHosts)
	initPathConfigFieldWithPath(
		false, false, token, pathConfig.TLS,
		"cacerts", &pathConfig.DefaultTLSTrustedRootsFile)

	/*ctx.WithFields(map[string]interface{}{
		"token":   pathConfig.Token,
		"usrHome": pathConfig.UserHome,
		"sysHome": pathConfig.Home,
		"sysEtc":  pathConfig.Etc,
		"sysTLS":  pathConfig.TLS,
		"sysLib":  pathConfig.Lib,
		"sysLog":  pathConfig.Log,
		"sysRun":  pathConfig.Run,
	}).Info("created new path config")*/

	return pathConfig
}

func initPathConfigFieldWithEnvVar(envVarName string, field *string) {
	if v := os.Getenv(envVarName); v != "" && gotil.FileExists(v) {
		*field = v
	}
}

func initPathConfigFieldWithPath(
	root, mkdir bool,
	token, parent, dir string, field *string) {
	defer func() {
		if !mkdir {
			return
		}
		os.MkdirAll(*field, 0755)
	}()

	if *field != "" {
		return
	}
	if root {
		*field = path.Join(parent, dir, token)
	} else {
		*field = path.Join(parent, dir)
	}
}
