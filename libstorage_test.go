package libstorage

import (
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/emccode/libstorage/client"
	"github.com/emccode/libstorage/driver"
	"github.com/emccode/libstorage/model"
)

const (
	mockDriverName = "mockDriver"
)

var (
	nextDeviceVals = []string{"/dev/mock0", "/dev/mock1", "/dev/mock2"}
)

func newDriver(config *gofig.Config) driver.Driver {
	md := &mockDriver{name: mockDriverName}
	var d driver.Driver = md
	return d
}

func getConfig(host string) *gofig.Config {
	if host == "" {
		host = "tcp://127.0.0.1:0"
	}
	config := gofig.New()
	config.Set("libstorage.host", host)
	config.Set("libstorage.drivers", []string{mockDriverName})
	config.Set("libstorage.profiles.enabled", true)
	config.Set("libstorage.profiles.groups", []string{"local=127.0.0.1"})
	return config
}

func mustGetClient(config *gofig.Config, t *testing.T) client.Client {
	if config == nil {
		config = getConfig("")
	}
	if err := Serve(config); err != nil {
		t.Fatalf("error serving libStorage service %v", err)
	}
	c, err := Dial(config)
	if err != nil {
		t.Fatalf("error dialing libStorage service at '%s' %v",
			config.Get("libstorage.host"), err)
	}
	if _, _, err := c.InitDrivers(); err != nil {
		t.Fatalf("error initializing libStorage drivers %v", err)
	}
	return c
}

func TestMain(m *testing.M) {
	if os.Getenv("LIBSTORAGE_DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	RegisterDriver(mockDriverName, newDriver)
	os.Exit(m.Run())
}

func TestGetRegisteredDriverNames(t *testing.T) {
	c := mustGetClient(nil, t)
	driverNames := c.GetRegisteredDriverNames()
	assert.Equal(t, 1, len(driverNames))
	assert.Equal(t, mockDriverName, driverNames[0])
}

func TestGetInitializedDriverNames(t *testing.T) {
	c := mustGetClient(nil, t)
	driverNames := c.GetInitializedDriverNames()
	assert.Equal(t, 1, len(driverNames))
	assert.Equal(t, mockDriverName, driverNames[0])
}

func TestGetVolumeMapping(t *testing.T) {
	c := mustGetClient(nil, t)
	bds, err := c.GetVolumeMapping()
	if err != nil {
		t.Fatal(err)
	}
	if len(bds) != 1 {
		t.Fatalf("len(blockDevices) != 1; == %d", len(bds))
	}
}

type mockDriver struct {
	config          *gofig.Config
	name            string
	nextDeviceIndex int
}

func (m *mockDriver) Init() error {
	return nil
}

func (m *mockDriver) Name() string {
	return m.name
}

func (m *mockDriver) GetVolumeMapping(
	ctx context.Context) ([]*model.BlockDevice, error) {
	return []*model.BlockDevice{&model.BlockDevice{
		DeviceName:   "test",
		ProviderName: m.name,
		InstanceID:   &model.InstanceID{ID: "mockDriverInstanceID"},
		Region:       "test",
	}}, nil
}

func (m *mockDriver) GetInstanceID() (*model.InstanceID, error) {
	return &model.InstanceID{ID: "mockDriverInstanceID"}, nil
}

func (m *mockDriver) GetInstance(
	ctx context.Context) (*model.Instance, error) {
	return &model.Instance{
		Name:         "test",
		InstanceID:   &model.InstanceID{ID: "mockDriverInstanceID"},
		ProviderName: m.name,
		Region:       "test"}, nil
}

func (m *mockDriver) GetVolume(
	ctx context.Context,
	volumeID, volumeName string) ([]*model.Volume, error) {
	return []*model.Volume{&model.Volume{
		Name:             volumeName,
		VolumeID:         "test",
		AvailabilityZone: "test",
	}}, nil
}

func (m *mockDriver) GetVolumeAttach(
	ctx context.Context,
	volumeID string) ([]*model.VolumeAttachment, error) {
	return []*model.VolumeAttachment{&model.VolumeAttachment{
		VolumeID: "test",
	}}, nil
}

func (m *mockDriver) CreateSnapshot(
	ctx context.Context,
	snapshotName,
	volumeID,
	description string) ([]*model.Snapshot, error) {
	return []*model.Snapshot{&model.Snapshot{
		VolumeID: "test",
	}}, nil
}

func (m *mockDriver) GetSnapshot(
	ctx context.Context,
	volumeID, snapshotID, snapshotName string) ([]*model.Snapshot, error) {
	return []*model.Snapshot{&model.Snapshot{
		VolumeID: "test",
	}}, nil
}

func (m *mockDriver) RemoveSnapshot(
	ctx context.Context, snapshotID string) error {
	return nil
}

func (m *mockDriver) CreateVolume(
	ctx context.Context,
	volumeName,
	volumeID,
	snapshotID,
	volumeType string,
	IOPS,
	size int64,
	availabilityZone string) (*model.Volume, error) {
	return &model.Volume{
		Name:             volumeName,
		VolumeID:         "test",
		AvailabilityZone: "test",
	}, nil
}

func (m *mockDriver) RemoveVolume(
	ctx context.Context,
	volumeID string) error {
	return nil
}

func (m *mockDriver) GetDeviceNextAvailable() (string, error) {
	next := nextDeviceVals[m.nextDeviceIndex]
	if m.nextDeviceIndex == 2 {
		m.nextDeviceIndex = 0
	} else {
		m.nextDeviceIndex++
	}
	return next, nil
}

func (m *mockDriver) AttachVolume(
	ctx context.Context,
	nextDeviceName,
	volumeID string) ([]*model.VolumeAttachment, error) {
	return []*model.VolumeAttachment{&model.VolumeAttachment{
		VolumeID: "test",
	}}, nil
}

func (m *mockDriver) DetachVolume(
	ctx context.Context,
	volumeID string) error {
	return nil
}

func (m *mockDriver) CopySnapshot(
	ctx context.Context,
	volumeID,
	snapshotID,
	snapshotName,
	destinationSnapshotName,
	destinationRegion string) (*model.Snapshot, error) {
	return &model.Snapshot{
		VolumeID: "test",
	}, nil
}

func (m *mockDriver) GetClientToolName(ctx context.Context) (string, error) {
	return "mockDriver.sh", nil
}

func (m *mockDriver) GetClientTool(ctx context.Context) ([]byte, error) {
	return nil, nil
}
