package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	apiconfig "github.com/AVENTER-UG/rexray/libstorage/api/utils/config"
)

type suiteRunner struct {
	t          *testing.T
	driverName string
	proto      string
}

func newSuiteRunner(t *testing.T, d, p string) (string, func()) {
	sr := &suiteRunner{t: t, driverName: d, proto: p}
	return fmt.Sprintf("%s (%s)", d, p), sr.Describe
}

func newTestRunner(driverName string) *testRunner {
	return &testRunner{
		store:           utils.NewStore(),
		driverName:      driverName,
		expectedVolID:   v0ID,
		expectedVolName: v0Name,
		expectedNextDev: v0NextDev,
	}
}

type testRunner struct {
	err             error
	ctx             types.Context
	config          gofig.Config
	configFileData  []byte
	pathConfig      *types.PathConfig
	client          types.Client
	server          types.Server
	srvErr          <-chan error
	store           types.Store
	vol             *types.Volume
	driverName      string
	sysHome         string
	usrHome         string
	proto           string
	laddr           string
	serverCrt       string
	serverKey       string
	clientCrt       string
	clientKey       string
	cacerts         string
	knownHosts      string
	volID           string
	volName         string
	nextDev         string
	expectedVolID   string
	expectedVolName string
	expectedNextDev string
}

func (t *testRunner) beforeEach() {
	t.serverCrt = suiteServerCrt
	t.serverKey = suiteServerKey
	t.clientCrt = suiteClientCrt
	t.clientKey = suiteClientKey
	t.cacerts = suiteTrustedCerts
	t.knownHosts = suiteKnownHosts
	t.volID = v0ID
	t.volName = v0Name
	t.nextDev = v0NextDev

	// get temp dirs for the sys and user home dirs
	newTempDir(&t.sysHome)
	newTempDir(&t.usrHome)

	// create a context and path config then process the
	// registered config registrations
	t.ctx = context.Background()
	Ω(t.ctx).ShouldNot(BeNil())
	t.pathConfig = utils.NewPathConfig(t.sysHome, "", t.usrHome)
	Ω(t.pathConfig).ShouldNot(BeNil())
	t.ctx = context.WithValue(t.ctx, context.PathConfigKey, t.pathConfig)
	registry.ProcessRegisteredConfigs(t.ctx)
}

func (t *testRunner) afterEach() {
	os.RemoveAll(t.sysHome)
	os.RemoveAll(t.usrHome)
	t.ctx = nil
	t.vol = nil
	t.sysHome = ""
	t.usrHome = ""
	t.proto = ""
	t.laddr = ""
	t.client = nil
	t.pathConfig = nil
	t.configFileData = nil
	t.err = nil
	t.serverCrt = ""
	t.serverKey = ""
	t.clientCrt = ""
	t.clientKey = ""
	t.cacerts = ""
	t.knownHosts = ""
	t.volID = ""
	t.volName = ""
	t.nextDev = ""

	if t.server != nil {
		Ω(t.server.Close()).ToNot(HaveOccurred())
		t.server = nil
		Ω(<-t.srvErr).ToNot(HaveOccurred())
	}

	os.Setenv("LIBSTORAGE_TLS_SOCKITTOME", "")
	os.Setenv("LIBSTORAGE_TLS_SERVERNAME", "")
	os.Setenv("LIBSTORAGE_TLS_CERTFILE", "")
	os.Setenv("LIBSTORAGE_TLS_KEYFILE", "")
	os.Setenv("LIBSTORAGE_TLS_TRUSTEDCERTSFILE", "")
}

func (t *testRunner) justBeforeEach() {
	// write the config data to disk
	configFilePath := path.Join(t.pathConfig.Etc, "config.yml")
	Ω(configFilePath).ShouldNot(BeAnExistingFile())
	Ω(ioutil.WriteFile(
		configFilePath,
		t.configFileData,
		0644)).ToNot(HaveOccurred())
	Ω(configFilePath).Should(BeARegularFile())

	// create a new config object
	t.config, t.err = apiconfig.NewConfig(t.ctx)
	Ω(t.err).ToNot(HaveOccurred())
	Ω(t.config).ShouldNot(BeNil())
}
