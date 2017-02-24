// +build !libstorage_storage_driver libstorage_storage_driver_fittedcloud

package fcagent

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	// FcVolMaxCloudDisk is the max number of constituent disks for a FittedCloud volume
	FcVolMaxCloudDisk = 6
)

// IsRunning checks if FittedCloud Agent is running
func IsRunning() (string, error) {
	var (
		cmdOut []byte
		err    error
		path   string
	)
	cmd := "fcagent"
	args := []string{"echo"}

	if path, err = exec.LookPath(cmd); err != nil {
		log.Debug(err)
	}
	log.Debug(path, args[0])

	if cmdOut, err = exec.Command(cmd, args...).Output(); err != nil {
		log.Debug(err)
	}
	return string(cmdOut), err
}

// CreateVol creates a FittedCluod volume
func CreateVol(tagName string, sizeGb int, encryption bool, kmsKeyID string) (string, error) {
	var (
		cmdOut    []byte
		err       error
		fcVolName string
	)
	sizeGbStr := strconv.Itoa(sizeGb)
	initGb := (sizeGb + FcVolMaxCloudDisk - 1) / FcVolMaxCloudDisk
	initGbStr := strconv.Itoa(initGb)
	log.Debug("sizeGbStr=" + sizeGbStr + " initGbStr=" + initGbStr)

	// Find the next available fcvol name
	fcVolName, err = GetFcVolName("next")
	if err != nil {
		return "", err
	}

	cmd := "fcagent"
	args := make([]string, 0, 8)
	args = append(args, "createvol")
	args = append(args, fcVolName)
	args = append(args, sizeGbStr)
	args = append(args, initGbStr)
	args = append(args, "--delete_on_termination=no")
	if encryption {
		args = append(args, "--encryption=yes")
		if kmsKeyID != "" {
			args = append(args, "--kms_key="+kmsKeyID)
		}
	} else {
		args = append(args, "--encryption=no")
	}
	args = append(args, "--ebs_tag="+tagName)

	log.Debug(cmd, args)

	if cmdOut, err = exec.Command(cmd, args...).Output(); err != nil {
		log.Debug(err)
		return string(cmdOut), err
	}
	return fcVolName, nil
}

// DelVol deletes a FittedCloud volume by name
func DelVol(fcVolName string) (string, error) {
	var (
		cmdOut []byte
		err    error
	)

	strs := strings.SplitAfterN(fcVolName, "/dev/", 2)
	if len(strs) == 2 && strs[0] == "/dev/" {
		fcVolName = strs[1]
	}
	log.Debug("fcVolName=", fcVolName)

	cmd := "fcagent"
	args := []string{
		"delvol",
		fcVolName,
		"-donotdelebs"}

	log.Debug(cmd, args)

	if cmdOut, err = exec.Command(cmd, args...).Output(); err != nil {
		log.Debug(err)
	}

	return string(cmdOut), err
}

func getFcVolName(devName string) (string, error) {
	var (
		cmdOut []byte
		err    error
	)

	cmd := "fcagent"
	args := []string{
		"getfcvolname",
		devName}

	log.Debug(cmd, args)

	if cmdOut, err = exec.Command(cmd, args...).Output(); err != nil {
		log.Debug(err)
	}

	return string(cmdOut), err
}

// GetFcVolName returns the FittedCloud volume name
func GetFcVolName(devName string) (string, error) {
	var (
		cmdOut string
		err    error
	)
	if cmdOut, err = getFcVolName(devName); err != nil {
		return "", err
	}
	cmdOut = strings.TrimRight(cmdOut, "\n")

	strs := strings.SplitAfterN(cmdOut, "fcvol=", 2)
	if len(strs) == 2 {
		if strs[1] == "NOTFOUND" {
			return cmdOut, errors.New("fcvol not found")
		}
		cmdOut = strs[1]
	}

	return cmdOut, err
}
