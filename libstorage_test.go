package libstorage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/akutz/gotil"
	"github.com/stretchr/testify/assert"

	"github.com/emccode/libstorage/api"
	"github.com/emccode/libstorage/client"
	"github.com/emccode/libstorage/context"
	"github.com/emccode/libstorage/driver"
)

const (
	localDevicesFile = "/tmp/libstorage/partitions"
	clientToolDir    = "/tmp/libstorage/bin"

	testServer1Name = "testServer1"
	testServer2Name = "testServer2"
	testServer3Name = "testServer3"

	mockDriver1Name = "mockDriver1"
	mockDriver2Name = "mockDriver2"
	mockDriver3Name = "mockDriver3"
)

var (
	nextDeviceVals = []string{"/dev/mock0", "/dev/mock1", "/dev/mock2"}
)

func newMockDriver1(config gofig.Config) driver.Driver {
	md := &mockDriver{name: mockDriver1Name}
	var d driver.Driver = md
	return d
}

func newMockDriver2(config gofig.Config) driver.Driver {
	md := &mockDriver{name: mockDriver2Name}
	var d driver.Driver = md
	return d
}

func newMockDriver3(config gofig.Config) driver.Driver {
	md := &mockDriver{name: mockDriver3Name}
	var d driver.Driver = md
	return d
}

func getConfig(host, server string, t *testing.T) gofig.Config {
	if host == "" {
		host = "tcp://127.0.0.1:0"
	}
	if server == "" {
		server = "testServer2"
	}
	config := gofig.New()
	configYaml := fmt.Sprintf(`
libstorage:
  host: %s
  server: %s
  profiles:
    enabled: true
    groups:
    - local=127.0.0.1
  client:
    tooldir: %s
    localdevicesfile: %s
    http:
      logging:
        enabled: false
        logrequest: false
        logresponse: false
  service:
    http:
      logging:
        enabled: false
        logrequest: false
        logresponse: false
    servers:
      %s:
        libstorage:
          driver: %s
          profiles:
            enabled: true
            groups:
            - remote=127.0.0.1
      %s:
        libstorage:
          driver: %s
      %s:
        libstorage:
          driver: %s
`,
		host, server,
		clientToolDir, localDevicesFile,
		testServer1Name, mockDriver1Name,
		testServer2Name, mockDriver2Name,
		testServer3Name, mockDriver3Name)

	t.Log(configYaml)
	configYamlBuf := []byte(configYaml)
	if err := config.ReadConfig(bytes.NewReader(configYamlBuf)); err != nil {
		panic(err)
	}
	return config
}

func mustGetClient(
	config gofig.Config,
	t *testing.T) (context.Context, client.Client) {
	if config == nil {
		config = getConfig("", "", t)
	}
	if err := Serve(config); err != nil {
		t.Fatalf("error serving libStorage service %v", err)
	}
	ctx := context.Background()
	c, err := Dial(ctx, config)
	if err != nil {
		t.Fatalf("error dialing libStorage service at '%s' %v",
			config.Get("libstorage.host"), err)
	}
	return ctx, c
}

func TestMain(m *testing.M) {
	os.MkdirAll(clientToolDir, 0755)
	ioutil.WriteFile(localDevicesFile, localDevicesFileBuf, 0644)

	if os.Getenv("LIBSTORAGE_DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
		os.Setenv("LIBSTORAGE_SERVICE_HTTP_LOGGING_ENABLED", "true")
		os.Setenv("LIBSTORAGE_SERVICE_HTTP_LOGGING_LOGREQUEST", "true")
		os.Setenv("LIBSTORAGE_SERVICE_HTTP_LOGGING_LOGRESPONSE", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_ENABLED", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_LOGREQUEST", "true")
		os.Setenv("LIBSTORAGE_CLIENT_HTTP_LOGGING_LOGRESPONSE", "true")
	}
	RegisterDriver(mockDriver1Name, newMockDriver1)
	RegisterDriver(mockDriver2Name, newMockDriver2)
	RegisterDriver(mockDriver3Name, newMockDriver3)

	exitCode := m.Run()

	os.RemoveAll(clientToolDir)
	os.RemoveAll(localDevicesFile)
	os.Exit(exitCode)
}

func TestGetServiceInfo(t *testing.T) {
	testGetServiceInfo(testServer1Name, mockDriver1Name, t)
	testGetServiceInfo(testServer2Name, mockDriver2Name, t)
	testGetServiceInfo(testServer3Name, mockDriver3Name, t)
}

