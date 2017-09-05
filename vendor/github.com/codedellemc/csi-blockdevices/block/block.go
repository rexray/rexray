package block

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	fsmap = map[string]struct{}{
		"xfs":   struct{}{},
		"ext3":  struct{}{},
		"ext4":  struct{}{},
		"btrfs": struct{}{},
	}
)

// Device is a struct for holding details about a compatible block device
type Device struct {
	Capacity uint64
	FullPath string
	Name     string
	RealDev  string
}

// Supported queries the underlying system to check if the required system
// executables are present
// If not, it returns an error
func Supported() error {
	switch runtime.GOOS {
	case "linux":
		fss, err := GetHostFileSystems("")
		if err != nil {
			return err
		}
		if len(fss) == 0 {
			return fmt.Errorf("%s", "No supported filesystems found")
		}
		return nil
	default:
		return fmt.Errorf("%s", "Plugin only supported on Linux OS")
	}
}

// GetHostFileSystems returns a slice of strings of filesystems supported by the
// host. Supported filesytems are restricted to ext3,ext4,xfs,btrfs
func GetHostFileSystems(binPath string) ([]string, error) {
	if binPath == "" {
		binPath = "/sbin"
	}

	s := filepath.Join(binPath, "mkfs.*")
	m, err := filepath.Glob(s)
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, nil
	}

	fields := log.Fields{
		"binpath":  binPath,
		"globpath": s,
		"binaries": m,
	}

	log.WithFields(fields).Debug("found mkfs binaries")

	fss := make([]string, 0)
	for _, f := range m {
		fs := filepath.Ext(f)
		fs = strings.TrimLeft(fs, ".")
		if _, ok := fsmap[fs]; ok {
			fss = append(fss, fs)
		}
	}
	log.WithField("filesystems", fss).Info("found supported filesystems")

	return fss, nil
}

// GetDeviceInDir returns a Device struct with info about the given device,
// by looking for name in dir.
func GetDeviceInDir(dir, name string) (*Device, error) {
	dp := filepath.Join(dir, name)
	return GetDevice(dp)
}

// GetDevice returns a Device struct with info about the given device, or
// an error if it doesn't exist or is not a block device
func GetDevice(path string) (*Device, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	m := fi.Mode()
	if m&os.ModeSymlink == 0 {
		return nil, fmt.Errorf("file %s is not a symlink", path)
	}

	// eval the symlink and make sure it points to a device
	d, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}

	// TODO does EvalSymlinks throw error if link is to non-
	// existent file? assuming so by masking error below
	ds, _ := os.Stat(d)
	dm := ds.Mode()
	if dm&os.ModeDevice == 0 {
		return nil, fmt.Errorf(
			"file linked to by %s is not a block device", path)
	}

	return &Device{
		Name: fi.Name(),
		// TODO: This size appears to report 0. FileInfo.Size() is system
		// dependent for non-regular files, so probably need to use a Linux
		// specific method here
		Capacity: uint64(ds.Size()),
		FullPath: path,
		RealDev:  d,
	}, nil
}

// ListDevices returns a slice of Device for all valid blockdevices found
// in the given device directory
func ListDevices(dir string) ([]*Device, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path %s does not exist", dir)
		}
		return nil, err
	}
	mode := fi.Mode()
	if !mode.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", dir)
	}

	fields := log.Fields{
		"path": dir,
	}

	devs := []*Device{}

	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error(err)
			return err
		}
		if info.IsDir() {
			if path == dir {
				return nil
			}
			log.WithField("file", path).Debug("skipping dir")
			return filepath.SkipDir
		}
		log.WithFields(fields).WithField("file", info.Name()).Debug(
			"examining file")

		dev, deverr := GetDevice(path)
		if deverr != nil {
			log.WithFields(fields).WithField("file", info.Name()).WithError(deverr).Debug(
				"not a device")
			return nil
		}
		log.WithFields(fields).WithField("file", info.Name()).Debug(
			"found device")
		devs = append(devs, dev)
		return nil
	}

	log.WithFields(fields).Debug("listing devices")
	err = filepath.Walk(dir, walk)
	if err != nil {
		return nil, err
	}
	return devs, nil
}
