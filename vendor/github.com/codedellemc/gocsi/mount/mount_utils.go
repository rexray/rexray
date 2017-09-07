package mount

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"io"
	"path"
	"strings"
)

// procMountsFields is fields per line in procMountsPath as per
// https://www.kernel.org/doc/Documentation/filesystems/proc.txt
const procMountsFields = 9

/*
From https://www.kernel.org/doc/Documentation/filesystems/proc.txt:

3.5	/proc/<pid>/mountinfo - Information about mounts
--------------------------------------------------------

This file contains lines of the form:

36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
(1)(2)(3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)

(1) mount ID:  unique identifier of the mount (may be reused after umount)
(2) parent ID:  ID of parent (or of self for the top of the mount tree)
(3) major:minor:  value of st_dev for files on filesystem
(4) root:  root of the mount within the filesystem
(5) mount point:  mount point relative to the process's root
(6) mount options:  per mount options
(7) optional fields:  zero or more fields of the form "tag[:value]"
(8) separator:  marks the end of the optional fields
(9) filesystem type:  name of filesystem of the form "type[.subtype]"
(10) mount source:  filesystem specific information or "none"
(11) super options:  per super block options

Parsers should ignore all unrecognised optional fields.  Currently the
possible optional fields are:

shared:X  mount is shared in peer group X
master:X  mount is slave to peer group X
propagate_from:X  mount is slave and receives propagation from peer group X (*)
unbindable  mount is unbindable
*/
func readProcMountsFrom(
	file io.Reader,
	info bool,
	expectedFields int) ([]*Info, uint32, error) {

	var (
		mountPoints []*Info
		mountSrcMap map[string]string
	)

	if info {
		mountPoints = []*Info{}
		mountSrcMap = map[string]string{}
	}

	hash := fnv.New32a()
	scanner := bufio.NewReader(file)
	for {
		line, err := scanner.ReadString('\n')

		if err == io.EOF {
			break
		}

		fields := strings.Fields(line)

		// Remove the optional fields that should be ignored.
		for {
			val := fields[6]
			fields = append(fields[:6], fields[7:]...)
			if val == "-" {
				break
			}
		}

		if len(fields) != expectedFields {
			return nil, 0, fmt.Errorf(
				"wrong number of fields (expected %d, got %d): %s",
				expectedFields, len(fields), line)
		}

		// Skip any lines where the source does not start with a leading
		// slash. This means this is not a mount on a "real" device.
		source := fields[7]
		if !strings.HasPrefix(source, "/") {
			continue
		}

		fmt.Fprintf(hash, "%s", line)

		if !info {
			continue
		}

		var (
			bindMountSource string

			root       = fields[3]
			mountPoint = fields[4]
			mountOpts  = strings.Split(fields[5], ",")
			fsType     = fields[6]
		)

		// If this is the first time a source is encountered in the
		// output then cache its mountPoint field as the filesystem path
		// to which the source is mounted as a non-bind mount.
		//
		// Subsequent encounters with the source will resolve it
		// to the cached root value in order to set the mount info's
		// Source field to the the cached mountPont field value + the
		// value of the current line's root field.
		if cachedMountPoint, ok := mountSrcMap[source]; ok {
			bindMountSource = path.Join(cachedMountPoint, root)
		} else {
			mountSrcMap[source] = mountPoint
		}

		mp := &Info{
			Device: source,
			Path:   mountPoint,
			Source: bindMountSource,
			Type:   fsType,
			Opts:   mountOpts,
		}

		mountPoints = append(mountPoints, mp)
	}
	return mountPoints, hash.Sum32(), nil
}
