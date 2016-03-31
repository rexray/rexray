package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	apiclient "github.com/emccode/libstorage/api/client"
	apiserver "github.com/emccode/libstorage/api/server"
)

var (
	tlsPath = fmt.Sprintf(
		"%s/src/github.com/emccode/libstorage/.tls", os.Getenv("GOPATH"))
	serverCrt    = fmt.Sprintf("%s/libstorage-server.crt", tlsPath)
	serverKey    = fmt.Sprintf("%s/libstorage-server.key", tlsPath)
	clientCrt    = fmt.Sprintf("%s/libstorage-client.crt", tlsPath)
	clientKey    = fmt.Sprintf("%s/libstorage-client.key", tlsPath)
	trustedCerts = fmt.Sprintf("%s/libstorage-ca.crt", tlsPath)
)

var (
	debugConfig = []byte(`
libstorage:
  server:
    http:
      logging:
        enabled: true
        logrequest: true
        logresponse: true
  client:
    http:
      logging:
        enabled: true
        logrequest: true
        logresponse: true
`)

	profilesConfig = []byte(`
libstorage:
  profiles:
    enabled: true
    groups:
    - local=127.0.0.1`)
)

// APITestFunc is a function that wraps a block of test logic for testing the
// API. An APITestFunc is executed four times:
//
//  1 - tcp
//  2 - tcp+tls
//  3 - sock
//  4 - sock+tls
type APITestFunc func(config gofig.Config, client apiclient.Client, t *testing.T)

// testHarness can be used by StorageDriver developers to quickly create
// test suites for their drivers.
type testHarness struct {
	servers []io.Closer
}

// Run executes the provided tests in a new test harness. Each test is
// executed against a new server instance.
func Run(
	t *testing.T,
	driverName string,
	config []byte,
	tests ...APITestFunc) {

	run(t, driverName, config, false, false, tests...)
}

// RunGroup executes the provided tests in a new test harness. All tests are
// executed against the same server instance.
func RunGroup(
	t *testing.T,
	driverName string,
	config []byte,
	tests ...APITestFunc) {

	run(t, driverName, config, false, true, tests...)
}

// Debug is the same as Run except with additional logging.
func Debug(
	t *testing.T,
	driverName string,
	config []byte,
	tests ...APITestFunc) {

	run(t, driverName, config, true, false, tests...)
}

// DebugGroup is the same as RunGroup except with additional logging.
func DebugGroup(
	t *testing.T,
	driverName string,
	config []byte,
	tests ...APITestFunc) {

	run(t, driverName, config, true, true, tests...)
}

func run(
	t *testing.T,
	driver string,
	configBuf []byte,
	debug, group bool,
	tests ...APITestFunc) error {

	th := &testHarness{}

	if !debug {
		debug, _ = strconv.ParseBool(os.Getenv("LIBSTORAGE_DEBUG"))
	}

	return th.run(t, driver, configBuf, debug, group, tests...)
}

func (th *testHarness) run(
	t *testing.T,
	driver string,
	configBuf []byte,
	debug, group bool,
	tests ...APITestFunc) error {

	config := gofig.New()

	if debug {
		log.SetLevel(log.DebugLevel)
		if err := config.ReadConfig(bytes.NewReader(debugConfig)); err != nil {
			return err
		}
	}

	if err := config.ReadConfig(bytes.NewReader(profilesConfig)); err != nil {
		return err
	}

	libstorageConfigMap := map[string]interface{}{
		"driver": driver,
		"server": map[string]interface{}{
			"services": map[string]interface{}{
				driver: nil,
			},
		},
	}

	initTestConfigs(libstorageConfigMap)

	config.Set("libstorage", libstorageConfigMap)

	if configBuf != nil {
		if err := config.ReadConfig(bytes.NewReader(configBuf)); err != nil {
			return err
		}
	}

	configs := []gofig.Config{
		config.Scope("libstorage.tests.tcp"),
		config.Scope("libstorage.tests.tcpTLS"),
		config.Scope("libstorage.tests.unix"),
		config.Scope("libstorage.tests.unixTLS"),
	}

	if group {
		for _, config := range configs {
			server, errs := apiserver.Serve(config)

			go func(errs <-chan error) {
				err := <-errs
				if err != nil {
					th.closeServers(t)
					t.Fatalf("server error: %v", err)
				}
			}(errs)

			if server != nil {
				th.servers = append(th.servers, server)
			}

			client, err := getClient(config)
			if err != nil {
				return err
			}

			for _, test := range tests {
				test(config, client, t)
			}
		}
	} else {
		for _, test := range tests {
			for _, config := range configs {
				server, errs := apiserver.Serve(config)

				go func(errs <-chan error) {
					err := <-errs
					if err != nil {
						th.closeServers(t)
						t.Fatalf("server error: %v", err)
					}
				}(errs)

				if server != nil {
					th.servers = append(th.servers, server)
				}

				client, err := getClient(config)
				if err != nil {
					return err
				}

				test(config, client, t)
			}
		}
	}

	th.closeServers(t)
	return nil
}

