package utils

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/efs"
)

const (
	raddr  = "169.254.169.254"
	mdtURL = "http://" + raddr + "/latest/meta-data/"
	iidURL = "http://" + raddr + "/latest/dynamic/instance-identity/document"
)

// IsEC2Instance returns a flag indicating whether the executing host is an EC2
// instance based on whether or not the metadata URL can be accessed.
func IsEC2Instance(ctx types.Context) (bool, error) {
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

type instanceIdentityDoc struct {
	InstanceID       string `json:"instanceId,omitempty"`
	Region           string `json:"region,omitempty"`
	AvailabilityZone string `json:"availabilityZone,omitempty"`
}

// InstanceID returns the instance ID for the local host.
func InstanceID(ctx types.Context) (*types.InstanceID, error) {
	req, err := http.NewRequest(http.MethodGet, iidURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	iid := instanceIdentityDoc{}
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&iid); err != nil {
		return nil, err
	}

	return &types.InstanceID{
		ID:     iid.InstanceID,
		Driver: efs.Name,
		Fields: map[string]string{
			efs.InstanceIDFieldRegion:           iid.Region,
			efs.InstanceIDFieldAvailabilityZone: iid.AvailabilityZone,
		},
	}, nil
}
