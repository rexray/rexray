package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/ebs"
)

const (
	raddr  = "169.254.169.254"
	iidURL = "http://" + raddr + "/latest/dynamic/instance-identity/document"
	bdmURL = "http://" + raddr + "/latest/meta-data/block-device-mapping/"
)

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
		Driver: ebs.Name,
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
