package mount

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	mountFilePath = "/proc/mounts"

	// How many times to retry for a consistent read of /proc/mounts.
	maxListTries = 3
	// Number of fields per line in /proc/mounts as per the fstab man page.
	expectedNumFieldsPerLine = 6
)

var (
	bindRemountOpts = []string{"remount"}
)

// getDiskFormat uses 'lsblk' to see if the given disk is unformatted
func getDiskFormat(disk string) (string, error) {
	args := []string{"-n", "-o", "FSTYPE", disk}

	f := log.Fields{
		"disk": disk,
	}
	log.WithFields(f).WithField("args", args).Info(
		"checking if disk is formatted using lsblk")
	buf, err := exec.Command("lsblk", args...).CombinedOutput()
	out := string(buf)
	log.WithField("output", out).Debug("lsblk output")

	if err != nil {
		log.WithFields(f).WithError(err).Error(
			"Could not determine if disk is formatted")
		return "", err
	}

	// Split lsblk output into lines. Unformatted devices should contain only
	// "\n". Beware of "\n\n", that's a device with one empty partition.
	out = strings.TrimSuffix(out, "\n") // Avoid last empty line
	lines := strings.Split(out, "\n")
	if lines[0] != "" {
		// The device is formatted
		return lines[0], nil
	}

	if len(lines) == 1 {
		// The device is unformatted and has no dependent devices
		return "", nil
	}

	// The device has dependent devices, most probably partitions (LVM, LUKS
	// and MD RAID are reported as FSTYPE and caught above).
	return "unknown data, probably partitions", nil
}

// formatAndMount uses unix utils to format and mount the given disk
func formatAndMount(source, target, fsType string, options []string) error {

	options = append(options, "defaults")
	f := log.Fields{
		"source":  source,
		"target":  target,
		"fsType":  fsType,
		"options": options,
	}

	// Try to mount the disk
	log.WithFields(f).Info("attempting to mount disk")
	mountErr := mount(source, target, fsType, options)
	if mountErr == nil {
		return nil
	}

	// Mount failed. This indicates either that the disk is unformatted or
	// it contains an unexpected filesystem.
	existingFormat, err := getDiskFormat(source)
	if err != nil {
		return err
	}
	if existingFormat == "" {
		// Disk is unformatted so format it.
		args := []string{source}
		// Use 'ext4' as the default
		if len(fsType) == 0 {
			fsType = "ext4"
		}

		if fsType == "ext4" || fsType == "ext3" {
			args = []string{"-F", source}
		}
		f["fsType"] = fsType
		log.WithFields(f).Info(
			"disk appears unformatted, attempting format")

		mkfsCmd := fmt.Sprintf("mkfs.%s", fsType)
		if err := exec.Command(mkfsCmd, args...).Run(); err != nil {
			log.WithFields(f).WithError(err).Error(
				"format of disk failed")
		}

		// the disk has been formatted successfully try to mount it again.
		log.WithFields(f).Info(
			"disk successfully formatted")
		return mount(source, target, fsType, options)
	}

	// Disk is already formatted and failed to mount
	if len(fsType) == 0 || fsType == existingFormat {
		// This is mount error
		return mountErr
	}

	// Block device is formatted with unexpected filesystem
	return fmt.Errorf(
		"failed to mount volume as %q; already contains %s: error: %v",
		fsType, existingFormat, mountErr)
}

// getMounts returns a slice of all the mounted filesystems
func getMounts() ([]*Info, error) {
	_, hash1, err := readProcMounts(mountFilePath, false)
	if err != nil {
		return nil, err
	}

	for i := 0; i < maxListTries; i++ {
		mps, hash2, err := readProcMounts(mountFilePath, true)
		if err != nil {
			return nil, err
		}
		if hash1 == hash2 {
			// Success
			return mps, nil
		}
		hash1 = hash2
	}
	return nil, fmt.Errorf(
		"failed to get a consistent snapshot of %v after %d tries",
		mountFilePath, maxListTries)
}

// bindMount performs a bind mount
func bindMount(source, target string, options []string) error {
	err := doMount("mount", source, target, "", []string{"bind"})
	if err != nil {
		return err
	}
	return doMount("mount", source, target, "", options)
}

// readProcMounts reads the given mountFilePath (normally /proc/mounts)
// and produces a hash of the contents and a list of the mounts as
// Info objects.
func readProcMounts(mountFilePath string, info bool) ([]*Info, uint32, error) {

	file, err := os.Open(mountFilePath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()
	return readProcMountsFrom(file, info)
}

func readProcMountsFrom(file io.Reader, info bool) ([]*Info, uint32, error) {

	var mountPoints []*Info
	if info {
		mountPoints = []*Info{}
	}

	hash := fnv.New32a()
	scanner := bufio.NewReader(file)
	for {
		line, err := scanner.ReadString('\n')
		if err == io.EOF {
			break
		}
		fields := strings.Fields(line)
		if len(fields) != expectedNumFieldsPerLine {
			return nil, 0, fmt.Errorf(
				"wrong number of fields (expected %d, got %d): %s",
				expectedNumFieldsPerLine, len(fields), line)
		}

		fmt.Fprintf(hash, "%s", line)

		if !info {
			continue
		}

		mp := &Info{
			Device: fields[0],
			Path:   fields[1],
			Type:   fields[2],
			Opts:   strings.Split(fields[3], ","),
		}

		freq, err := strconv.Atoi(fields[4])
		if err != nil {
			return nil, 0, err
		}
		mp.Freq = freq

		pass, err := strconv.Atoi(fields[5])
		if err != nil {
			return nil, 0, err
		}
		mp.Pass = pass

		mountPoints = append(mountPoints, mp)
	}
	return mountPoints, hash.Sum32(), nil
}
