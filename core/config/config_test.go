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

	"github.com/emccode/rexray/util"
)

var (
	tmpPrefixDirs []string
	usrRexRayFile string
)

func TestMain(m *testing.M) {
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

	evs := os.Environ()
	for _, v := range evs {
		k := strings.Split(v, "=")[0]
		os.Setenv(k, "")
	}

	c := New()

	if c.Host != "tcp://:7979" {
		t.Fatalf("c.Host != tcp://:7979, == %s", c.Host)
	}

	if c.LogLevel != "info" {
		t.Fatalf("c.LogLevel != info, == %d", c.LogLevel)
	}

	if len(c.OSDrivers) != 1 && c.OSDrivers[0] != "linux" {
		t.Fatalf("c.OSDrivers != []string{\"linux\"}, == %v", c.OSDrivers)
	}

	if len(c.VolumeDrivers) != 1 && c.VolumeDrivers[0] != "docker" {
		t.Fatalf("c.VolumeDrivers != []string{\"docker\"}, == %v", c.VolumeDrivers)
	}

	if c.DockerSize != 0 {
		t.Fatalf("c.DockerSize != 0, == %d", c.DockerSize)
	}
}

func TestCopy(t *testing.T) {
	newPrefixDir("TestCopy", t)

	etcRexRayCfg := util.EtcFilePath("config.yml")
	t.Logf("etcRexRayCfg=%s", etcRexRayCfg)
	util.WriteStringToFile(string(yamlConfig1), etcRexRayCfg)

	c := New()

	assertLogLevel(t, c, "error")
	assertStorageDrivers(t, c)
	assertOsDrivers1(t, c)

	cc, _ := c.Copy()

	assertLogLevel(t, cc, "error")
	assertStorageDrivers(t, cc)
	assertOsDrivers1(t, cc)

	cJSON, _ := c.ToJSON()
	ccJSON, _ := cc.ToJSON()

	cMap := map[string]interface{}{}
	ccMap := map[string]interface{}{}
	json.Unmarshal([]byte(cJSON), cMap)
	json.Unmarshal([]byte(ccJSON), ccJSON)

	if !reflect.DeepEqual(cMap, ccMap) {
		t.Fail()
	}
}

func TestEnvVars(t *testing.T) {
	c := New()
	c.Viper.Set("awsSecretKey", "Hello, world.")
	if !util.StringInSlice("AWS_SECRET_KEY=Hello, world.", c.EnvVars()) {
		t.Fail()
	}

	if util.StringInSlice("AWS_SECRET_KEY=Hello, world.", New().EnvVars()) {
		t.Fail()
	}
}

func TestJSONMarshalStrategy(t *testing.T) {
	c := New()
	if c.JSONMarshalStrategy() != JSONMarshalSecure {
		t.Fail()
	}
	c.SetJSONMarshalStrategy(JSONMarshalPlainText)
	if c.JSONMarshalStrategy() != JSONMarshalPlainText {
		t.Fail()
	}
	c.SetJSONMarshalStrategy(JSONMarshalSecure)
}

func TestToJson(t *testing.T) {
	c := New()

	if err := c.ReadConfig(bytes.NewReader(yamlConfig1)); err != nil {
		t.Fatal(err)
	}

	c.AwsAccessKey = "MyAwsAccessKey"
	c.AwsSecretKey = "MyAwsSecretKey"
	var err error
	var strJSON string
	if strJSON, err = c.ToJSON(); err != nil {
		t.Fatal(err)
	}

	t.Log(strJSON)

	map1 := map[string]interface{}{}
	map2 := map[string]interface{}{}
	json.Unmarshal([]byte(strJSON), map1)
	json.Unmarshal([]byte(jsonConfig), map2)

	if !reflect.DeepEqual(map1, map2) {
		t.Fail()
	}
}

