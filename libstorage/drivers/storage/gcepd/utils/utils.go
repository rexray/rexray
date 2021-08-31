package utils

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/gcepd"
)

const (
	mdtURL  = "http://metadata.google.internal/computeMetadata/v1/"
	iidURL  = mdtURL + "instance/id"
	nameURL = mdtURL + "instance/hostname"
	pidURL  = mdtURL + "project/project-id"
	zoneURL = mdtURL + "instance/zone"
	diskURL = mdtURL + "instance/disks/?recursive=true"

	// DiskNameRX contains the regex pattern for matching a valid GCE disk
	// name the first character must be a lowercase letter, and all following
	// characters must be a dash, lowercase letter, or digit, except the
	// last character, which cannot be a dash.
	DiskNameRX = `[a-z](?:[-a-z0-9]*[a-z0-9])?$`
)

var (
	diskRegex = regexp.MustCompile(`^` + DiskNameRX)
)

// IsGCEInstance returns a flag indicating whether the executing host is a GCE
// instance based on whether or not the metadata URL can be accessed.
func IsGCEInstance(ctx types.Context) (bool, error) {
	client := &http.Client{Timeout: time.Duration(1 * time.Second)}
	req, err := http.NewRequest(http.MethodHead, mdtURL, nil)
	if err != nil {
		return false, err
	}
	res, err := doRequestWithClient(ctx, client, req)
	if err != nil {
		if terr, ok := err.(net.Error); ok && terr.Timeout() {
			return false, nil
		}
		return false, err
	}
	if res.StatusCode >= 200 || res.StatusCode <= 299 {
		return true, nil
	}
	return false, nil
}

// InstanceID returns the instance ID for the local host.
func InstanceID(ctx types.Context) (*types.InstanceID, error) {

	hostname, err := getCurrentShortHostname(ctx)
	if err != nil {
		return nil, err
	}

	projectID, err := getCurrentProjectID(ctx)
	if err != nil {
		return nil, err
	}

	zone, err := getCurrentZone(ctx)
	if err != nil {
		return nil, err
	}

	return &types.InstanceID{
		ID:     *hostname,
		Driver: gcepd.Name,
		Fields: map[string]string{
			gcepd.InstanceIDFieldProjectID: projectID,
			gcepd.InstanceIDFieldZone:      zone,
		},
	}, nil
}

func getMetadata(ctx types.Context, url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Metadata-Flavor", "Google")

	res, err := doRequest(ctx, req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	s := buf.String()
	return s, nil
}

func getCurrentInstanceID(ctx types.Context) (string, error) {
	return getMetadata(ctx, iidURL)
}

func getCurrentHostname(ctx types.Context) (string, error) {
	return getMetadata(ctx, nameURL)
}

func getCurrentShortHostname(ctx types.Context) (*string, error) {
	fqdn, err := getCurrentHostname(ctx)
	if err != nil {
		return nil, err
	}
	hostname := strings.Split(fqdn, ".")[0]
	return &hostname, nil
}

func getCurrentProjectID(ctx types.Context) (string, error) {
	return getMetadata(ctx, pidURL)
}

func getCurrentZone(ctx types.Context) (string, error) {
	zone, err := getMetadata(ctx, zoneURL)
	if err != nil {
		return "", err
	}
	zone = GetIndex(zone)
	return zone, nil
}

// GetIndex returns the trailing "document" in a URL
func GetIndex(href string) string {
	hrefFields := strings.Split(href, "/")
	return hrefFields[len(hrefFields)-1]
}

// Disk holds the data returned in the disks metadata
type Disk struct {
	DeviceName string `json:"deviceName"`
	Index      uint32 `json:"index"`
	Mode       string `json:"mode"`
	Type       string `json:"type"`
}

// GetDisks returns a string slice containing the names of all the disks
// attached to the local instance
func GetDisks(ctx types.Context) (map[string]string, error) {
	disksJSON, err := getMetadata(ctx, diskURL)
	if err != nil {
		return nil, err
	}

	disks := make(map[string]string)
	diskList := make([]*Disk, 1, 1)

	// TODO: fix taking the bytes from original request being returned
	// as a string, only to be cast back to byte here
	err = json.Unmarshal([]byte(disksJSON), &diskList)
	if err != nil {
		return nil, err
	}

	for _, disk := range diskList {
		disks[disk.DeviceName] = disk.DeviceName
	}

	return disks, nil
}

// IsValidDiskName returns a boolean of whether the given name is valid for a
// GCE disk
func IsValidDiskName(name *string) bool {
	if name == nil || *name == "" {
		return false
	}
	return diskRegex.MatchString(*name)
}
