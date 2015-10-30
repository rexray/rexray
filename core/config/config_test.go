package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	//log "github.com/Sirupsen/logrus"

	"github.com/emccode/rexray/util"
)

var (
	tmpPrefixDirs []string
	usrRexRayFile string
)

func TestMain(m *testing.M) {
	//log.SetLevel(log.DebugLevel)
	usrRexRayDir := fmt.Sprintf("%s/.rexray", util.HomeDir())
	os.MkdirAll(usrRexRayDir, 0755)
	usrRexRayFile = fmt.Sprintf("%s/%s.%s", usrRexRayDir, "config", "yml")
	usrRexRayFileBak := fmt.Sprintf("%s.bak", usrRexRayFile)

	os.Remove(usrRexRayFileBak)
	os.Rename(usrRexRayFile, usrRexRayFileBak)

	exitCode := m.Run()
	for _, d := range tmpPrefixDirs {
		os.RemoveAll(d)
	}

	os.Remove(usrRexRayFile)
	os.Rename(usrRexRayFileBak, usrRexRayFile)
	os.Exit(exitCode)
}

func newPrefixDir(testName string, t *testing.T) string {
	tmpDir, err := ioutil.TempDir(
		"", fmt.Sprintf("rexray-core-config_test-%s", testName))
	if err != nil {
		t.Fatal(err)
	}

	util.Prefix(tmpDir)
	os.MkdirAll(tmpDir, 0755)
	tmpPrefixDirs = append(tmpPrefixDirs, tmpDir)
	return tmpDir
}

func TestAssertConfigDefaults(t *testing.T) {
	newPrefixDir("TestAssertConfigDefaults", t)
	wipeEnv()
	c := New()

	osDrivers := c.GetStringSlice("osDrivers")
	volDrivers := c.GetStringSlice("volumeDrivers")

	assertString(t, c, "host", "tcp://:7979")
	assertString(t, c, "logLevel", "warn")

	if len(osDrivers) != 1 || osDrivers[0] != "linux" {
		t.Fatalf("osDrivers != []string{\"linux\"}, == %v", osDrivers)
	}

	if len(volDrivers) != 1 || volDrivers[0] != "docker" {
		t.Fatalf("volumeDrivers != []string{\"docker\"}, == %v", volDrivers)
	}
}

func TestAssertTestRegistration(t *testing.T) {
	newPrefixDir("TestAssertTestRegistration", t)
	wipeEnv()
	Register(testRegistration())
	c := New()

	userName := c.GetString("mockProvider.username")
	password := c.GetString("mockProvider.password")
	useCerts := c.GetBool("mockProvider.useCerts")
	minVolSize := c.GetInt("mockProvider.Docker.minVolSize")

	if userName != "admin" {
		t.Fatalf("mockProvider.userName != admin, == %s", userName)
	}

	if password != "" {
		t.Fatalf("mockProvider.password != '', == %s", password)
	}

	if !useCerts {
		t.Fatalf("mockProvider.useCerts != true, == %v", useCerts)
	}

	if minVolSize != 16 {
		t.Fatalf("minVolSize != 16, == %d", minVolSize)
	}
}

func TestBaselineJSON(t *testing.T) {
	newPrefixDir("TestBaselineJSON", t)
	wipeEnv()
	Register(testRegistration())
	c := New()

	var err error
	var cJSON string
	if cJSON, err = c.ToJSON(); err != nil {
		t.Fatal(err)
	}

	cMap := map[string]interface{}{}
	ccMap := map[string]interface{}{}

	if err := json.Unmarshal([]byte(cJSON), &cMap); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(
		[]byte(jsonConfigBaseline), &ccMap); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(cMap, ccMap) {
		t.Fail()
	}

	if reflect.DeepEqual(map[string]interface{}{}, ccMap) {
		t.Fail()
	}

	if reflect.DeepEqual(cMap, map[string]interface{}{}) {
		t.Fail()
	}
}

