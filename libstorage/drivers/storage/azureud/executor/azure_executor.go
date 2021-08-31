package executor

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/azureud"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/azureud/utils"
)

// driver is the storage executor for the azureud storage driver.
type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(azureud.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	ctx.Info("azureud_executor: Init")
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return azureud.Name
}

// Supported returns a flag indicating whether or not the platform
// implementing the executor is valid for the host on which the executor
// resides.
func (d *driver) Supported(
	ctx types.Context,
	opts types.Store) (bool, error) {

	if !gotil.FileExistsInPath("lsscsi") {
		ctx.Error("lsscsi executable not found in PATH")
		return false, nil
	}

	return utils.IsAzureInstance(ctx)
}

// InstanceID returns the instance ID from the current instance from metadata
func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {
	return utils.InstanceID(ctx)
}

// NextDevice returns the next available device.
func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {

	return "", types.ErrNotImplemented
}

var (
	devRX  = regexp.MustCompile(`^/dev/sd[c-z]$`)
	scsiRx = regexp.MustCompile(`^\[\d+:\d+:\d+:(\d+)\]$`)
)

// Retrieve device paths currently attached and/or mounted
func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {

	// Read all of the attached devices
	scsiDevs, err := getSCSIDevs()
	if err != nil {
		return nil, err
	}

	devMap := map[string]string{}

	scanner := bufio.NewScanner(bytes.NewReader(scsiDevs))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		device := fields[len(fields)-1]
		if !devRX.MatchString(device) {
			continue
		}

		matches := scsiRx.FindStringSubmatch(fields[0])
		if matches == nil {
			continue
		}

		lun := matches[1]
		devMap[lun] = device
	}

	ld := &types.LocalDevices{Driver: d.Name()}
	if len(devMap) > 0 {
		ld.DeviceMap = devMap
	}

	ctx.WithField("devicemap", ld.DeviceMap).Debug("local devices")

	return ld, nil
}

func getSCSIDevs() ([]byte, error) {

	out, err := exec.Command("lsscsi").Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			stderr := string(exiterr.Stderr)
			log.Errorf("Unable to get scsi devices: %s", stderr)
			return nil,
				goof.Newf("Unable to get scsi devices: %s",
					stderr)
		}
		return nil, goof.WithError("Unable to get scsci devices", err)
	}

	return out, nil
}