func testGetServiceInfo(server, driver string, t *testing.T) {
	config := getConfig("", server, t)
	ctx, c := mustGetClient(config, t)
	args := &api.GetServiceInfoArgs{}
	info, err := c.GetServiceInfo(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, server, info.Name)
	assert.Equal(t, driver, info.Driver)
	assert.Equal(t, 3, len(info.RegisteredDrivers))
	assert.True(t, gotil.StringInSlice(mockDriver1Name, info.RegisteredDrivers))
	assert.True(t, gotil.StringInSlice(mockDriver2Name, info.RegisteredDrivers))
	assert.True(t, gotil.StringInSlice(mockDriver3Name, info.RegisteredDrivers))
}

func TestGetVolumeMappingServerAndDriver1(t *testing.T) {
	testGetVolumeMapping(testServer1Name, mockDriver1Name, t)
}

func TestGetVolumeMappingServerAndDriver2(t *testing.T) {
	testGetVolumeMapping(testServer2Name, mockDriver2Name, t)
}

func TestGetVolumeMappingServerAndDriver3(t *testing.T) {
	testGetVolumeMapping(testServer3Name, mockDriver3Name, t)
}

func testGetVolumeMapping(server, driver string, t *testing.T) {
	config := getConfig("", server, t)
	ctx, c := mustGetClient(config, t)
	args := &api.GetVolumeMappingArgs{}
	bds, err := c.GetVolumeMapping(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(bds))
	bd := bds[0]
	assert.Equal(t, getDeviceName(driver), bd.DeviceName)
	assert.Equal(t, driver+"Provider", bd.ProviderName)
	assert.Equal(t, driver+"Region", bd.Region)
	assertInstanceID(t, driver, bd.InstanceID)
}

func TestGetNextAvailableDeviceNameServerAndDriver1(t *testing.T) {
	testGetNextAvailableDeviceName(testServer1Name, mockDriver1Name, t)
}

func TestGetNextAvailableDeviceNameServerAndDriver2(t *testing.T) {
	testGetNextAvailableDeviceName(testServer2Name, mockDriver2Name, t)
}

func TestGetNextAvailableDeviceNameServerAndDriver3(t *testing.T) {
	testGetNextAvailableDeviceName(testServer3Name, mockDriver3Name, t)
}

