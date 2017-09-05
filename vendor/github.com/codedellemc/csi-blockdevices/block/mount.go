package block

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
	mntCmd        = "mount"
	mountFilePath = "/proc/mounts"

	// How many times to retry for a consistent read of /proc/mounts.
	maxListTries = 3
	// Number of fields per line in /proc/mounts as per the fstab man page.
	expectedNumFieldsPerLine = 6
)

// MountPoint represents a single line in /proc/mounts
type MountPoint struct {
	Device string
	Path   string
	Type   string
	Opts   []string
	Freq   int
	Pass   int
}

// Most of this file is based on k8s.io/pkg/util/mount

// GetDiskFormat uses 'lsblk' to see if the given disk is unformated
func GetDiskFormat(disk string) (string, error) {
	args := []string{"-n", "-o", "FSTYPE", disk}

	f := log.Fields{
		"disk": disk,
	}
	log.WithFields(f).WithField("args", args).Info(
		"checking if disk is formatted using lsblk")
	dataOut, err := exec.Command("lsblk", args...).CombinedOutput()
	output := string(dataOut)
	log.WithField("output", output).Debug()

	if err != nil {
		log.WithFields(f).WithError(err).Error(
			"Could not determine if disk is formatted")
		return "", err
	}

	// Split lsblk output into lines. Unformatted devices should contain only
	// "\n". Beware of "\n\n", that's a device with one empty partition.
	output = strings.TrimSuffix(output, "\n") // Avoid last empty line
	lines := strings.Split(output, "\n")
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

// FormatAndMount uses unix utils to format and mount the given disk
func FormatAndMount(
	source string,
	target string,
	fstype string,
	options []string) error {

	options = append(options, "defaults")
	f := log.Fields{
		"source":  source,
		"target":  target,
		"fs":      fstype,
		"options": options,
	}

	// Try to mount the disk
	log.WithFields(f).Info("attempting to mount disk")
	mountErr := Mount(source, target, fstype, options)
	if mountErr != nil {
		// Mount failed. This indicates either that the disk is unformatted or
		// it contains an unexpected filesystem.
		existingFormat, err := GetDiskFormat(source)
		if err != nil {
			return err
		}
		if existingFormat == "" {
			// Disk is unformatted so format it.
			args := []string{source}
			// Use 'ext4' as the default
			if len(fstype) == 0 {
				fstype = "ext4"
			}

			if fstype == "ext4" || fstype == "ext3" {
				args = []string{"-F", source}
			}
			f["fs"] = fstype
			log.WithFields(f).Info(
				"disk appears unformatted, attempting format")
			err := exec.Command("mkfs."+fstype, args...).Run()
			if err == nil {
				// the disk has been formatted successfully try to mount it again.
				log.WithFields(f).Info(
					"disk successfully formatted")
				return Mount(source, target, fstype, options)
			}
			log.WithFields(f).WithError(err).Error(
				"format of disk failed")
			return err
		}

		// Disk is already formatted and failed to mount
		if len(fstype) == 0 || fstype == existingFormat {
			// This is mount error
			return mountErr
		}

		// Block device is formatted with unexpected filesystem
		return fmt.Errorf(
			"failed to mount the volume as %q, it already contains %s. Mount error: %v",
			fstype, existingFormat, mountErr)
	}
	return mountErr
}

// Mount mounts source to target as fstype with given options. 'source' and 'fstype' must
// be an emtpy string in case it's not required, e.g. for remount, or for auto filesystem
// type, where kernel handles fs type for you. The mount 'options' is a list of options,
// currently come from mount(8), e.g. "ro", "remount", "bind", etc. If no more option is
// required, call Mount with an empty string list or nil.
func Mount(
	source string,
	target string,
	fstype string,
	options []string) error {

	// All Linux distros are expected to be shipped with a mount utility that an support bind mounts.
	bind, bindRemountOpts := isBind(options)
	if bind {
		err := doMount(source, target, fstype, []string{"bind"})
		if err != nil {
			return err
		}
		return doMount(source, target, fstype, bindRemountOpts)
	}
	return doMount(source, target, fstype, options)
}

// Unmount unmounts the target.
func Unmount(target string) error {
	f := log.Fields{
		"path": target,
		"cmd":  "umount",
	}
	log.WithFields(f).Info("unmount command")
	dataOut, err := exec.Command("umount", target).CombinedOutput()
	if err != nil {
		out := string(dataOut)
		f["output"] = out
		log.WithFields(f).WithError(err).Error("unmount failed")
		return fmt.Errorf("unmount failed: %v\nUnmounting arguments: %s\nOutput: %s",
			err, target, string(out))
	}
	return nil
}

// isBind detects whether a bind mount is being requested and makes the remount options to
// use in case of bind mount, due to the fact that bind mount doesn't respect mount options.
// The list equals:
//   options - 'bind' + 'remount' (no duplicate)
func isBind(options []string) (bool, []string) {
	bindRemountOpts := []string{"remount"}
	bind := false

	if len(options) != 0 {
		for _, option := range options {
			switch option {
			case "bind":
				bind = true
				break
			case "remount":
				break
			default:
				bindRemountOpts = append(bindRemountOpts, option)
			}
		}
	}

	return bind, bindRemountOpts
}

// doMount runs the mount command.
func doMount(
	source string,
	target string,
	fstype string,
	options []string) error {

	mountArgs := makeMountArgs(source, target, fstype, options)
	args := strings.Join(mountArgs, " ")

	f := log.Fields{
		"cmd":  mntCmd,
		"args": args,
	}
	log.WithFields(f).Info("mount command")

	dataOut, err := exec.Command(mntCmd, mountArgs...).CombinedOutput()
	if err != nil {
		output := string(dataOut)
		log.WithFields(f).WithField("output", output).WithError(
			err).Error("Mount Failed")
		return fmt.Errorf("mount failed: %v\nmounting arguments: %s\noutput: %s",
			err, args, output)
	}
	return nil
}

// makeMountArgs makes the arguments to the mount(8) command.
func makeMountArgs(
	source, target, fstype string,
	options []string) []string {

	// Build mount command as follows:
	//   mount [-t $fstype] [-o $options] [$source] $target

	// Remove all duplicates and empty strings from options
	opts := make([]string, 0)
	for _, x := range options {
		if x != "" && !contains(opts, x) {
			opts = append(opts, x)
		}
	}

	mountArgs := []string{}
	if len(fstype) > 0 {
		mountArgs = append(mountArgs, "-t", fstype)
	}
	if len(options) > 0 {
		mountArgs = append(mountArgs, "-o", strings.Join(opts, ","))
	}
	if len(source) > 0 {
		mountArgs = append(mountArgs, source)
	}
	mountArgs = append(mountArgs, target)

	return mountArgs
}

// GetMounts returns a slice of all mounted filesystems
func GetMounts() ([]MountPoint, error) {
	hash1, err := readProcMounts(mountFilePath, nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < maxListTries; i++ {
		mps := []MountPoint{}
		hash2, err := readProcMounts(mountFilePath, &mps)
		if err != nil {
			return nil, err
		}
		if hash1 == hash2 {
			// Success
			return mps, nil
		}
		hash1 = hash2
	}
	return nil, fmt.Errorf("failed to get a consistent snapshot of %v after %d tries", mountFilePath, maxListTries)
}

// GetDevMounts returns a slice of all mounts for dev
func GetDevMounts(dev string) ([]MountPoint, error) {
	allMnts, err := GetMounts()
	if err != nil {
		return nil, err
	}

	mnts := make([]MountPoint, 0)
	for _, m := range allMnts {
		if m.Device == dev {
			mnts = append(mnts, m)
		}
	}

	return mnts, nil
}

// readProcMounts reads the given mountFilePath (normally /proc/mounts) and produces a hash
// of the contents.  If the out argument is not nil, this fills it with MountPoint structs.
func readProcMounts(
	mountFilePath string,
	out *[]MountPoint) (uint32, error) {

	file, err := os.Open(mountFilePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return readProcMountsFrom(file, out)
}

func readProcMountsFrom(
	file io.Reader,
	out *[]MountPoint) (uint32, error) {

	hash := fnv.New32a()
	scanner := bufio.NewReader(file)
	for {
		line, err := scanner.ReadString('\n')
		if err == io.EOF {
			break
		}
		fields := strings.Fields(line)
		if len(fields) != expectedNumFieldsPerLine {
			return 0, fmt.Errorf("wrong number of fields (expected %d, got %d): %s", expectedNumFieldsPerLine, len(fields), line)
		}

		fmt.Fprintf(hash, "%s", line)

		if out != nil {
			mp := MountPoint{
				Device: fields[0],
				Path:   fields[1],
				Type:   fields[2],
				Opts:   strings.Split(fields[3], ","),
			}

			freq, err := strconv.Atoi(fields[4])
			if err != nil {
				return 0, err
			}
			mp.Freq = freq

			pass, err := strconv.Atoi(fields[5])
			if err != nil {
				return 0, err
			}
			mp.Pass = pass

			*out = append(*out, mp)
		}
	}
	return hash.Sum32(), nil
}

func contains(list []string, item string) bool {
	for _, x := range list {
		if x == item {
			return true
		}
	}
	return false
}
