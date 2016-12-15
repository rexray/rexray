// +build !libstorage_storage_driver libstorage_storage_driver_rbd

package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/akutz/goof"
)

const (
	cephCmd   = "ceph"
	rbdCmd    = "rbd"
	formatOpt = "--format"
	jsonArg   = "json"
	poolOpt   = "--pool"

	bytesPerGiB = 1024 * 1024 * 1024
)

type rbdMappedEntry struct {
	Device string `json:"device"`
	Name   string `json:"name"`
	Pool   string `json:"pool"`
	Snap   string `json:"snap"`
}

//RBDImage holds details about an RBD image
type RBDImage struct {
	Name   string `json:"image"`
	Size   int64  `json:"size"`
	Format uint   `json:"format"`
	Pool   string
}

//RBDInfo holds low-level details about an RBD image
type RBDInfo struct {
	Name            string   `json:"name"`
	Size            int64    `json:"size"`
	Objects         int64    `json:"objects"`
	Order           int64    `json:"order"`
	ObjectSize      int64    `json:"object_size"`
	BlockNamePrefix string   `json:"block_name_prefix"`
	Format          int64    `json:"format"`
	Features        []string `json:"features"`
	Pool            string
}

//GetRadosPools returns a slice containing all the pool names
func GetRadosPools() ([]*string, error) {

	cmd := exec.Command(cephCmd, "osd", "pool", "ls", formatOpt, jsonArg)
	log.Debugf("running command: %v", cmd.Args)

	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			stderr := string(exiterr.Stderr)
			log.Errorf("Unable to get pools: %s", stderr)
			return nil,
				goof.Newf("Unable to get pools: %s", stderr)
		}
		return nil, goof.WithError("Unable to get pools", err)
	}

	var pools []string
	err = json.Unmarshal(out, &pools)
	if err != nil {
		return nil, goof.WithError(
			"Unable to parse ceph lspools", err)
	}

	return ConvStrArrayToPtr(pools), nil
}

//GetRBDImages returns a slice of RBD image info
func GetRBDImages(pool *string) ([]*RBDImage, error) {

	cmd := exec.Command(rbdCmd, "ls", "-p", *pool, "-l", formatOpt, jsonArg)
	log.Debugf("running command: %v", cmd.Args)

	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			stderr := string(exiterr.Stderr)
			log.Errorf("Unable to get rbd images: %s", stderr)
			return nil,
				goof.Newf("Unable to get rbd images: %s",
					stderr)
		}
		return nil, goof.WithError("Unable to get rbd images", err)
	}

	var rbdList []*RBDImage

	err = json.Unmarshal(out, &rbdList)
	if err != nil {
		return nil, goof.WithError(
			"Unable to parse rbd ls", err)
	}

	for _, info := range rbdList {
		info.Pool = *pool
	}

	return rbdList, nil
}

//GetRBDInfo gets low-level details about an RBD image
func GetRBDInfo(pool *string, name *string) (*RBDInfo, error) {

	cmd := exec.Command(
		rbdCmd, "info", "-p", *pool, *name, formatOpt, jsonArg)

	log.Debugf("running command: %v", cmd.Args)

	out, err := cmd.Output()

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 2 {
					// image does not exist
					return nil, nil
				}
			}
			stderr := string(exiterr.Stderr)
			log.Errorf("Unable to get rbd info: %s", stderr)
			return nil,
				goof.Newf("Unable to get rbd info: %s",
					stderr)
		}
		return nil, goof.WithError("Unable to get rbd info", err)
	}

	info := &RBDInfo{}

	err = json.Unmarshal(out, info)
	if err != nil {
		return nil, goof.WithError(
			"Unable to parse rbd info", err)
	}

	info.Pool = *pool

	return info, nil
}

//GetVolumeID returns an RBD Volume formatted as <pool>.<imageName>
func GetVolumeID(pool, image *string) *string {

	volumeID := fmt.Sprintf("%s.%s", *pool, *image)
	return &volumeID
}

//GetMappedRBDs returns a map of RBDs currently mapped to the *local* host
func GetMappedRBDs() (map[string]string, error) {

	out, err := exec.Command(
		rbdCmd, "showmapped", formatOpt, jsonArg).Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			stderr := string(exiterr.Stderr)
			log.Errorf("Unable to get rbd map: %s", stderr)
			return nil,
				goof.Newf("Unable to get RBD map: %s",
					stderr)
		}
		return nil, goof.WithError("Unable to get RBD map", err)
	}

	devMap := map[string]string{}
	rbdMap := map[string]*rbdMappedEntry{}

	err = json.Unmarshal(out, &rbdMap)
	if err != nil {
		return nil, goof.WithError(
			"Unable to parse rbd showmapped", err)
	}

	for _, mapped := range rbdMap {
		volumeID := GetVolumeID(&mapped.Pool, &mapped.Name)
		devMap[*volumeID] = mapped.Device
	}

	return devMap, nil
}

