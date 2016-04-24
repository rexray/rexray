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

	if !config.IsSet("tls") {
		return nil, nil
	}

	if !config.IsSet("tls.keyFile") {
		return nil, goof.New("keyFile required")
	}
	keyFile := config.GetString("tls.keyFile")
	if !gotil.FileExists(keyFile) {
		return nil, goof.WithField("path", keyFile, "invalid key file")
	}
	f("tls.keyFile", keyFile)

	if !config.IsSet("tls.certFile") {
		return nil, goof.New("certFile required")
	}
	certFile := config.GetString("tls.certFile")
	if !gotil.FileExists(certFile) {
		return nil, goof.WithField("path", certFile, "invalid cert file")
	}
	f("tls.certFile", certFile)

	cer, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}

	if config.IsSet("tls.serverName") {
		serverName := config.GetString("tls.serverName")
		tlsConfig.ServerName = serverName
		f("tls.serverName", serverName)
	}

	if config.IsSet("tls.clientCertRequired") {
		clientCertRequired := config.GetBool("tls.clientCertRequired")
		if clientCertRequired {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		f("tls.clientCertRequired", clientCertRequired)
	}

	if config.IsSet("tls.trustedCertsFile") {
		trustedCertsFile := config.GetString("tls.trustedCertsFile")

		if !gotil.FileExists(trustedCertsFile) {
			return nil, goof.WithField(
				"path", trustedCertsFile, "invalid trust file")
		}

		f("tls.trustedCertsFile", trustedCertsFile)

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
