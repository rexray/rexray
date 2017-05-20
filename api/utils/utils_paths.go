package utils

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/akutz/gotil"
	"github.com/codedellemc/libstorage/api/types"
)

// NewPathConfig returns a new path configuration object.
func NewPathConfig(ctx types.Context, home, token string) *types.PathConfig {

	if token == "" {
		if v := os.Getenv("LIBSTORAGE_APPTOKEN"); v != "" {
			token = v
		} else {
			token = "libstorage"
		}
	}

	pathConfig := &types.PathConfig{
		Token:    token,
		UserHome: path.Join(gotil.HomeDir(), fmt.Sprintf(".%s", token)),
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
		envVarHomeLSX    = fmt.Sprintf("%s_HOME_LSX", ucTok)
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
	ctx.WithField("path", pathConfig.Home).Debug("mkdir -p")

	// init the other paths
	initPathConfigFieldWithEnvVar(ctx, envVarHomeEtc, &pathConfig.Etc)
	initPathConfigFieldWithEnvVar(ctx, envVarHomeEtcTLS, &pathConfig.TLS)
	initPathConfigFieldWithEnvVar(ctx, envVarHomeLib, &pathConfig.Lib)
	initPathConfigFieldWithEnvVar(ctx, envVarHomeLog, &pathConfig.Log)
	initPathConfigFieldWithEnvVar(ctx, envVarHomeRun, &pathConfig.Run)
	initPathConfigFieldWithEnvVar(ctx, envVarHomeLSX, &pathConfig.LSX)

	var (
		lsxNameBuf = &bytes.Buffer{}
		root       = pathConfig.Home == "/"
	)

	fmt.Fprint(lsxNameBuf, "lsx-")
	fmt.Fprint(lsxNameBuf, runtime.GOOS)
	if runtime.GOARCH != "amd64" {
		fmt.Fprint(lsxNameBuf, "-")
		fmt.Fprint(lsxNameBuf, runtime.GOARCH)
	}
	if runtime.GOOS == "windows" {
		fmt.Fprint(lsxNameBuf, ".exe")
	}
	lsx := lsxNameBuf.String()
	ctx.WithField("lsx", lsx).Debug("lsx binary name")

	initPathConfigFieldWithPath(
		ctx, root, true, token, pathConfig.Home, "etc", &pathConfig.Etc)
	initPathConfigFieldWithPath(
		ctx, false, true, token, pathConfig.Etc, "tls", &pathConfig.TLS)
	initPathConfigFieldWithPath(
		ctx, root, true, token, pathConfig.Home, "var/lib", &pathConfig.Lib)
	initPathConfigFieldWithPath(
		ctx, root, true, token, pathConfig.Home, "var/log", &pathConfig.Log)
	initPathConfigFieldWithPath(
		ctx, root, true, token, pathConfig.Home, "var/run", &pathConfig.Run)
	initPathConfigFieldWithPath(
		ctx, false, false, token, pathConfig.Lib, lsx, &pathConfig.LSX)
	initPathConfigFieldWithPath(
		ctx, false, false, token, pathConfig.TLS,
		fmt.Sprintf("%s.crt", token), &pathConfig.DefaultTLSCertFile)
	initPathConfigFieldWithPath(
		ctx, false, false, token, pathConfig.TLS,
		fmt.Sprintf("%s.key", token), &pathConfig.DefaultTLSKeyFile)
	initPathConfigFieldWithPath(
		ctx, false, false, token, pathConfig.TLS,
		"known_hosts", &pathConfig.DefaultTLSKnownHosts)
	initPathConfigFieldWithPath(
		ctx, false, false, token, pathConfig.TLS,
		"cacerts", &pathConfig.DefaultTLSTrustedRootsFile)

	return pathConfig
}

func initPathConfigFieldWithEnvVar(
	ctx types.Context, envVarName string, field *string) {
	if v := os.Getenv(envVarName); v != "" && gotil.FileExists(v) {
		*field = v
	}
}

func initPathConfigFieldWithPath(
	ctx types.Context, root, mkdir bool,
	token, parent, dir string, field *string) {
	defer func() {
		if !mkdir {
			return
		}
		os.MkdirAll(*field, 0755)
		ctx.WithField("path", *field).Debug("mkdir -p")
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