func testGetNextAvailableDeviceName(server, driver string, t *testing.T) {
	config := getConfig("", server, t)
	ctx, c := mustGetClient(config, t)
	name, err := c.GetNextAvailableDeviceName(
		ctx, &api.GetNextAvailableDeviceNameArgs{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, name, getNextDeviceName(driver))
}

func TestGetClientTool(t *testing.T) {
	ctx, c := mustGetClient(nil, t)
	args := &api.GetClientToolArgs{}
	_, err := c.GetClientTool(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
}

func assertInstanceID(t *testing.T, driver string, iid *api.InstanceID) {
	assert.NotNil(t, iid)
	assert.Equal(t, driver+"InstanceID", iid.ID)
	md := iid.Metadata
	assert.NotNil(t, md)
	tmd, ok := md.(map[string]interface{})
	assert.True(t, ok, "metadata is map")
	emd := mockDriverInstanceIDMetadata()
	assert.EqualValues(t, emd["min"], tmd["min"])
	assert.EqualValues(t, emd["max"], tmd["max"])
	assert.EqualValues(t, emd["rad"], tmd["rad"])
	assert.EqualValues(t, emd["totally"], tmd["totally"])
}

type mockDriver struct {
	config gofig.Config
	name   string
}

func (m *mockDriver) pwn(v string) string {
	return fmt.Sprintf("%s%s", m.name, v)
}

func mockDriverInstanceIDMetadata() map[string]interface{} {
	return map[string]interface{}{
		"min":     0,
		"max":     10,
		"rad":     "cool",
		"totally": "tubular",
	}
}

func (m *mockDriver) iid() *api.InstanceID {
	return &api.InstanceID{
		ID:       m.pwn("InstanceID"),
		Metadata: mockDriverInstanceIDMetadata(),
	}
}

func (m *mockDriver) Init() error {
	return nil
}

func (m *mockDriver) Name() string {
	return m.name
}

func (m *mockDriver) GetNextAvailableDeviceName(
	ctx context.Context,
	args *api.GetNextAvailableDeviceNameArgs) (
	*api.NextAvailableDeviceName, error) {
	return &api.NextAvailableDeviceName{
		Prefix:  "xvd",
		Pattern: `\w`,
		Ignore:  getDeviceIgnore(m.name),
	}, nil
}

func (m *mockDriver) GetVolumeMapping(
	ctx context.Context,
	args *api.GetVolumeMappingArgs) ([]*api.BlockDevice, error) {
	return []*api.BlockDevice{&api.BlockDevice{
		DeviceName:   getDeviceName(m.name),
		ProviderName: m.pwn("Provider"),
		InstanceID:   m.iid(),
		Region:       m.pwn("Region"),
	}}, nil
}

func (m *mockDriver) GetInstance(
	ctx context.Context,
	args *api.GetInstanceArgs) (*api.Instance, error) {
	return &api.Instance{
		Name:         m.pwn("Name"),
		InstanceID:   m.iid(),
		ProviderName: m.pwn("Provider"),
		Region:       m.pwn("Region"),
	}, nil
}

func (m *mockDriver) GetVolume(
	ctx context.Context,
	args *api.GetVolumeArgs) ([]*api.Volume, error) {
	return []*api.Volume{&api.Volume{
		Name:             m.pwn(args.Optional.VolumeName),
		VolumeID:         m.pwn("VolumeID"),
		AvailabilityZone: m.pwn("AvailabilityZone"),
	}}, nil
}

func (m *mockDriver) GetVolumeAttach(
	ctx context.Context,
	args *api.GetVolumeAttachArgs) ([]*api.VolumeAttachment, error) {
	return []*api.VolumeAttachment{&api.VolumeAttachment{
		VolumeID: m.pwn("VolumeID"),
	}}, nil
}

func (m *mockDriver) CreateSnapshot(
	ctx context.Context,
	args *api.CreateSnapshotArgs) ([]*api.Snapshot, error) {
	return []*api.Snapshot{&api.Snapshot{
		VolumeID: m.pwn("VolumeID"),
	}}, nil
}

func (m *mockDriver) GetSnapshot(
	ctx context.Context,
	args *api.GetSnapshotArgs) ([]*api.Snapshot, error) {
	return []*api.Snapshot{&api.Snapshot{
		VolumeID: m.pwn("VolumeID"),
	}}, nil
}

func (m *mockDriver) RemoveSnapshot(
	ctx context.Context,
	args *api.RemoveSnapshotArgs) error {
	return nil
}

func (m *mockDriver) CreateVolume(
	ctx context.Context,
	args *api.CreateVolumeArgs) (*api.Volume, error) {
	return &api.Volume{
		Name:             m.pwn(args.Optional.VolumeName),
		VolumeID:         m.pwn("VolumeID"),
		AvailabilityZone: m.pwn("AvailabilityZone"),
	}, nil
}

func (m *mockDriver) RemoveVolume(
	ctx context.Context,
	args *api.RemoveVolumeArgs) error {
	return nil
}

func (m *mockDriver) AttachVolume(
	ctx context.Context,
	args *api.AttachVolumeArgs) ([]*api.VolumeAttachment, error) {
	return []*api.VolumeAttachment{&api.VolumeAttachment{
		VolumeID: m.pwn("VolumeID"),
	}}, nil
}

func (m *mockDriver) DetachVolume(
	ctx context.Context,
	args *api.DetachVolumeArgs) error {
	return nil
}

func (m *mockDriver) CopySnapshot(
	ctx context.Context,
	args *api.CopySnapshotArgs) (*api.Snapshot, error) {
	return &api.Snapshot{
		VolumeID: m.pwn("VolumeID"),
	}, nil
}

func (m *mockDriver) GetClientTool(
	ctx context.Context,
	args *api.GetClientToolArgs) (*api.ClientTool, error) {

	clientTool := &api.ClientTool{
		Name: m.pwn("-clientTool.sh"),
	}

	if args.Optional.OmitBinary {
		return clientTool, nil
	}

	jsonBuf, err := json.Marshal(m.iid())
	if err != nil {
		return nil, err
	}
	script := fmt.Sprintf(`#!/bin/sh
	case "$1" in
	"GetInstanceID")
	    echo '%s'
	    ;;
	"GetNextAvailableDeviceName")
	    echo $BLOCK_DEVICES_JSON
	    ;;
	esac
`, string(jsonBuf))

	clientTool.Data = []byte(script)
	return clientTool, nil
}

func getDeviceName(driver string) string {
	var deviceName string
	switch driver {
	case mockDriver1Name:
		deviceName = "/dev/xvdb"
	case mockDriver2Name:
		deviceName = "/dev/xvda"
	case mockDriver3Name:
		deviceName = "/dev/xvdc"
	}
	return deviceName
}

func getNextDeviceName(driver string) string {
	var deviceName string
	switch driver {
	case mockDriver1Name:
		deviceName = "/dev/xvdc"
	case mockDriver3Name:
		deviceName = "/dev/xvdb"
	}
	return deviceName
}

func getDeviceIgnore(driver string) bool {
	if driver == mockDriver2Name {
		return true
	}
	return false
}

var localDevicesFileBuf = []byte(`
major minor  #blocks  name

  11        0    4050944 sr0
   8        0   67108864 sda
   8        1     512000 sda1
   8        2   66595840 sda2
 253        0    4079616 dm-0
 253        1   42004480 dm-1
 253        2   20508672 dm-2
 1024       1   20508672 xvda
   7        0  104857600 loop0
   7        1    2097152 loop1
 253        3  104857600 dm-3
`)