func TestToSecureJson(t *testing.T) {
	c := New()

	if err := c.ReadConfig(bytes.NewReader(yamlConfig2)); err != nil {
		t.Fatal(err)
	}

	c.AwsAccessKey = "MyAwsAccessKey"
	c.AwsSecretKey = "MyAwsSecretKey"
	var err error
	var strJSON string
	if strJSON, err = c.ToSecureJSON(); err != nil {
		t.Fatal(err)
	}

	t.Log(strJSON)

	map1 := map[string]interface{}{}
	map2 := map[string]interface{}{}
	json.Unmarshal([]byte(strJSON), map1)
	json.Unmarshal([]byte(secureJSONConfig), map2)

	if !reflect.DeepEqual(map1, map2) {
		t.Fail()
	}
}

func TestMarshalToJson(t *testing.T) {
	c := New()

	if err := c.ReadConfig(bytes.NewReader(yamlConfig1)); err != nil {
		t.Fatal(err)
	}

	c.AwsAccessKey = "MyAwsAccessKey"
	c.AwsSecretKey = "MyAwsSecretKey"
	c.SetJSONMarshalStrategy(JSONMarshalPlainText)

	var err error
	var buff []byte
	if buff, err = json.MarshalIndent(c, "", "  "); err != nil {
		t.Fatal(err)
	}

	strJSON := string(buff)

	t.Log(strJSON)

	map1 := map[string]interface{}{}
	map2 := map[string]interface{}{}
	json.Unmarshal([]byte(strJSON), map1)
	json.Unmarshal([]byte(jsonConfig), map2)

	if !reflect.DeepEqual(map1, map2) {
		t.Fail()
	}
}

func TestMarshalToSecureJson(t *testing.T) {
	c := New()

	if err := c.ReadConfig(bytes.NewReader(yamlConfig2)); err != nil {
		t.Fatal(err)
	}

	c.AwsAccessKey = "MyAwsAccessKey"
	c.AwsSecretKey = "MyAwsSecretKey"
	var err error
	var buff []byte
	if buff, err = json.MarshalIndent(c, "", "  "); err != nil {
		t.Fatal(err)
	}

	strJSON := string(buff)

	t.Log(strJSON)

	map1 := map[string]interface{}{}
	map2 := map[string]interface{}{}
	json.Unmarshal([]byte(strJSON), map1)
	json.Unmarshal([]byte(secureJSONConfig), map2)

	if !reflect.DeepEqual(map1, map2) {
		t.Fail()
	}
}

func TestFromJson(t *testing.T) {
	c, err := FromJSON(jsonConfig)
	if err != nil {
		t.Fatal(err)
	}
	assertLogLevel(t, c, "error")
	assertStorageDrivers(t, c)
	assertOsDrivers1(t, c)
	assertAwsSecretKey(t, c, "MyAwsSecretKey")
}

func TestFromJsonWithErrors(t *testing.T) {
	_, err := FromJSON("///*")
	if err == nil {
		t.Fatal("expected unmarshalling error")
	}
}

func TestFromSecureJson(t *testing.T) {
	c, err := FromJSON(secureJSONConfig)
	if err != nil {
		t.Fatal(err)
	}
	assertLogLevel(t, c, "debug")
	assertStorageDrivers(t, c)
	assertOsDrivers2(t, c)
	assertAwsSecretKey(t, c, "")
}

func TestNewWithUserConfigFile(t *testing.T) {
	util.WriteStringToFile(string(yamlConfig1), usrRexRayFile)
	defer os.RemoveAll(usrRexRayFile)

	c := New()

	assertLogLevel(t, c, "error")
	assertStorageDrivers(t, c)
	assertOsDrivers1(t, c)

	if err := c.ReadConfig(bytes.NewReader(yamlConfig2)); err != nil {
		t.Fatal(err)
	}

	assertLogLevel(t, c, "debug")
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

	assertLogLevel(t, c, "error")
	assertStorageDrivers(t, c)
	assertOsDrivers1(t, c)

	if err := c.ReadConfig(bytes.NewReader(yamlConfig2)); err != nil {
		t.Fatal(err)
	}

	assertLogLevel(t, c, "debug")
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

	assertLogLevel(t, c, "error")
	assertStorageDrivers(t, c)
	assertOsDrivers1(t, c)

	if err := c.ReadConfig(bytes.NewReader(yamlConfig2)); err != nil {
		t.Fatal(err)
	}

	assertLogLevel(t, c, "debug")
	assertStorageDrivers(t, c)
	assertOsDrivers2(t, c)
}

