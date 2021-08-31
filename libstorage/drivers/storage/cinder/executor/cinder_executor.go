package executor

import (
	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/goof"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/cinder"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type driver struct {
	config   gofig.Config
	osDriver types.OSDriver
}

func init() {
	registry.RegisterStorageExecutor(cinder.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config

	var err error
	if d.osDriver, err = registry.NewOSDriver(runtime.GOOS); err != nil {
		return err
	}
	if err = d.osDriver.Init(ctx, config); err != nil {
		return err
	}

	return nil
}

func (d *driver) Name() string {
	return cinder.Name
}

func (d *driver) InstanceID(ctx types.Context, opts types.Store) (*types.InstanceID, error) {
	fields := map[string]interface{}{}
	uuid, err := getInstanceIDFromMetadataServer()
	if err != nil {
		fields["metadataServer"] = err
		uuid, err = getInstanceIDFromConfigDrive(ctx, d)
		if err != nil {
			fields["configDrive"] = err
			uuid, err = getInstanceIDWithDMIDecode()
			if err != nil {
				fields["dmidecode"] = err
				return nil, goof.WithFields(fields, "unable to get InstanceID from any sources")
			}
		}
	}

	iid := &types.InstanceID{Driver: cinder.Name, ID: strings.ToLower(uuid)}

	return iid, nil
}

func parseUUID(metadata []byte) (string, error) {
	var decodedJSON interface{}
	err := json.Unmarshal(metadata, &decodedJSON)
	if err != nil {
		return "", goof.WithError("error unmarshalling metadata", err)
	}
	decodedJSONMap, ok := decodedJSON.(map[string]interface{})
	if !ok {
		return "", goof.New("error casting metadata decoded JSON")
	}
	uuid, ok := decodedJSONMap["uuid"].(string)
	if !ok {
		return "", goof.New("error casting metadata uuid field")
	}

	return uuid, nil
}

func execCommand(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	out, err := command.Output()
	if exiterr, ok := err.(*exec.ExitError); ok {
		stderr := string(exiterr.Stderr)
		return "", goof.WithFieldE("stderr", stderr, "execute command failed", err)
	} else if err != nil {
		return "", goof.WithError("execute command failed", err)
	}
	return string(out), nil
}

// the code of getInstanceIDFromConfigDrive is mostly copied from k8s OpenStack driver
// https://github.com/kubernetes/kubernetes/blob/master/pkg/cloudprovider/providers/openstack/metadata.go
// Copyright to the original authors (Apache license)

// Config drive is defined as an iso9660 or vfat (deprecated) drive
// with the "config-2" label.
// http://docs.openstack.org/user-guide/cli-config-drive.html
const configDriveLabel = "config-2"
const configDrivePath = "openstack/latest/meta_data.json"

func getInstanceIDFromConfigDrive(ctx types.Context, d *driver) (string, error) {
	// Try to read instance UUID from config drive.
	dev := "/dev/disk/by-label/" + configDriveLabel
	if _, err := os.Stat(dev); os.IsNotExist(err) {
		cmdOut, err := execCommand(
			"blkid", "-l",
			"-t", "LABEL="+configDriveLabel,
			"-o", "device",
		)

		if err != nil {
			return "", goof.WithError("Unable to run blkid", err)
		}
		dev = strings.TrimSpace(string(cmdOut))
	}

	mntdir, err := ioutil.TempDir("", "configdrive")
	if err != nil {
		return "", err
	}
	defer os.Remove(mntdir)

	mountOpts := types.DeviceMountOpts{
		FsType:       "iso9660",
		MountOptions: "ro",
	}
	err = d.osDriver.Mount(ctx, dev, mntdir, &mountOpts)
	if err != nil {
		mountOpts.FsType = "vfat"
		err = d.osDriver.Mount(ctx, dev, mntdir, &mountOpts)
	}
	if err != nil {
		return "", goof.WithFieldE("device", dev, "error mounting configdrive", err)
	}
	defer d.osDriver.Unmount(ctx, mntdir, utils.NewStore())

	metadataBytes, err := ioutil.ReadFile(filepath.Join(mntdir, configDrivePath))
	if err != nil {
		return "", goof.WithError("error reading metadata file on config drive", err)
	}

	return parseUUID(metadataBytes)
}

func getInstanceIDFromMetadataServer() (string, error) {
	const metadataURL = "http://169.254.169.254/openstack/latest/meta_data.json"
	httpClient := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	resp, err := httpClient.Get(metadataURL)
	if err != nil {
		return "", goof.WithFieldE("url", metadataURL, "error getting metadata from server", err)
	}
	defer resp.Body.Close()

	metadataBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", goof.WithError("io error reading metadata", err)
	}

	return parseUUID(metadataBytes)
}

func getInstanceIDWithDMIDecode() (string, error) {
	cmdOut, err := execCommand("dmidecode", "-t", "system")
	if err != nil {
		return "", goof.WithError("error calling dmidecode", err)
	}

	rp := regexp.MustCompile("UUID:(.*)")
	uuid := strings.Replace(rp.FindString(string(cmdOut)), "UUID: ", "", -1)

	return strings.ToLower(uuid), nil
}

func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", types.ErrNotImplemented
}

func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {
	devicesMap := make(map[string]string)

	file := "/proc/partitions"
	contentBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil,
			goof.WithFieldE("file", file, "error reading file", err)
	}

	content := string(contentBytes)

	lines := strings.Split(content, "\n")
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			devicePath := "/dev/" + fields[3]
			devicesMap[devicePath] = ""
		}
	}

	return &types.LocalDevices{
		Driver:    cinder.Name,
		DeviceMap: devicesMap,
	}, nil
}
