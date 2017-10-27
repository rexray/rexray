package mount

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var (
	bindRemountOpts = []string{}
	mountRX         = regexp.MustCompile(`^(.+) on (.+) \((.+)\)$`)
)

// getDiskFormat uses 'lsblk' to see if the given disk is unformated
func getDiskFormat(disk string) (string, error) {
	mps, err := getMounts()
	if err != nil {
		return "", err
	}
	for _, i := range mps {
		if i.Device == disk {
			return i.Type, nil
		}
	}
	return "", fmt.Errorf("failed to get disk format: %s", disk)
}

// formatAndMount uses unix utils to format and mount the given disk
func formatAndMount(source, target, fsType string, options []string) error {
	return ErrNotImplemented
}

// getMounts returns a slice of all the mounted filesystems
func getMounts() ([]*Info, error) {
	out, err := exec.Command("mount").CombinedOutput()
	if err != nil {
		return nil, err
	}
	scan := bufio.NewScanner(bytes.NewReader(out))
	mps := []*Info{}
	for scan.Scan() {
		m := mountRX.FindStringSubmatch(scan.Text())
		if len(m) != 4 {
			continue
		}
		device := m[1]
		if !strings.HasPrefix(device, "/") {
			continue
		}
		var (
			path    = m[2]
			source  = device
			options = strings.Split(m[3], ",")
		)
		if len(options) == 0 {
			return nil, fmt.Errorf("invalid mount options: %s", device)
		}
		for i, v := range options {
			options[i] = strings.TrimSpace(v)
		}
		fsType := options[0]
		if len(options) > 1 {
			options = options[1:]
		} else {
			options = nil
		}
		mps = append(mps, &Info{
			Device: device,
			Path:   path,
			Source: source,
			Type:   fsType,
			Opts:   options,
		})
	}
	return mps, nil
}

// bindMount performs a bind mount
func bindMount(source, target string, options []string) error {
	return doMount("bindfs", source, target, "", options)
}