func TestReadNilConfig(t *testing.T) {
	if err := New().ReadConfig(nil); err == nil {
		t.Fatal("expected nil config error")
	}
}

func assertAwsSecretKey(t *testing.T, c *Config, expected string) {
	val := c.Viper.GetString("awsSecretKey")
	if val != expected {
		t.Fatalf("viper.awsSecretKey != %s; == %v", expected, val)
	}
	if c.AwsSecretKey != expected {
		t.Fatalf("config.awsSecretKey != %s; == %v", expected, c.AwsSecretKey)
	}
}

func assertLogLevel(t *testing.T, c *Config, expected string) {
	val := c.Viper.GetString("logLevel")
	if val != expected {
		t.Fatalf("viper.logLevel != %s; == %v", expected, val)
	}
	if c.LogLevel != expected {
		t.Fatalf("config.logLevel != %s; == %v", expected, c.LogLevel)
	}
}

func assertStorageDrivers(t *testing.T, c *Config) {
	sd := c.Viper.GetStringSlice("storageDrivers")
	if sd == nil {
		t.Fatalf("viper.storageDrivers == nil")
	}
	if c.StorageDrivers == nil {
		t.Fatalf("config.storageDrivers == nil")
	}

	if len(sd) != 2 {
		t.Fatalf("len(viper.storageDrivers) != 2; == %d", len(sd))
	}
	if len(c.StorageDrivers) != 2 {
		t.Fatalf("len(config.storageDrivers) != 2; == %d", len(c.StorageDrivers))
	}

	if sd[0] != "ec2" {
		t.Fatalf("viper.sd[0] != ec2; == %v", sd[0])
	}
	if c.StorageDrivers[0] != "ec2" {
		t.Fatalf("config.sd[0] != ec2; == %v", c.StorageDrivers[0])
	}

	if sd[1] != "xtremio" {
		t.Fatalf("viper.sd[1] != xtremio; == %v", sd[1])
	}
	if c.StorageDrivers[1] != "xtremio" {
		t.Fatalf("config.sd[1] != xtremio; == %v", c.StorageDrivers[1])
	}
}

func assertOsDrivers1(t *testing.T, c *Config) {
	od := c.Viper.GetStringSlice("osDrivers")
	if od == nil {
		t.Fatalf("viper.osDrivers == nil")
	}
	if c.OSDrivers == nil {
		t.Fatalf("config.osDrivers == nil")
	}

	if len(od) != 1 {
		t.Fatalf("len(viper.osDrivers) != 1; == %d", len(od))
	}
	if len(c.OSDrivers) != 1 {
		t.Fatalf("len(config.osDrivers) != 1; == %d", len(c.OSDrivers))
	}

	if od[0] != "linux" {
		t.Fatalf("viper.od[0] != linux; == %v", od[0])
	}
	if c.OSDrivers[0] != "linux" {
		t.Fatalf("config.od[0] != linux; == %v", c.OSDrivers[0])
	}
}

func assertOsDrivers2(t *testing.T, c *Config) {
	od := c.Viper.GetStringSlice("osDrivers")
	if od == nil {
		t.Fatalf("viper.osDrivers == nil")
	}
	if c.OSDrivers == nil {
		t.Fatalf("config.osDrivers == nil")
	}

	if len(od) != 2 {
		t.Fatalf("len(viper.osDrivers) != 2; == %d", len(od))
	}
	if len(c.OSDrivers) != 2 {
		t.Fatalf("len(config.osDrivers) != 2; == %d", len(c.OSDrivers))
	}

	if od[0] != "darwin" {
		t.Fatalf("viper.od[0] != darwin; == %v", od[0])
	}
	if c.OSDrivers[0] != "darwin" {
		t.Fatalf("config.od[0] != darwin; == %v", c.OSDrivers[0])
	}

	if od[1] != "linux" {
		t.Fatalf("viper.od[1] != linux; == %v", od[1])
	}
	if c.OSDrivers[1] != "linux" {
		t.Fatalf("config.od[1] != linux; == %v", c.OSDrivers[1])
	}
}

