package utils

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/emccode/libstorage/api/types"
)

// ParseTLSConfig returns a new TLS configuration.
func ParseTLSConfig(
	config gofig.Config, fields log.Fields) (*tls.Config, error) {

	f := func(k string, v interface{}) {
		if fields == nil {
			return
		}
		fields[k] = v
	}

	if !config.IsSet(types.ConfigTLS) {
		return nil, nil
	}

	if config.IsSet(types.ConfigTLSDisabled) {
		tlsDisabled := config.GetBool(types.ConfigTLSDisabled)
		if tlsDisabled {
			f(types.ConfigTLSDisabled, true)
			return nil, nil
		}
	}

	if !config.IsSet(types.ConfigTLSKeyFile) {
		return nil, goof.New("keyFile required")
	}
	keyFile := config.GetString(types.ConfigTLSKeyFile)
	if !gotil.FileExists(keyFile) {
		return nil, goof.WithField("path", keyFile, "invalid key file")
	}
	f(types.ConfigTLSKeyFile, keyFile)

	if !config.IsSet(types.ConfigTLSCertFile) {
		return nil, goof.New("certFile required")
	}
	certFile := config.GetString(types.ConfigTLSCertFile)
	if !gotil.FileExists(certFile) {
		return nil, goof.WithField("path", certFile, "invalid cert file")
	}
	f(types.ConfigTLSCertFile, certFile)

	cer, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}

	if config.IsSet(types.ConfigTLSServerName) {
		serverName := config.GetString(types.ConfigTLSServerName)
		tlsConfig.ServerName = serverName
		f(types.ConfigTLSServerName, serverName)
	}

	if config.IsSet(types.ConfigTLSClientCertRequired) {
		clientCertRequired := config.GetBool(types.ConfigTLSClientCertRequired)
		if clientCertRequired {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		f(types.ConfigTLSClientCertRequired, clientCertRequired)
	}

	if config.IsSet(types.ConfigTLSTrustedCertsFile) {
		trustedCertsFile := config.GetString(types.ConfigTLSTrustedCertsFile)

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
