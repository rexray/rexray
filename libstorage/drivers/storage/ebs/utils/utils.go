package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/ebs"
)

const (
	raddr  = "169.254.169.254"
	mdtURL = "http://" + raddr + "/latest/meta-data/"
	iidURL = "http://" + raddr + "/latest/dynamic/instance-identity/document"
	bdmURL = "http://" + raddr + "/latest/meta-data/block-device-mapping/"
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
func InstanceID(
	ctx types.Context,
	driverName string) (*types.InstanceID, error) {

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
		Driver: driverName,
		Fields: map[string]string{
			ebs.InstanceIDFieldRegion:           iid.Region,
			ebs.InstanceIDFieldAvailabilityZone: iid.AvailabilityZone,
		},
	}, nil
}

// BlockDevices returns the EBS devices attached to the local host.
func BlockDevices(ctx types.Context) ([]byte, error) {

	req, err := http.NewRequest(http.MethodGet, bdmURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// BlockDeviceName returns the name of the provided EBS device.
func BlockDeviceName(
	ctx types.Context,
	device string) ([]byte, error) {

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s%s", bdmURL, device),
		nil)
	if err != nil {
		return nil, err
	}

	res, err := doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