var yamlConfig1 = []byte(`logLevel: error
storageDrivers:
- ec2
- xtremio
osDrivers:
- linux`)

var yamlConfig2 = []byte(`logLevel: debug
osDrivers:
- darwin
- linux`)

var jsonConfig = `{
    "LogLevel": "error",
    "StorageDrivers": [
        "ec2",
        "xtremio"
    ],
    "VolumeDrivers": [
        "docker"
    ],
    "OsDrivers": [
        "linux"
    ],
    "MinVolSize": 0,
    "RemoteManagement": false,
    "DockerVolumeType": "",
    "DockerIops": 0,
    "DockerSize": 0,
    "DockerAvailabilityZone": "",
    "AwsAccessKey": "MyAwsAccessKey",
    "AwsRegion": "",
    "RackspaceAuthUrl": "",
    "RackspaceUserId": "",
    "RackspaceUserName": "",
    "RackspaceTenantId": "",
    "RackspaceTenantName": "",
    "RackspaceDomainId": "",
    "RackspaceDomainName": "",
	"OpenstackAuthUrl": "",
    "OpenstackUserId": "",
    "OpenstackUserName": "",
    "OpenstackTenantId": "",
    "OpenstackTenantName": "",
    "OpenstackDomainId": "",
    "OpenstackDomainName": "",
	"OpenstackRegionName": "",
    "ScaleIoEndpoint": "",
    "ScaleIoInsecure": false,
    "ScaleIoUseCerts": true,
    "ScaleIoUserName": "",
    "ScaleIoSystemId": "",
    "ScaleIoSystemName": "",
    "ScaleIoProtectionDomainId": "",
    "ScaleIoProtectionDomainName": "",
    "ScaleIoStoragePoolId": "",
    "ScaleIoStoragePoolName": "",
    "XtremIoEndpoint": "",
    "XtremIoUserName": "",
    "XtremIoInsecure": false,
    "XtremIoDeviceMapper": false,
    "XtremIoMultipath": false,
    "XtremIoRemoteManagement": false,
    "AwsSecretKey": "MyAwsSecretKey",
    "RackspacePassword": "",
    "ScaleIoPassword": "",
    "XtremIoPassword": ""
}`

var secureJSONConfig = `{
    "LogLevel": "debug",
    "StorageDrivers": [
        "ec2",
        "xtremio"
    ],
    "VolumeDrivers": [
        "docker"
    ],
    "OsDrivers": [
        "darwin",
        "linux"
    ],
    "MinVolSize": 0,
    "RemoteManagement": false,
    "DockerVolumeType": "",
    "DockerIops": 0,
    "DockerSize": 0,
    "DockerAvailabilityZone": "",
    "AwsAccessKey": "MyAwsAccessKey",
    "AwsRegion": "",
    "RackspaceAuthUrl": "",
    "RackspaceUserId": "",
    "RackspaceUserName": "",
    "RackspaceTenantId": "",
    "RackspaceTenantName": "",
    "RackspaceDomainId": "",
    "RackspaceDomainName": "",
	"OpenstackAuthUrl": "",
	"OpenstackUserId": "",
	"OpenstackUserName": "",
	"OpenstackTenantId": "",
	"OpenstackTenantName": "",
	"OpenstackDomainId": "",
	"OpenstackDomainName": "",
	"OpenstackRegionName": "",
    "ScaleIoEndpoint": "",
    "ScaleIoInsecure": false,
    "ScaleIoUseCerts": true,
    "ScaleIoUserName": "",
    "ScaleIoSystemId": "",
    "ScaleIoSystemName": "",
    "ScaleIoProtectionDomainId": "",
    "ScaleIoProtectionDomainName": "",
    "ScaleIoStoragePoolId": "",
    "ScaleIoStoragePoolName": "",
    "XtremIoEndpoint": "",
    "XtremIoUserName": "",
    "XtremIoInsecure": false,
    "XtremIoDeviceMapper": false,
    "XtremIoMultipath": false,
    "XtremIoRemoteManagement": false
}`
