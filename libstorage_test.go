package libstorage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"
	gocontext "golang.org/x/net/context"

	"github.com/emccode/libstorage/api"
	"github.com/emccode/libstorage/client"
	"github.com/emccode/libstorage/context"
	"github.com/emccode/libstorage/driver"
)

const (
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
  service:
    http:
      logging:
        enabled: false
        logrequest: false
        logresponse: false
    servers:
      %s:
        libstorage:
          drivers:
          - %s
      %s:
        libstorage:
          drivers:
          - %s
      %s:
        libstorage:
          drivers:
          - %s
`,
		host, server,
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
	t *testing.T) (gocontext.Context, client.Client) {
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
	if os.Getenv("LIBSTORAGE_DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	RegisterDriver(mockDriver1Name, newMockDriver1)
	RegisterDriver(mockDriver2Name, newMockDriver2)
	RegisterDriver(mockDriver3Name, newMockDriver3)
	os.Exit(m.Run())
}

func TestGetRegisteredDriverNames(t *testing.T) {
	testGetRegisteredDriverNames(testServer1Name, t)
	testGetRegisteredDriverNames(testServer2Name, t)
	testGetRegisteredDriverNames(testServer3Name, t)
}

func testGetRegisteredDriverNames(server string, t *testing.T) {
	config := getConfig("", server, t)
	ctx, c := mustGetClient(config, t)
	args := &api.GetDriverNamesArgs{}
	driverNames, err := c.GetRegisteredDriverNames(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(driverNames))
	assert.Contains(t, driverNames, mockDriver1Name)
	assert.Contains(t, driverNames, mockDriver2Name)
	assert.Contains(t, driverNames, mockDriver3Name)
}

func TestGetInitializedDriverNamesServerAndDriver1(t *testing.T) {
	testGetInitializedDriverNames(testServer1Name, mockDriver1Name, t)
}

func TestGetInitializedDriverNamesServerAndDriver2(t *testing.T) {
	testGetInitializedDriverNames(testServer2Name, mockDriver2Name, t)
}

func TestGetInitializedDriverNamesServerAndDriver3(t *testing.T) {
	testGetInitializedDriverNames(testServer3Name, mockDriver3Name, t)
}

func testGetInitializedDriverNames(server, driver string, t *testing.T) {
	config := getConfig("", server, t)
	ctx, c := mustGetClient(config, t)
	args := &api.GetDriverNamesArgs{}
	driverNames, err := c.GetInitializedDriverNames(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(driverNames))
	assert.Contains(t, driverNames, driver)
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
	assert.Equal(t, driver+"Device", bd.DeviceName)
	assert.Equal(t, driver+"Provider", bd.ProviderName)
	assert.Equal(t, driver+"Region", bd.Region)
	assertInstanceID(t, driver, bd.InstanceID)
}

func TestGetNextAvailableDeviceNameServerAndDriver1(t *testing.T) {
	testGetNextAvailableDeviceName(testServer1Name, mockDriver1Name, t)
}
func testGetNextAvailableDeviceName(server, driver string, t *testing.T) {
	config := getConfig("", server, t)
	ctx, c := mustGetClient(config, t)
	name, err := c.GetNextAvailableDeviceName(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("nextAvailableDeviceName=%s", name)
}

func TestGetClientToolName(t *testing.T) {
	ctx, c := mustGetClient(nil, t)
	args := &api.GetClientToolNameArgs{}
	_, err := c.GetClientToolName(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
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

func (m *mockDriver) GetVolumeMapping(
	ctx context.Context,
	args *api.GetVolumeMappingArgs) ([]*api.BlockDevice, error) {
	return []*api.BlockDevice{&api.BlockDevice{
		DeviceName:   m.pwn("Device"),
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

func (m *mockDriver) GetClientToolName(
	ctx context.Context,
	args *api.GetClientToolNameArgs) (string, error) {
	return m.pwn("-clientTool.sh"), nil
}

func (m *mockDriver) GetClientTool(
	ctx context.Context,
	args *api.GetClientToolArgs) ([]byte, error) {

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
	return []byte(script), nil
}
