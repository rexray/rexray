package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"

	apiserver "github.com/emccode/libstorage/api/server"
	"github.com/emccode/libstorage/api/server/executors"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
	"github.com/emccode/libstorage/client"
)

var (
	lsxbin string

	lsxLinuxInfo, _  = executors.ExecutorInfoInspect("lsx-linux", false)
	lsxDarwinInfo, _ = executors.ExecutorInfoInspect("lsx-darwin", false)
	// lsxWindowsInfo, _ = executors.ExecutorInfoInspect("lsx-windows.exe", false)

	tcpTest        bool
	tcpTLSTest, _  = strconv.ParseBool(os.Getenv("LIBSTORAGE_TEST_TCP_TLS"))
	sockTest, _    = strconv.ParseBool(os.Getenv("LIBSTORAGE_TEST_SOCK"))
	sockTLSTest, _ = strconv.ParseBool(os.Getenv("LIBSTORAGE_TEST_SOCK_TLS"))
)

func init() {
	goof.IncludeFieldsInFormat = true
	if runtime.GOOS == "windows" {
		lsxbin = "lsx-windows.exe"
	} else {
		lsxbin = fmt.Sprintf("lsx-%s", runtime.GOOS)
	}

	var err error
	tcpTest, err = strconv.ParseBool(os.Getenv("LIBSTORAGE_TEST_TCP"))
	if err != nil {
		tcpTest = true
	}
}

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
  readTimeout: 300
  writeTimeout: 300
  logging:
    httpRequests: true
    httpResponses: true
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
type APITestFunc func(config gofig.Config, client types.Client, t *testing.T)

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
	tests ...APITestFunc) {

	th := &testHarness{}

	if !debug {
		debug, _ = strconv.ParseBool(os.Getenv("LIBSTORAGE_DEBUG"))
	}

	th.run(t, driver, configBuf, debug, group, tests...)
}

func (th *testHarness) run(
	t *testing.T,
	driver string,
	configBuf []byte,
	debug, group bool,
	tests ...APITestFunc) {

	if !testing.Verbose() {
		buf := &bytes.Buffer{}
		log.StandardLogger().Out = buf
		defer func() {
			if t.Failed() {
				io.Copy(os.Stderr, buf)
			}
		}()
	}

	wg := &sync.WaitGroup{}

	if group {
		config := getTestConfig(t, configBuf, debug)
		configNames, configs := getTestConfigs(t, driver, config)

		for x, config := range configs {

			wg.Add(1)
			go func(x int, config gofig.Config) {
				defer wg.Done()
				server, err, errs := apiserver.Serve(config)
				if err != nil {
					th.closeServers(t)
					t.Fatal(err)
				}
				go func() {
					err := <-errs
					if err != nil {
						th.closeServers(t)
						t.Fatalf("server (%s) error: %v", configNames[x], err)
					}
				}()

				th.servers = append(th.servers, server)

				c, err := client.New(config)
				assert.NoError(t, err)
				assert.NotNil(t, c)
				if err != nil || c == nil {
					t.Fatalf("err=%v, client=%v", err, c)
				}

				for _, test := range tests {
					test(config, c, t)

					if t.Failed() {
						cj, err := config.ToJSON()
						if err != nil {
							t.Fatal(err)
						}
						fmt.Printf("client.config=%s\n", cj)
					}
				}
			}(x, config)
		}
	} else {
		for _, test := range tests {
			config := getTestConfig(t, configBuf, debug)
			configNames, configs := getTestConfigs(t, driver, config)

			for x, config := range configs {

				wg.Add(1)
				go func(test APITestFunc, x int, config gofig.Config) {

					defer wg.Done()
					server, err, errs := apiserver.Serve(config)
					if err != nil {
						th.closeServers(t)
						t.Fatal(err)
					}
					go func() {
						err := <-errs
						if err != nil {
							th.closeServers(t)
							t.Fatalf("server (%s) error: %v", configNames[x], err)
						}
					}()

					th.servers = append(th.servers, server)

					c, err := client.New(config)
					if err != nil {
						t.Fatal(err)
					}
					assert.NoError(t, err)
					assert.NotNil(t, c)

					if c == nil {
						panic(fmt.Sprintf("err=%v, client=%v", err, c))
					}

					test(config, c, t)

					if t.Failed() {
						cj, err := config.ToJSON()
						if err != nil {
							t.Fatal(err)
						}
						fmt.Printf("client.config=%s\n", cj)
					}
				}(test, x, config)
			}
		}
	}

	wg.Wait()
	th.closeServers(t)
}

func getTestConfig(t *testing.T, configBuf []byte, debug bool) gofig.Config {
	config := gofig.New()

	if debug {
		log.SetLevel(log.DebugLevel)
		err := config.ReadConfig(bytes.NewReader(debugConfig))
		if err != nil {
			t.Fatal(err)
		}
	}

	if err := config.ReadConfig(bytes.NewReader(profilesConfig)); err != nil {
		t.Fatal(err)
	}

	if configBuf != nil {
		if err := config.ReadConfig(bytes.NewReader(configBuf)); err != nil {
			t.Fatal(err)
		}
	}

	return config
}

func getTestConfigs(
	t *testing.T,
	driver string,
	config gofig.Config) (map[int]string, []gofig.Config) {

	libstorageConfigMap := map[string]interface{}{
		"server": map[string]interface{}{
			"services": map[string]interface{}{
				driver: map[string]interface{}{
					"libstorage": map[string]interface{}{
						"storage": map[string]interface{}{
							"driver": driver,
						},
					},
				},
			},
		},
	}

	initTestConfigs(libstorageConfigMap)

	libstorageConfig := map[string]interface{}{
		"libstorage": libstorageConfigMap,
	}

	yamlBuf, err := yaml.Marshal(libstorageConfig)
	assert.NoError(t, err)
	assert.NoError(t, config.ReadConfig(bytes.NewReader(yamlBuf)))

	configNames := map[int]string{}
	configs := []gofig.Config{}

	if tcpTest {
		configNames[len(configNames)] = "tcp"
		configs = append(configs, config.Scope("libstorage.tests.tcp"))
	}
	if tcpTLSTest {
		configNames[len(configNames)] = "tcpTLS"
		configs = append(configs, config.Scope("libstorage.tests.tcpTLS"))
	}
	if sockTest {
		configNames[len(configNames)] = "unix"
		configs = append(configs, config.Scope("libstorage.tests.unix"))
	}
	if sockTLSTest {
		configNames[len(configNames)] = "unixTLS"
		configs = append(configs, config.Scope("libstorage.tests.unixTLS"))
	}

	return configNames, configs
}

func (th *testHarness) closeServers(t *testing.T) {
	for _, server := range th.servers {
		if server == nil {
			panic("testharness.server is nil")
		}
		if err := server.Close(); err != nil {
			t.Fatalf("error closing server: %v", err)
		}
	}
}

func initTestConfigs(config map[string]interface{}) {
	tcpHost := fmt.Sprintf("tcp://127.0.0.1:%d", gotil.RandomTCPPort())
	tcpTLSHost := fmt.Sprintf("tcp://127.0.0.1:%d", gotil.RandomTCPPort())
	unixHost := fmt.Sprintf("unix://%s", utils.GetTempSockFile())
	unixTLSHost := fmt.Sprintf("unix://%s", utils.GetTempSockFile())

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

// LogAsJSON logs the object as JSON using the test logger.
func LogAsJSON(i interface{}, t *testing.T) {
	buf, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	t.Logf("%s\n", string(buf))
}
