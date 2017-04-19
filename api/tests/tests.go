package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"testing"

	log "github.com/Sirupsen/logrus"
	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"

	"github.com/akutz/goof"
	"github.com/akutz/gotil"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"

	apiserver "github.com/codedellemc/libstorage/api/server"
	"github.com/codedellemc/libstorage/api/server/executors"
	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"
	"github.com/codedellemc/libstorage/client"
)

var (
	lsxbin string

	lsxLinuxInfo, _  = executors.ExecutorInfoInspect("lsx-linux", false)
	lsxDarwinInfo, _ = executors.ExecutorInfoInspect("lsx-darwin", false)
	// lsxWindowsInfo, _ = executors.ExecutorInfoInspect("lsx-windows.exe", false)

	tcpTest bool

	tcpTLSTest, _ = strconv.ParseBool(
		os.Getenv("LIBSTORAGE_TEST_TCP_TLS"))
	tcpTLSPeersTest, _ = strconv.ParseBool(
		os.Getenv("LIBSTORAGE_TEST_TCP_TLS_PEERS"))

	sockTest, _    = strconv.ParseBool(os.Getenv("LIBSTORAGE_TEST_SOCK"))
	sockTLSTest, _ = strconv.ParseBool(os.Getenv("LIBSTORAGE_TEST_SOCK_TLS"))

	printConfigOnFail, _ = strconv.ParseBool(os.Getenv(
		"LIBSTORAGE_TEST_PRINT_CONFIG_ON_FAIL"))
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
	tlsPath = path.Join(
		os.Getenv("GOPATH"),
		"/src/github.com/codedellemc/libstorage/.tls")
	serverCrt    = path.Join(tlsPath, "libstorage-server.crt")
	serverKey    = path.Join(tlsPath, "libstorage-server.key")
	clientCrt    = path.Join(tlsPath, "libstorage-client.crt")
	clientKey    = path.Join(tlsPath, "libstorage-client.key")
	trustedCerts = path.Join(tlsPath, "libstorage-ca.crt")
	knownHosts   = path.Join(tlsPath, "known_hosts")
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

	run(t, types.IntegrationClient, nil,
		driverName, config, false, false, tests...)
}

// RunWithOnClientError executes the provided tests in a new test harness with
// the specified on client error delegate. Each test is executed against a new
// server instance.
func RunWithOnClientError(
	t *testing.T,
	onClientError func(error),
	driverName string,
	config []byte,
	tests ...APITestFunc) {

	run(t, types.IntegrationClient, onClientError,
		driverName, config, false, false, tests...)
}

// RunWithClientType executes the provided tests in a new test harness with
// the specified client type. Each test is executed against a new server
// instance.
func RunWithClientType(
	t *testing.T,
	clientType types.ClientType,
	driverName string,
	config []byte,
	tests ...APITestFunc) {

	run(t, clientType, nil, driverName, config, false, false, tests...)
}

// RunGroup executes the provided tests in a new test harness. All tests are
// executed against the same server instance.
func RunGroup(
	t *testing.T,
	driverName string,
	config []byte,
	tests ...APITestFunc) {

	run(t, types.IntegrationClient, nil,
		driverName, config, false, true, tests...)
}

// RunGroupWithClientType executes the provided tests in a new test harness
// with the specified client type. All tests are executed against the same
// server instance.
func RunGroupWithClientType(
	t *testing.T,
	clientType types.ClientType,
	driverName string,
	config []byte,
	tests ...APITestFunc) {

	run(t, clientType, nil, driverName, config, false, true, tests...)
}

// Debug is the same as Run except with additional logging.
func Debug(
	t *testing.T,
	driverName string,
	config []byte,
	tests ...APITestFunc) {

	run(t, types.IntegrationClient, nil,
		driverName, config, true, false, tests...)
}

// DebugGroup is the same as RunGroup except with additional logging.
func DebugGroup(
	t *testing.T,
	driverName string,
	config []byte,
	tests ...APITestFunc) {

	run(
		t, types.IntegrationClient, nil,
		driverName, config, true, true, tests...)
}

func run(
	t *testing.T,
	clientType types.ClientType,
	onNewClientError func(err error),
	driver string,
	configBuf []byte,
	debug, group bool,
	tests ...APITestFunc) {

	th := &testHarness{}

	if !debug {
		debug, _ = strconv.ParseBool(os.Getenv("LIBSTORAGE_DEBUG"))
	}

	th.run(t, clientType, onNewClientError,
		driver, configBuf, debug, group, tests...)
}

