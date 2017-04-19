package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/codedellemc/libstorage/api/types"
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
	fields log.Fields,
	roots ...string) (tlsConfig *types.TLSConfig, tlsErr error) {

	ctx.Debug("parsing tls config")

	f := func(k string, v interface{}) {
		if fields == nil {
			return
		}
		fields[k] = v
		ctx.WithField(k, v).Debug("tls field set")
	}

	// defer the parsing of the cert, key, and cacerts files so that no
	// matter how tls is configured these files might be loaded. This behavior
	// is to accomodate the fact that the files can be placed in default
	// locations and thus there is no reason not to use them if they are
	// placed in their known locations
	defer func() {
		if tlsConfig == nil {
			return
		}

		defer func() {
			if tlsErr != nil {
				tlsConfig = nil
				ctx.Error(tlsErr)
			}
		}()

		// always check for the user's known_hosts file
		func() {
			khFile := path.Join(gotil.HomeDir(), ".libstorage", "known_hosts")
			if gotil.FileExists(khFile) {
				tlsConfig.UsrKnownHosts = khFile
				tlsConfig.VerifyPeers = true
			}
		}()

		// always check for the system's known_hosts file
		if tlsErr = func() error {
			if !isSet(config, types.ConfigTLSKnownHosts, roots...) {
				return nil
			}
			khFile := getString(config, types.ConfigTLSKnownHosts, roots...)

			// is the known_hosts file the same as the default known_hosts
			// file? It's not possible to use os.SameFile as the files may not
			// yet exist
			isDefKH := strings.EqualFold(
				khFile, types.DefaultTLSKnownHosts.Path())

			if !gotil.FileExists(khFile) {
				if !isDefKH {
					return goof.WithField(
						"path", khFile, "invalid known_hosts file")
				}
				return nil
			}

			f(types.ConfigTLSKnownHosts, khFile)

			tlsConfig.SysKnownHosts = khFile
			tlsConfig.VerifyPeers = true

			return nil
		}(); tlsErr != nil {
			return
		}

		// always check for the cacerts file
		if tlsErr = func() error {
			if !isSet(config, types.ConfigTLSTrustedCertsFile, roots...) {
				return nil
			}

			caCerts := getString(
				config, types.ConfigTLSTrustedCertsFile, roots...)

			// is the key file the same as the default cacerts file? It's not
			// possible to use os.SameFile as the files may not yet exist
			isDefCA := strings.EqualFold(
				caCerts, types.DefaultTLSTrustedRootsFile.Path())

			if !gotil.FileExists(caCerts) {
				if !isDefCA {
					return goof.WithField(
						"path", caCerts, "invalid cacerts file")
				}
				return nil
			}

			f(types.ConfigTLSTrustedCertsFile, caCerts)

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

			certPool := x509.NewCertPool()
			certPool.AppendCertsFromPEM(buf)
			tlsConfig.RootCAs = certPool
			tlsConfig.ClientCAs = certPool

			return nil
		}(); tlsErr != nil {
			return
		}

		// always check for the cert and key files
		tlsErr = func() error {
			if !isSet(config, types.ConfigTLSKeyFile, roots...) {
				return nil
			}
			keyFile := getString(config, types.ConfigTLSKeyFile, roots...)

			// is the key file the same as the default key file? It's not
			// possible to use os.SameFile as the files may not yet exist
			isDefKF := strings.EqualFold(
				keyFile, types.DefaultTLSKeyFile.Path())

			if !gotil.FileExists(keyFile) {
				if !isDefKF {
					return goof.WithField(
						"path", keyFile, "invalid key file")
				}
				return nil
			}

			f(types.ConfigTLSKeyFile, keyFile)

			crtFile := getString(config, types.ConfigTLSCertFile, roots...)

			// is the key file the same as the default cert file? It's not
			// possible to use os.SameFile as the files may not yet exist
			isDefCF := strings.EqualFold(
				crtFile, types.DefaultTLSCertFile.Path())

			if !gotil.FileExists(crtFile) {
				if !isDefCF {
					return goof.WithField(
						"path", crtFile, "invalid crt file")
				}
				return nil
			}

			f(types.ConfigTLSCertFile, crtFile)
			cer, err := tls.LoadX509KeyPair(crtFile, keyFile)
			if err != nil {
				return goof.WithError("error loading x509 pair", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cer}
			return nil
		}()
	}()

	if !isSet(config, types.ConfigTLS, roots...) {
		ctx.Info("tls not configured")
		return nil, nil
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
		ctx.WithField(types.ConfigTLS, true).Info("tls enabled")
		return &types.TLSConfig{Config: tls.Config{}}, nil
	}

	if v := getString(config, types.ConfigTLS, roots...); v != "" {
		// check to see if TLS is enabled with insecure
		if strings.EqualFold(v, "insecure") {
			f(types.ConfigTLS, "insecure")
			ctx.WithField(types.ConfigTLS, "insecure").Info("tls enabled")
			return &types.TLSConfig{
				Config: tls.Config{InsecureSkipVerify: true},
			}, nil
		}

		// check to see if TLS is enabled with peers
		if strings.EqualFold(v, "verifyPeers") {
			f(types.ConfigTLS, "verifyPeers")
			ctx.WithField(types.ConfigTLS, "verifyPeers").Info("tls enabled")
			return &types.TLSConfig{
				Config:      tls.Config{InsecureSkipVerify: true},
				VerifyPeers: true,
			}, nil
		}

		// check to see if TLS is enabled with an expected sha256 fingerprint
		kh, err := ParseKnownHost(ctx, v)
		if err != nil {
			ctx.Error(err)
			return nil, err
		}
		if kh != nil {
			ctx.WithField(types.ConfigTLS, v).Info("tls enabled")
			return &types.TLSConfig{
				Config:      tls.Config{InsecureSkipVerify: true},
				VerifyPeers: true,
				KnownHost:   kh,
			}, nil
		}
	}

	// tls is enabled; figure out its configuration
	tlsConfig = &types.TLSConfig{Config: tls.Config{}}

	if getBool(config, types.ConfigTLSInsecure, roots...) {
		tlsConfig.InsecureSkipVerify = true
		f(types.ConfigTLSInsecure, true)
	}

	if getBool(config, types.ConfigTLSVerifyPeers, roots...) {
		tlsConfig.VerifyPeers = true
		tlsConfig.InsecureSkipVerify = true
		f(types.ConfigTLSVerifyPeers, true)
	}

	if getBool(
		config, types.ConfigTLSClientCertRequired, roots...) {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		f(types.ConfigTLSClientCertRequired, true)
	}

	if v := getString(config, types.ConfigTLSServerName, roots...); v != "" {
		tlsConfig.ServerName = v
		f(types.ConfigTLSServerName, v)
	}

	return tlsConfig, nil
}
