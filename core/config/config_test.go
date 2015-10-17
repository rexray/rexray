package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/emccode/rexray/util"
)

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

func TestNew(t *testing.T) {

	usrRexRayDir := fmt.Sprintf("%s/.rexray", util.HomeDir())
	os.MkdirAll(usrRexRayDir, 0755)
	usrRexRayFile := fmt.Sprintf("%s/%s.%s", usrRexRayDir, "config", "yml")
	usrRexRayFileBak := fmt.Sprintf("%s.bak", usrRexRayFile)

	os.Remove(usrRexRayFileBak)
	os.Rename(usrRexRayFile, usrRexRayFileBak)
	defer func() {
		os.Remove(usrRexRayFile)
		os.Rename(usrRexRayFileBak, usrRexRayFile)
	}()

	util.WriteStringToFile(string(yamlConfig1), usrRexRayFile)

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