func getClient(config gofig.Config) (apiclient.Client, error) {
	client, err := apiclient.Dial(nil, config)
	if err != nil {
		return nil, goof.WithFieldE(
			"host", config.Get("libstorage.host"),
			"error dialing libStorage service", err)
	}
	return client, nil
}

func (th *testHarness) closeServers(t *testing.T) {
	for _, server := range th.servers {
		if err := server.Close(); err != nil {
			t.Errorf("error closing server: %v", err)
		}
	}
}

func initTestConfigs(config map[string]interface{}) {
	tcpHost := fmt.Sprintf("tcp://127.0.0.1:%d", gotil.RandomTCPPort())
	tcpTLSHost := fmt.Sprintf("tcp://127.0.0.1:%d", gotil.RandomTCPPort())
	unixHost := fmt.Sprintf("unix://%s", getTempSockFile())
	unixTLSHost := fmt.Sprintf("unix://%s", getTempSockFile())

	clientTLSConfig := func() map[string]interface{} {
		return map[string]interface{}{
			"serverName":       "libstorage-server",
			"certFile":         clientCrt,
			"keyFile":          clientKey,
			"trustedCertsFile": trustedCerts,
		}
	}

	serverTLSConfig := func() map[string]interface{} {
		return map[string]interface{}{
			"serverName":         "libstorage-server",
			"certFile":           serverCrt,
			"keyFile":            serverKey,
			"trustedCertsFile":   trustedCerts,
			"clientCertRequired": true,
		}
	}

	config["tests"] = map[string]interface{}{

		"tcp": map[string]interface{}{
			"libstorage": map[string]interface{}{
				"host": tcpHost,
				"server": map[string]interface{}{
					"endpoints": map[string]interface{}{
						"localhost": map[string]interface{}{
							"address": tcpHost,
						},
					},
				},
			},
		},

		"tcpTLS": map[string]interface{}{
			"libstorage": map[string]interface{}{
				"host": tcpTLSHost,
				"server": map[string]interface{}{
					"endpoints": map[string]interface{}{
						"localhost": map[string]interface{}{
							"address": tcpTLSHost,
							"tls":     serverTLSConfig(),
						},
					},
				},
				"client": map[string]interface{}{
					"tls": clientTLSConfig(),
				},
			},
		},

		"unix": map[string]interface{}{
			"libstorage": map[string]interface{}{
				"host": unixHost,
				"server": map[string]interface{}{
					"endpoints": map[string]interface{}{
						"localhost": map[string]interface{}{
							"address": unixHost,
						},
					},
				},
			},
		},

		"unixTLS": map[string]interface{}{
			"libstorage": map[string]interface{}{
				"host": unixTLSHost,
				"server": map[string]interface{}{
					"endpoints": map[string]interface{}{
						"localhost": map[string]interface{}{
							"address": unixTLSHost,
							"tls":     serverTLSConfig(),
						},
					},
				},
				"client": map[string]interface{}{
					"tls": clientTLSConfig(),
				},
			},
		},
	}
}

func getTempSockFile() string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	name := f.Name()
	os.RemoveAll(name)
	return fmt.Sprintf("%s.sock", name)
}

// LogAsJSON logs the object as JSON using the test logger.
func LogAsJSON(i interface{}, t *testing.T) {
	buf, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	t.Logf("%s\n", string(buf))
}
