package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
)

var knownHostRX = regexp.MustCompile(`(?i)^([^\s]+?)\s([^\s]+?)\s(.+)$`)

// ParseKnownHost parses a known host line that's in the expected format:
// "host algorithm fingerprint".
func ParseKnownHost(
	ctx types.Context,
	text string) (*types.TLSKnownHost, error) {

	m := knownHostRX.FindStringSubmatch(text)
	if len(m) == 0 {
		return nil, nil
	}

	ctx.WithFields(map[string]interface{}{
		"khHost": m[1],
		"khAlg":  m[2],
		"khSig":  m[3],
	}).Debug("parsing known_hosts file fields")

	buf, err := hex.DecodeString(strings.Replace(m[3], ":", "", -1))
	if err != nil {
		return nil, goof.WithError(
			"error decoding known host fingerprint", err)
	}
	return &types.TLSKnownHost{
		Host:        m[1],
		Alg:         m[2],
		Fingerprint: buf,
	}, nil
}

// ParseTLSConfig returns a new TLS configuration.
func ParseTLSConfig(
	ctx types.Context,
	config gofig.Config,
	proto string,
	fields log.Fields,
	roots ...string) (tlsConfig *types.TLSConfig, tlsErr error) {

	if strings.EqualFold(proto, "unix") {
		enable, _ := strconv.ParseBool(
			os.Getenv("LIBSTORAGE_TLS_SOCKITTOME"))
		if !enable {
			ctx.Debug("disabling tls for unix sockets")
			return nil, nil
		}
	}

	ctx.Debug("parsing tls config")

	pathConfig := context.MustPathConfig(ctx)

	f := func(k string, v interface{}) {
		if fields == nil {
			return
		}
		fields[k] = v
		ctx.WithField(k, v).Debug("tls field set")
	}

	newTLS := func(k string, v interface{}) {
		if tlsConfig != nil {
			return
		}
		ctx.WithField(k, v).Info("tls enabled")
		tlsConfig = &types.TLSConfig{Config: tls.Config{}}
	}

	// check to see if TLS is disabled
	if ok := getBool(config, types.ConfigTLSDisabled, roots...); ok {
		f(types.ConfigTLSDisabled, true)
		ctx.WithField(types.ConfigTLSDisabled, false).Info("tls disabled")
		return nil, nil
	}

	// check to see if TLS is enabled with a simple truthy value
	if ok := getBool(config, types.ConfigTLS, roots...); ok {
		f(types.ConfigTLS, true)
		newTLS(types.ConfigTLS, true)

	} else if v := getString(config, types.ConfigTLS, roots...); v != "" {
		// check to see if TLS is disabled
		if strings.EqualFold(v, "false") {
			f(types.ConfigTLS, "false")
			ctx.WithField(types.ConfigTLS, "false").Info("tls disabled")
			return nil, nil
		}

		// check to see if TLS is enabled with insecure
		if strings.EqualFold(v, "insecure") {
			f(types.ConfigTLS, "insecure")
			newTLS(types.ConfigTLS, "insecure")
			tlsConfig.InsecureSkipVerify = true

			// check to see if TLS is enabled with peers
		} else if strings.EqualFold(v, "verifyPeers") {
			f(types.ConfigTLS, "verifyPeers")
			newTLS(types.ConfigTLS, "verifyPeers")
			tlsConfig.InsecureSkipVerify = true
			tlsConfig.VerifyPeers = true

			// check to see if TLS is enabled with an expected sha256
			// fingerprint
		} else if kh, err := ParseKnownHost(ctx, v); err != nil {
			ctx.Error(err)
			return nil, err
		} else if kh != nil {
			f(types.ConfigTLS, kh.String())
			newTLS(types.ConfigTLS, kh.String())
			tlsConfig.InsecureSkipVerify = true
			tlsConfig.VerifyPeers = true
			tlsConfig.KnownHost = kh
		}
	}

	// always check for the user's known_hosts file
	if gotil.FileExists(pathConfig.UserDefaultTLSKnownHosts) {
		newTLS("usrKnownHosts", pathConfig.UserDefaultTLSKnownHosts)
		f("usrKnownHosts", pathConfig.UserDefaultTLSKnownHosts)
		tlsConfig.UsrKnownHosts = pathConfig.UserDefaultTLSKnownHosts
		tlsConfig.InsecureSkipVerify = true
		tlsConfig.VerifyPeers = true
	}

	// always check for the system's known_hosts file
	if err := func() error {
		if !isSet(config, types.ConfigTLSKnownHosts, roots...) {
			return nil
		}
		khFile := getString(config, types.ConfigTLSKnownHosts, roots...)

		// is the known_hosts file the same as the default known_hosts
		// file? It's not possible to use os.SameFile as the files may not
		// yet exist
		isDefKH := strings.EqualFold(khFile, pathConfig.DefaultTLSKnownHosts)

		if !gotil.FileExists(khFile) {
			if !isDefKH {
				return goof.WithFields(map[string]interface{}{
					"path":        khFile,
					"defaultPath": pathConfig.DefaultTLSKnownHosts,
				}, "invalid known_hosts file")
			}
			return nil
		}

		newTLS(types.ConfigTLSKnownHosts, khFile)
		f(types.ConfigTLSKnownHosts, khFile)
		tlsConfig.SysKnownHosts = khFile
		tlsConfig.VerifyPeers = true
		tlsConfig.InsecureSkipVerify = true

		return nil
	}(); err != nil {
		return nil, err
	}

	// always check for the cacerts file
	if err := func() error {
		if !isSet(config, types.ConfigTLSTrustedCertsFile, roots...) {
			return nil
		}

		caCerts := getString(config, types.ConfigTLSTrustedCertsFile, roots...)

		// is the key file the same as the default cacerts file? It's not
		// possible to use os.SameFile as the files may not yet exist
		isDefCA := strings.EqualFold(
			caCerts, pathConfig.DefaultTLSTrustedRootsFile)

		if !gotil.FileExists(caCerts) {
			if !isDefCA {
				return goof.WithField("path", caCerts, "invalid cacerts file")
			}
			return nil
		}

		buf, err := func() ([]byte, error) {
			f, err := os.Open(caCerts)
			if err != nil {
				return nil, goof.WithFieldE(
					"path", caCerts, "error opening cacerts file", err)
			}
			defer f.Close()
			buf, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, goof.WithFieldE(
					"path", caCerts, "error reading cacerts file", err)
			}
			return buf, nil
		}()
		if err != nil {
			return err
		}

		newTLS(types.ConfigTLSTrustedCertsFile, caCerts)
		f(types.ConfigTLSTrustedCertsFile, caCerts)

		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(buf)
		tlsConfig.RootCAs = certPool
		tlsConfig.ClientCAs = certPool

		return nil
	}(); err != nil {
		return nil, err
	}

	// always check for the cert and key files
	if err := func() error {
		if !isSet(config, types.ConfigTLSKeyFile, roots...) {
			return nil
		}
		keyFile := getString(config, types.ConfigTLSKeyFile, roots...)

		// is the key file the same as the default key file? It's not
		// possible to use os.SameFile as the files may not yet exist
		isDefKF := strings.EqualFold(keyFile, pathConfig.DefaultTLSKeyFile)

		if !gotil.FileExists(keyFile) {
			if !isDefKF {
				return goof.WithField("path", keyFile, "invalid key file")
			}
			return nil
		}

		crtFile := getString(config, types.ConfigTLSCertFile, roots...)

		// is the cert file the same as the default cert file? It's not
		// possible to use os.SameFile as the files may not yet exist
		isDefCF := strings.EqualFold(crtFile, pathConfig.DefaultTLSCertFile)

		if !gotil.FileExists(crtFile) {
			if !isDefCF {
				return goof.WithField("path", crtFile, "invalid crt file")
			}
			return nil
		}

		newTLS(types.ConfigTLSKeyFile, keyFile)
		f(types.ConfigTLSKeyFile, keyFile)
		f(types.ConfigTLSCertFile, crtFile)

		cer, err := tls.LoadX509KeyPair(crtFile, keyFile)
		if err != nil {
			return goof.WithError("error loading x509 pair", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cer}
		return nil
	}(); err != nil {
		return nil, err
	}

	if v := getString(
		config,
		types.ConfigTLSInsecure, roots...); v != "" {

		bv, _ := strconv.ParseBool(v)
		newTLS(types.ConfigTLSInsecure, bv)
		f(types.ConfigTLSInsecure, bv)
		tlsConfig.InsecureSkipVerify = bv
	}

	if v := getString(
		config,
		types.ConfigTLSVerifyPeers, roots...); v != "" {

		bv, _ := strconv.ParseBool(v)
		newTLS(types.ConfigTLSVerifyPeers, bv)
		f(types.ConfigTLSVerifyPeers, bv)
		tlsConfig.VerifyPeers = bv
		tlsConfig.InsecureSkipVerify = bv
	}

	if v := getString(
		config,
		types.ConfigTLSClientCertRequired, roots...); v != "" {

		bv, _ := strconv.ParseBool(v)
		newTLS(types.ConfigTLSClientCertRequired, bv)
		f(types.ConfigTLSClientCertRequired, bv)
		if bv {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}

	if v := getString(
		config,
		types.ConfigTLSServerName, roots...); v != "" {

		newTLS(types.ConfigTLSServerName, v)
		f(types.ConfigTLSServerName, v)
		tlsConfig.ServerName = v
	}

	return tlsConfig, nil
}