func TestToJSON(t *testing.T) {
	newPrefixDir("TestToJSON", t)
	wipeEnv()
	Register(testRegistration())
	c := New()

	if err := c.ReadConfig(bytes.NewReader(yamlConfig1)); err != nil {
		t.Fatal(err)
	}

	var err error
	var cJSON string
	if cJSON, err = c.ToJSON(); err != nil {
		t.Fatal(err)
	}

	t.Log(cJSON)
	t.Log(jsonConfigWithYamlConfig1)

	cMap := map[string]interface{}{}
	ccMap := map[string]interface{}{}

	if err := json.Unmarshal([]byte(cJSON), &cMap); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(
		[]byte(jsonConfigWithYamlConfig1), &ccMap); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(cMap, ccMap) {
		t.Fatal("json not equal pre minVolSize change")
	}

	mvs := c.GetInt("mockprovider.docker.minvolsize")
	if mvs != 32 {
		t.Fatal("mvs != 32")
	}

	c.Set("mockprovider.docker.minvolsize", 128)
	mvs = c.GetInt("mockprovider.docker.minvolsize")
	if mvs != 128 {
		t.Fatal("mvs != 128")
	}

	if cJSON, err = c.ToJSON(); err != nil {
		t.Fatal(err)
	}

	cMap = map[string]interface{}{}

	if err := json.Unmarshal([]byte(cJSON), &cMap); err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(cMap, ccMap) {
		t.Fatal("json equal post minVolSize change")
	}
}

func TestToJSONCompact(t *testing.T) {
	newPrefixDir("TestToJSONCompact", t)
	wipeEnv()
	Register(testRegistration())
	c := New()

	if err := c.ReadConfig(bytes.NewReader(yamlConfig1)); err != nil {
		t.Fatal(err)
	}

	var err error
	var cJSON string
	if cJSON, err = c.ToJSONCompact(); err != nil {
		t.Fatal(err)
	}

	cMap := map[string]interface{}{}
	ccMap := map[string]interface{}{}

	if err := json.Unmarshal([]byte(cJSON), &cMap); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(
		[]byte(jsonConfigWithYamlConfig1), &ccMap); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(cMap, ccMap) {
		t.Fail()
	}

	mvs := c.GetInt("mockprovider.docker.minvolsize")
	if mvs != 32 {
		t.Fatal("mvs != 32")
	}

	c.Set("mockprovider.docker.minvolsize", 128)
	mvs = c.GetInt("mockprovider.docker.minvolsize")
	if mvs != 128 {
		t.Fatal("mvs != 128")
	}

	if cJSON, err = c.ToJSONCompact(); err != nil {
		t.Fatal(err)
	}

	cMap = map[string]interface{}{}

	if err := json.Unmarshal([]byte(cJSON), &cMap); err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(cMap, ccMap) {
		t.Fail()
	}
}

func TestFromJSON(t *testing.T) {
	newPrefixDir("TestFromJSON", t)
	wipeEnv()
	Register(testRegistration())

	c := New()

	if err := c.ReadConfig(bytes.NewReader(yamlConfig1)); err != nil {
		t.Fatal(err)
	}

	var err error
	var cJSON string
	if cJSON, err = c.ToJSON(); err != nil {
		t.Fatal(err)
	}

	var cfj *Config
	var cfJSON string
	if cfj, err = FromJSON(jsonConfigWithYamlConfig1); err != nil {
		t.Fatal(err)
	}
	if cfJSON, err = cfj.ToJSON(); err != nil {
		t.Fatal(err)
	}

	cMap := map[string]interface{}{}
	cfjMap := map[string]interface{}{}

	if err := json.Unmarshal([]byte(cJSON), &cMap); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(cfJSON), &cfjMap); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(cMap, cfjMap) {
		t.Fail()
	}
}

func TestFromJSONWithErrors(t *testing.T) {
	_, err := FromJSON("///*")
	if err == nil {
		t.Fatal("expected unmarshalling error")
	}
}

func TestEnvVars(t *testing.T) {
	newPrefixDir("TestEnvVars", t)
	wipeEnv()
	Register(testRegistration())
	c := New()

	if err := c.ReadConfig(bytes.NewReader(yamlConfig1)); err != nil {
		t.Fatal(err)
	}

	fev := c.EnvVars()

	for _, v := range fev {
		t.Log(v)
	}

	assertEnvVar("REXRAY_HOST=tcp://:7979", fev, t)
	assertEnvVar("REXRAY_LOGLEVEL=error", fev, t)
	assertEnvVar("REXRAY_STORAGEDRIVERS=ec2 xtremio", fev, t)
	assertEnvVar("REXRAY_OSDRIVERS=linux", fev, t)
	assertEnvVar("REXRAY_VOLUMEDRIVERS=docker", fev, t)
	assertEnvVar("MOCKPROVIDER_USERNAME=admin", fev, t)
	assertEnvVar("MOCKPROVIDER_USECERTS=true", fev, t)
	assertEnvVar("MOCKPROVIDER_DOCKER_MINVOLSIZE=32", fev, t)
}

