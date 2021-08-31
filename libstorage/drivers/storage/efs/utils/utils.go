package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/efs"
)

const (
	raddr  = "169.254.169.254"
	mdtURL = "http://" + raddr + "/latest/meta-data/"
	iidURL = "http://" + raddr + "/latest/dynamic/instance-identity/document"
	macURL = "http://" + raddr + "/latest/meta-data/mac"
	subURL = "http://" + raddr +
		`/latest/meta-data/network/interfaces/macs/%s/subnet-id`
	sgpURL = "http://" + raddr +
		`/latest/meta-data/network/interfaces/macs/%s/security-group-ids`
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

	mac, err := getMAC(ctx)
	if err != nil {
		return nil, err
	}

	subnetID, err := ResolveSubnetWithMAC(ctx, mac)
	if err != nil {
		return nil, err
	}

	secGroups, err := getSecurityGroups(ctx, mac)
	if err != nil {
		return nil, err
	}

	iidFields := map[string]string{
		efs.InstanceIDFieldRegion:           iid.Region,
		efs.InstanceIDFieldAvailabilityZone: iid.AvailabilityZone,
	}

	if len(secGroups) > 0 {
		iidFields[efs.InstanceIDFieldSecurityGroups] = strings.Join(
			secGroups, ";")
	}

	return &types.InstanceID{
		ID:     subnetID,
		Driver: efs.Name,
		Fields: iidFields,
	}, nil
}

// ResolveSubnet determines the VPC subnet ID on the running AWS instance.
func ResolveSubnet(ctx types.Context) (string, error) {
	mac, err := getMAC(ctx)
	if err != nil {
		return "", err
	}
	return ResolveSubnetWithMAC(ctx, mac)
}

// ResolveSubnetWithMAC determines the VPC subnet ID on the running AWS
// instance.
func ResolveSubnetWithMAC(ctx types.Context, mac string) (string, error) {
	subnetID, err := getSubnetID(ctx, mac)
	if err != nil {
		return "", err
	}
	return subnetID, nil
}

func getMAC(ctx types.Context) (string, error) {
	req, err := http.NewRequest(http.MethodGet, macURL, nil)
	if err != nil {
		return "", err
	}
	res, err := doRequest(ctx, req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func getSubnetID(ctx types.Context, mac string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(subURL, mac), nil)
	if err != nil {
		return "", err
	}
	res, err := doRequest(ctx, req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func getSecurityGroups(ctx types.Context, mac string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(sgpURL, mac), nil)
	if err != nil {
		return nil, err
	}
	res, err := doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var secGroups []string
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		secGroups = append(secGroups, scanner.Text())
	}
	return secGroups, nil
}