//RBDCreate creates a new RBD volume on the cluster
func RBDCreate(
	pool *string,
	image *string,
	sizeGB *int64,
	objectSize *string,
	features []*string) error {

	cmd := exec.Command(
		rbdCmd, "create", poolOpt, *pool,
		"--object-size", *objectSize,
		"--size", strconv.FormatInt(*sizeGB, 10)+"G",
	)

	for _, feature := range features {
		cmd.Args = append(cmd.Args, "--image-feature")
		cmd.Args = append(cmd.Args, *feature)
	}

	cmd.Args = append(cmd.Args, *image)
	log.Debugf("running command: %v", cmd.Args)

	err := cmd.Run()

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			stderr := string(exiterr.Stderr)
			log.Errorf("Unable to create RBD: %s", stderr)
			return goof.Newf("Unable to create RBD: %s",
				stderr)
		}
		return goof.WithError("Unable to create RBD", err)
	}

	return nil
}

//RBDRemove deletes the RBD volume on the cluster
func RBDRemove(pool *string, image *string) error {
	cmd := exec.Command(rbdCmd, "rm", poolOpt, *pool, "--no-progress",
		*image,
	)
	log.Debugf("running command: %v", cmd.Args)

	err := cmd.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			stderr := string(exiterr.Stderr)
			log.Errorf("Unable to delete RBD: %s", stderr)
			return goof.Newf("Error deleting RBD: %s",
				stderr)
		}
		return goof.WithError("Error deleting RBD", err)
	}

	return nil
}

//RBDMap attaches the given RBD image to the *local* host
func RBDMap(pool, image *string) (string, error) {

	cmd := exec.Command(rbdCmd, "map", poolOpt, *pool, *image)
	log.Debugf("running command: %v", cmd.Args)

	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			stderr := string(exiterr.Stderr)
			log.Errorf("Unable to map RBD: %s", stderr)
			return "",
				goof.Newf("Unable to map RBD: %s",
					stderr)
		}
		return "", goof.WithError("Unable to map RBD", err)
	}

	return strings.TrimSpace(string(out)), nil
}

//RBDUnmap detaches the given RBD device from the *local* host
func RBDUnmap(device *string) error {

	cmd := exec.Command(rbdCmd, "unmap", *device)
	log.Debugf("running command: %v", cmd.Args)

	err := cmd.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			stderr := string(exiterr.Stderr)
			log.Errorf("Unable to unmap RBD: %s", stderr)
			return goof.Newf("Unable to unmap RBD: %s",
				stderr)
		}
		return goof.WithError("Unable to unmap RBD", err)
	}

	return nil
}

//GetRBDStatus returns a map of RBD status info
func GetRBDStatus(pool, image *string) (map[string]interface{}, error) {

	cmd := exec.Command(
		rbdCmd, "status", poolOpt, *pool, *image, formatOpt, jsonArg,
	)
	log.Debugf("running command: %v", cmd.Args)

	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			stderr := string(exiterr.Stderr)
			log.Errorf("Unable to get RBD status: %s", stderr)
			return nil, goof.Newf("Unable to get RBD status: %s",
				stderr)
		}
		return nil, goof.WithError("Unable to get RBD status", err)
	}

	watcherMap := map[string]interface{}{}

	err = json.Unmarshal(out, &watcherMap)
	if err != nil {
		return nil, goof.WithError(
			"Unable to parse rbd status", err)
	}

	return watcherMap, nil
}

//RBDHasWatchers returns true if RBD image has watchers
func RBDHasWatchers(pool *string, image *string) (bool, error) {

	m, err := GetRBDStatus(pool, image)
	if err != nil {
		return false, err
	}

	return len(m["watchers"].(map[string]interface{})) > 0, nil
}

//ConvStrArrayToPtr converts the slice of strings to a slice of pointers to str
func ConvStrArrayToPtr(strArr []string) []*string {
	ptrArr := make([]*string, len(strArr))
	for i := range strArr {
		ptrArr[i] = &strArr[i]
	}
	return ptrArr
}