func assertEnvVar(s string, evs []string, t *testing.T) {
	if !util.StringInSlice(s, evs) {
		t.Fatal(s)
	}
}

func TestCopy(t *testing.T) {
	newPrefixDir("TestCopy", t)
	wipeEnv()
	Register(testRegistration())

	etcRexRayCfg := util.EtcFilePath("config.yml")
	t.Logf("etcRexRayCfg=%s", etcRexRayCfg)
	util.WriteStringToFile(string(yamlConfig1), etcRexRayCfg)

	c := New()

	assertString(t, c, "logLevel", "error")
	assertStorageDrivers(t, c)
	assertOsDrivers1(t, c)

	cc, _ := c.Copy()

	assertString(t, cc, "logLevel", "error")
	assertStorageDrivers(t, cc)
	assertOsDrivers1(t, cc)

	cJSON, _ := c.ToJSON()
	ccJSON, _ := cc.ToJSON()

	cMap := map[string]interface{}{}
	ccMap := map[string]interface{}{}

	if err := json.Unmarshal([]byte(cJSON), &cMap); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(ccJSON), &ccMap); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(cMap, ccMap) {
		t.Fail()
	}
}

func TestNewWithUserConfigFile(t *testing.T) {
	util.WriteStringToFile(string(yamlConfig1), usrRexRayFile)
	defer os.RemoveAll(usrRexRayFile)

	c := New()

	assertString(t, c, "logLevel", "error")
	assertStorageDrivers(t, c)
	assertOsDrivers1(t, c)

	if err := c.ReadConfig(bytes.NewReader(yamlConfig2)); err != nil {
		t.Fatal(err)
	}

	assertString(t, c, "logLevel", "debug")
	assertStorageDrivers(t, c)
	assertOsDrivers2(t, c)
}

func TestNewWithUserConfigFileWithErrors(t *testing.T) {
	util.WriteStringToFile(string(yamlConfig1), usrRexRayFile)
	defer os.RemoveAll(usrRexRayFile)

	os.Chmod(usrRexRayFile, 0000)
	New()
}

func TestNewWithGlobalConfigFile(t *testing.T) {
	newPrefixDir("TestNewWithGlobalConfigFile", t)

	etcRexRayCfg := util.EtcFilePath("config.yml")
	t.Logf("etcRexRayCfg=%s", etcRexRayCfg)
	util.WriteStringToFile(string(yamlConfig1), etcRexRayCfg)

	c := New()

	assertString(t, c, "logLevel", "error")
	assertStorageDrivers(t, c)
	assertOsDrivers1(t, c)

	if err := c.ReadConfig(bytes.NewReader(yamlConfig2)); err != nil {
		t.Fatal(err)
	}

	assertString(t, c, "logLevel", "debug")
	assertStorageDrivers(t, c)
	assertOsDrivers2(t, c)
}

func TestNewWithGlobalConfigFileWithErrors(t *testing.T) {
	newPrefixDir("TestNewWithGlobalConfigFileWithErrors", t)

	etcRexRayCfg := util.EtcFilePath("config.yml")
	t.Logf("etcRexRayCfg=%s", etcRexRayCfg)
	util.WriteStringToFile(string(yamlConfig1), etcRexRayCfg)

	os.Chmod(etcRexRayCfg, 0000)
	New()
}

