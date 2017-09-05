// +build linux darwin

package mount

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// mount mounts source to target as fsType with given options.
//
// The parameters 'source' and 'fsType' must be empty strings in case they
// are not required, e.g. for remount, or for an auto filesystem type where
// the kernel handles fsType automatically.
//
// The 'options' parameter is a list of options. Please see mount(8) for
// more information. If no options are required then please invoke Mount
// with an empty or nil argument.
func mount(source, target, fsType string, options []string) error {

	// All Linux distributes should support bind mounts.
	if options, ok := isBind(options); ok {
		return bindMount(source, target, options)
	}
	return doMount("mount", source, target, fsType, options)
}

// doMount runs the mount command.
func doMount(mntCmd, source, target, fsType string, options []string) error {

	mountArgs := makeMountArgs(source, target, fsType, options)
	args := strings.Join(mountArgs, " ")

	f := log.Fields{
		"cmd":  mntCmd,
		"args": args,
	}
	log.WithFields(f).Info("mount command")

	buf, err := exec.Command(mntCmd, mountArgs...).CombinedOutput()
	if err != nil {
		out := string(buf)
		log.WithFields(f).WithField("output", out).WithError(
			err).Error("mount Failed")
		return fmt.Errorf(
			"mount failed: %v\nmounting arguments: %s\noutput: %s",
			err, args, out)
	}
	return nil
}

// unmount unmounts the target.
func unmount(target string) error {
	f := log.Fields{
		"path": target,
		"cmd":  "umount",
	}
	log.WithFields(f).Info("unmount command")
	buf, err := exec.Command("umount", target).CombinedOutput()
	if err != nil {
		out := string(buf)
		f["output"] = out
		log.WithFields(f).WithError(err).Error("unmount failed")
		return fmt.Errorf(
			"unmount failed: %v\nunmounting arguments: %s\nOutput: %s",
			err, target, out)
	}
	return nil
}

// isBind detects whether a bind mount is being requested and determines
// which remount options are needed. A secondary mount operation is
// required for bind mounts as the initial operation does not apply the
// request mount options.
//
// The returned options will be "bind", "remount", and the provided
// list of options.
func isBind(options []string) ([]string, bool) {
	bind := false
	remountOpts := append([]string(nil), bindRemountOpts...)

	for _, o := range options {
		switch o {
		case "bind":
			bind = true
			break
		case "remount":
			break
		default:
			remountOpts = append(remountOpts, o)
		}
	}

	return remountOpts, bind
}

// getDevMounts returns a slice of all mounts for dev
func getDevMounts(dev string) ([]*Info, error) {
	allMnts, err := getMounts()
	if err != nil {
		return nil, err
	}

	mnts := make([]*Info, 0)
	for _, m := range allMnts {
		if m.Device == dev {
			mnts = append(mnts, m)
		}
	}

	return mnts, nil
}

// makeMountArgs makes the arguments to the mount(8) command.
func makeMountArgs(source, target, fsType string, options []string) []string {

	// Build mount command as follows:
	//   mount [-t $fsType] [-o $options] [$source] $target

	// Remove all duplicates and empty strings from options
	opts := make([]string, 0)
	for _, x := range options {
		if x != "" && !contains(opts, x) {
			opts = append(opts, x)
		}
	}

	mountArgs := []string{}
	if len(fsType) > 0 {
		mountArgs = append(mountArgs, "-t", fsType)
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
