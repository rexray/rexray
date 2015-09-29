package config

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/emccode/rexray/util"
)

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

func assertLogLevel(t *testing.T, c *Config, expected string) {
	ll := c.Viper.GetString("logLevel")
	if ll != expected {
		t.Fatalf("viper.logLevel != %s; == %v", expected, ll)
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
	if c.OsDrivers == nil {
		t.Fatalf("config.osDrivers == nil")
	}

	if len(od) != 1 {
		t.Fatalf("len(viper.osDrivers) != 1; == %d", len(od))
	}
	if len(c.OsDrivers) != 1 {
		t.Fatalf("len(config.osDrivers) != 1; == %d", len(c.OsDrivers))
	}

	if od[0] != "linux" {
		t.Fatalf("viper.od[0] != linux; == %v", od[0])
	}
	if c.OsDrivers[0] != "linux" {
		t.Fatalf("config.od[0] != linux; == %v", c.OsDrivers[0])
	}
}

func assertOsDrivers2(t *testing.T, c *Config) {
	od := c.Viper.GetStringSlice("osDrivers")
	if od == nil {
		t.Fatalf("viper.osDrivers == nil")
	}
	if c.OsDrivers == nil {
		t.Fatalf("config.osDrivers == nil")
	}

	if len(od) != 2 {
		t.Fatalf("len(viper.osDrivers) != 2; == %d", len(od))
	}
	if len(c.OsDrivers) != 2 {
		t.Fatalf("len(config.osDrivers) != 2; == %d", len(c.OsDrivers))
	}

	if od[0] != "darwin" {
		t.Fatalf("viper.od[0] != darwin; == %v", od[0])
	}
	if c.OsDrivers[0] != "darwin" {
		t.Fatalf("config.od[0] != darwin; == %v", c.OsDrivers[0])
	}

	if od[1] != "linux" {
		t.Fatalf("viper.od[1] != linux; == %v", od[1])
	}
	if c.OsDrivers[1] != "linux" {
		t.Fatalf("config.od[1] != linux; == %v", c.OsDrivers[1])
	}
}