func (th *testHarness) run(
	t *testing.T,
	clientType types.ClientType,
	onNewClientError func(error),
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
		config := getTestConfig(t, clientType, configBuf, debug)
		configNames, configs := getTestConfigs(t, driver, config)

		for x, config := range configs {

			wg.Add(1)
			go func(x int, config gofig.Config) {

				defer wg.Done()
				server, errs, err := apiserver.Serve(nil, config)
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

				c, err := client.New(nil, config)
				if onNewClientError != nil {
					onNewClientError(err)
				} else if err != nil {
					t.Fatal(err)
				} else if !assert.NotNil(t, c) {
					t.FailNow()
				} else {
					for _, test := range tests {
						test(config, c, t)
						if t.Failed() && printConfigOnFail {
							cj, err := config.ToJSON()
							if err != nil {
								t.Fatal(err)
							}
							fmt.Printf("client.config=%s\n", cj)
						}
					}
				}
			}(x, config)
		}
	} else {
		for _, test := range tests {
			config := getTestConfig(t, clientType, configBuf, debug)
			configNames, configs := getTestConfigs(t, driver, config)

			for x, config := range configs {

				wg.Add(1)
				go func(test APITestFunc, x int, config gofig.Config) {

					defer wg.Done()
					server, errs, err := apiserver.Serve(nil, config)
					if err != nil {
						th.closeServers(t)
						t.Fatal(err)
					}
					go func() {
						err := <-errs
						if err != nil {
							th.closeServers(t)
							t.Fatalf(
								"server (%s) error: %v",
								configNames[x], err)
						}
					}()

					th.servers = append(th.servers, server)

					c, err := client.New(nil, config)
					if onNewClientError != nil {
						onNewClientError(err)
					} else if err != nil {
						t.Fatal(err)
					} else if !assert.NotNil(t, c) {
						t.FailNow()
					} else {
						test(config, c, t)
					}

					if t.Failed() && printConfigOnFail {
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

var clientTypeConfigFormat = `
libstorage:
  client:
    type: %s
`

func getTestConfig(
	t *testing.T,
	clientType types.ClientType,
	configBuf []byte,
	debug bool) gofig.Config {

	config := gofigCore.New()

	if debug {
		log.SetLevel(log.DebugLevel)
		err := config.ReadConfig(bytes.NewReader(debugConfig))
		if err != nil {
			t.Fatal(err)
		}
	}

	clientTypeConfig := []byte(fmt.Sprintf(clientTypeConfigFormat, clientType))
	if err := config.ReadConfig(bytes.NewReader(clientTypeConfig)); err != nil {
		t.Fatal(err)
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
		configs = append(configs, config.Scope(
			"libstorage.tests.tcp").Scope(
			"testing"))
	}
	if tcpTLSTest {
		configNames[len(configNames)] = "tcpTLS"
		configs = append(configs, config.Scope(
			"libstorage.tests.tcpTLS").Scope(
			"test"))
	}
	if tcpTLSPeersTest {
		configNames[len(configNames)] = "tcpTLSPeers"
		configs = append(configs, config.Scope(
			"libstorage.tests.tcpTLSPeers").Scope(
			"test"))
	}
	if sockTest {
		configNames[len(configNames)] = "unix"
		configs = append(configs, config.Scope(
			"libstorage.tests.unix").Scope(
			"test"))
	}
	if sockTLSTest {
		configNames[len(configNames)] = "unixTLS"
		configs = append(configs, config.Scope(
			"libstorage.tests.unixTLS").Scope(
			"test"))
	}
	if sockTLSTest {
		configNames[len(configNames)] = "unixTLS"
		configs = append(configs, config.Scope(
			"libstorage.tests.unixTLSPeers").Scope(
			"test"))
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

	clientTLSConfig := func(peers bool) map[string]interface{} {
		if peers {
			return map[string]interface{}{
				"verifyPeers": true,
				"knownHosts":  knownHosts,
			}
		}
		return map[string]interface{}{
			"serverName":       "libstorage-server",
			"certFile":         clientCrt,
			"keyFile":          clientKey,
			"trustedCertsFile": trustedCerts,
		}
	}

	serverTLSConfig := func(clientCertRequired bool) map[string]interface{} {
		return map[string]interface{}{
			"serverName":         "libstorage-server",
			"certFile":           serverCrt,
			"keyFile":            serverKey,
			"trustedCertsFile":   trustedCerts,
			"clientCertRequired": clientCertRequired,
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
							"tls":     serverTLSConfig(true),
						},
					},
				},
				"client": map[string]interface{}{
					"tls": clientTLSConfig(false),
				},
			},
		},

		"tcpTLSPeers": map[string]interface{}{
			"libstorage": map[string]interface{}{
				"host": tcpTLSHost,
				"server": map[string]interface{}{
					"endpoints": map[string]interface{}{
						"localhost": map[string]interface{}{
							"address": tcpTLSHost,
							"tls":     serverTLSConfig(false),
						},
					},
				},
				"client": map[string]interface{}{
					"tls": clientTLSConfig(true),
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
							"tls":     serverTLSConfig(true),
						},
					},
				},
				"client": map[string]interface{}{
					"tls": clientTLSConfig(false),
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