func TestReadConfigFile(t *testing.T) {
	var err error
	var tmp *os.File
	if tmp, err = ioutil.TempFile("", "TestReadConfigFile"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(yamlConfig1); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	os.Chmod(tmp.Name(), 0000)

	c := New()
	if err := c.ReadConfigFile(tmp.Name()); err == nil {
		t.Fatal("expected error reading config file")
	}
}

func TestReadConfig(t *testing.T) {
	c := NewConfig(false, false, "config", "yml")
	if err := c.ReadConfig(bytes.NewReader(yamlConfig1)); err != nil {
		t.Fatal(err)
	}

	assertString(t, c, "logLevel", "error")
	assertStorageDrivers(t, c)
	assertOsDrivers1(t, c)

	if err := c.ReadConfig(bytes.NewReader(yamlConfig2)); err != nil {
		t.Fatal(err)
	}

	assertString(t, c, "logLevel", "debug")
	assertStorageDrivers(t, c)
	assertOsDrivers2(t, c)
}

func TestReadNilConfig(t *testing.T) {
	if err := New().ReadConfig(nil); err == nil {
		t.Fatal("expected nil config error")
	}
}

func wipeEnv() {
	evs := os.Environ()
	for _, v := range evs {
		k := strings.Split(v, "=")[0]
		os.Setenv(k, "")
	}
}

func printConfig(c *Config, t *testing.T) {
	for k, v := range c.v.AllSettings() {
		t.Logf("%s=%v\n", k, v)
	}
}

func testRegistration() *Registration {
	r := NewRegistration("Mock Provider")
	r.Yaml(`mockProvider:
    userName: admin
    useCerts: true
    docker:
        MinVolSize: 16
`)
	r.Key(String, "", "admin", "", "mockProvider.userName")
	r.Key(String, "", "", "", "mockProvider.password")
	r.Key(Bool, "", false, "", "mockProvider.useCerts")
	r.Key(Int, "", 16, "", "mockProvider.docker.minVolSize")
	r.Key(Bool, "i", true, "", "mockProvider.insecure")
	r.Key(Int, "m", 256, "", "mockProvider.docker.maxVolSize")
	return r
}

func assertString(t *testing.T, c *Config, key, expected string) {
	v := c.GetString(key)
	if v != expected {
		t.Fatalf("%s != %s; == %v", key, expected, v)
	}
}

func assertStorageDrivers(t *testing.T, c *Config) {
	sd := c.GetStringSlice("storageDrivers")
	if sd == nil {
		t.Fatalf("storageDrivers == nil")
	}

	if len(sd) != 2 {
		t.Fatalf("len(storageDrivers) != 2; == %d", len(sd))
	}

	if sd[0] != "ec2" {
		t.Fatalf("sd[0] != ec2; == %v", sd[0])
	}

	if sd[1] != "xtremio" {
		t.Fatalf("sd[1] != xtremio; == %v", sd[1])
	}
}

func assertOsDrivers1(t *testing.T, c *Config) {
	od := c.GetStringSlice("osDrivers")
	if od == nil {
		t.Fatalf("osDrivers == nil")
	}
	if len(od) != 1 {
		t.Fatalf("len(osDrivers) != 1; == %d", len(od))
	}
	if od[0] != "linux" {
		t.Fatalf("od[0] != linux; == %v", od[0])
	}
}

func assertOsDrivers2(t *testing.T, c *Config) {
	od := c.GetStringSlice("osDrivers")
	if od == nil {
		t.Fatalf("osDrivers == nil")
	}
	if len(od) != 2 {
		t.Fatalf("len(osDrivers) != 2; == %d", len(od))
	}
	if od[0] != "darwin" {
		t.Fatalf("od[0] != darwin; == %v", od[0])
	}
	if od[1] != "linux" {
		t.Fatalf("od[1] != linux; == %v", od[1])
	}
}

var yamlConfig1 = []byte(`logLevel: error
storageDrivers:
- ec2
- xtremio
osDrivers:
- linux
mockProvider:
  userName: admin
  useCerts: true
  docker:
    MinVolSize: 32
`)

var yamlConfig2 = []byte(`logLevel: debug
osDrivers:
- darwin
- linux
`)

var jsonConfigBaseline = `{
    "host": "tcp://:7979",
    "loglevel": "warn",
    "mockprovider": {
        "docker": {
            "MinVolSize": 16
        },
        "useCerts": true,
        "userName": "admin"
    },
    "osdrivers": [
        "linux"
    ],
    "volumedrivers": [
        "docker"
    ]
}
`

var jsonConfigWithYamlConfig1 = `{
    "host": "tcp://:7979",
    "loglevel": "error",
    "mockprovider": {
        "docker": {
            "MinVolSize": 32
        },
        "useCerts": true,
        "userName": "admin"
    },
    "osdrivers": [
        "linux"
    ],
    "storagedrivers": [
        "ec2",
        "xtremio"
    ],
    "volumedrivers": [
        "docker"
    ]
}
`
