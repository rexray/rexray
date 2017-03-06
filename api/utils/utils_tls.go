package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/codedellemc/libstorage/api/types"
)

// ParseTLSConfig returns a new TLS configuration.
func ParseTLSConfig(
	ctx types.Context,
	config gofig.Config,
	fields log.Fields,
	roots ...string) (*types.TLSConfig, error) {

	ctx.Debug("parsing tls config")

	f := func(k string, v interface{}) {
		if fields == nil {
			return
		}
		fields[k] = v
	}

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
		// check to see if TLS is enabled with a simple insecure value
		if strings.EqualFold(v, "insecure") {
			f(types.ConfigTLS, "insecure")
			ctx.WithField(types.ConfigTLS, "insecure").Info("tls enabled")
			return &types.TLSConfig{
				Config: tls.Config{InsecureSkipVerify: true},
			}, nil
		}

		// check to see if TLS is enabled with an expected sha256 fingerprint
		shaRX := regexp.MustCompile(`^(?i)sha256:(.+)$`)
		if m := shaRX.FindStringSubmatch(v); len(m) > 0 {
			ctx.WithField(types.ConfigTLS, v).Info("tls enabled")
			s := strings.Join(strings.Split(m[1], ":"), "")
			buf, err := hex.DecodeString(s)
			if err != nil {
				ctx.WithError(err).Error("error decoding tls cert fingerprint")
				return nil, err
			}
			return &types.TLSConfig{
				Config:          tls.Config{InsecureSkipVerify: true},
				PeerFingerprint: buf,
			}, nil
		}
	}

	// tls is enabled; figure out its configuration
	tlsConfig := &types.TLSConfig{Config: tls.Config{}}

	// if the tls config is set to insecure, then mark it as so
	insecure := getBool(config, types.ConfigTLSInsecure, roots...)
	if insecure {
		f(types.ConfigTLSInsecure, true)
		tlsConfig.InsecureSkipVerify = true
	}

	if isSet(config, types.ConfigTLSKeyFile, roots...) {
		keyFile := getString(config, types.ConfigTLSKeyFile, roots...)
		if !gotil.FileExists(keyFile) {
			return nil, goof.WithField("path", keyFile, "invalid key file")
		}
		f(types.ConfigTLSKeyFile, keyFile)
		certFile := getString(config, types.ConfigTLSCertFile, roots...)
		if !gotil.FileExists(certFile) {
			return nil, goof.WithField("path", certFile, "invalid cert file")
		}
		f(types.ConfigTLSCertFile, certFile)
		cer, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cer}
	}

	if isSet(config, types.ConfigTLSServerName, roots...) {
		serverName := getString(config, types.ConfigTLSServerName, roots...)
		tlsConfig.ServerName = serverName
		f(types.ConfigTLSServerName, serverName)
	}

	if isSet(config, types.ConfigTLSClientCertRequired, roots...) {
		clientCertRequired := getBool(
			config, types.ConfigTLSClientCertRequired, roots...)
		if clientCertRequired {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		f(types.ConfigTLSClientCertRequired, clientCertRequired)
	}

	if isSet(config, types.ConfigTLSTrustedCertsFile, roots...) {
		trustedCertsFile := getString(
			config, types.ConfigTLSTrustedCertsFile, roots...)

		if !gotil.FileExists(trustedCertsFile) {
			return nil, goof.WithField(
				"path", trustedCertsFile, "invalid trust file")
		}

		f(types.ConfigTLSTrustedCertsFile, trustedCertsFile)

		buf, err := func() ([]byte, error) {
			f, err := os.Open(trustedCertsFile)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			buf, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, err
			}
			return buf, nil
		}()
		if err != nil {
			return nil, err
		}

		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(buf)
		tlsConfig.RootCAs = certPool
		tlsConfig.ClientCAs = certPool
	}

	return tlsConfig, nil
}
