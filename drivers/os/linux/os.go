package linux

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/docker/docker/pkg/mount"
	osdriver "github.com/emccode/rexray/drivers/os"
	"github.com/opencontainers/runc/libcontainer/label"
)

func init() {
	osName := "linux"
	osdriver.Register(osName, Init)
}

var (
	ErrUnknownOS             = errors.New("Encountered an unknown OS")
	ErrUnknownFilesystemType = errors.New("Unknown filesystem type")
)

type Driver struct{}

func Init() (osdriver.Driver, error) {
	if runtime.GOOS == "linux" {
		if os.Getenv("REXRAY_DEBUG") == "true" {
			log.Println("OS Driver Initialized: " + runtime.GOOS)
		}

		return &Driver{}, nil
	} else {
		return nil, ErrUnknownOS
	}

}

func (driver *Driver) GetMounts(deviceName, mountPoint string) ([]*mount.Info, error) {

	mounts, err := mount.GetMounts()
	if err != nil {
		return nil, err
	}

	if mountPoint == "" && deviceName == "" {
		return mounts, nil
	} else if mountPoint != "" && deviceName != "" {
		return nil, errors.New("Cannot specify mountPoint and deviceName")
	}

	var matchedMounts []*mount.Info
	for _, mount := range mounts {
		if mount.Mountpoint == mountPoint || mount.Source == deviceName {
			matchedMounts = append(matchedMounts, mount)
		}
	}
	return matchedMounts, nil
}

func (driver *Driver) Mounted(mountPoint string) (bool, error) {
	return mount.Mounted(mountPoint)
}

func (driver *Driver) Unmount(mountPoint string) error {
	return mount.Unmount(mountPoint)
}

func (driver *Driver) Mount(device, target, mountOptions, mountLabel string) error {

	fsType, err := probeFsType(device)
	if err != nil {
		return err
	}

	options := label.FormatMountLabel("", mountLabel)
	options = fmt.Sprintf("%s,%s", mountOptions, mountLabel)
	if fsType == "xfs" {
		options = fmt.Sprintf("%s,nouuid", mountOptions)
	}

	if err := mount.Mount(device, target, fsType, options); err != nil {
		return fmt.Errorf("Couldn't mount directory %s at %s: %s", device, target, err)
	}

	return nil
}

// Format will look for ext4/xfs and overwrite it is it doesn't exist
func (driver *Driver) Format(deviceName, newFsType string, overwriteFs bool) error {

	var fsDetected bool

	fsType, err := probeFsType(deviceName)
	if err != nil {
		if err != ErrUnknownFilesystemType {
			return err
		}
	}
	if fsType != "" {
		fsDetected = true
	}

	if overwriteFs || !fsDetected {
		switch newFsType {
		case "ext4":
			if err := exec.Command("mkfs.ext4", deviceName).Run(); err != nil {
				return fmt.Errorf("Problem creating filesystem on %s with error %s", deviceName, err)
			}
		case "xfs":
			if err := exec.Command("mkfs.xfs", "-f", deviceName).Run(); err != nil {
				return fmt.Errorf("Problem creating filesystem on %s with error %s", deviceName, err)
			}
		default:
			return errors.New("Unsupported FS")
		}
	}

	return nil
}

// from github.com/docker/docker/daemon/graphdriver/devmapper/
// this should be abstracted outside of graphdriver but within Docker package, here temporarily
type probeData struct {
	fsName string
	magic  string
	offset uint64
}

func probeFsType(device string) (string, error) {
	probes := []probeData{
		{"btrfs", "_BHRfS_M", 0x10040},
		{"ext4", "\123\357", 0x438},
		{"xfs", "XFSB", 0},
	}

	maxLen := uint64(0)
	for _, p := range probes {
		l := p.offset + uint64(len(p.magic))
		if l > maxLen {
			maxLen = l
		}
	}

	file, err := os.Open(device)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, maxLen)
	l, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	if uint64(l) != maxLen {
		return "", fmt.Errorf("unable to detect filesystem type of %s, short read", device)
	}

	for _, p := range probes {
		if bytes.Equal([]byte(p.magic), buffer[p.offset:p.offset+uint64(len(p.magic))]) {
			return p.fsName, nil
		}
	}

	return "", ErrUnknownFilesystemType
}
