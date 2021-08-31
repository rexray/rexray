package utils

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/azureud"
)

const (
	raddr    = "169.254.169.254"
	maintURL = "http://" + raddr + "/metadata/v1/maintenance"
)

// IsAzureInstance returns a flag indicating whether the executing host
// is an Azure instance .
func IsAzureInstance(ctx types.Context) (bool, error) {
	client := &http.Client{Timeout: time.Duration(1 * time.Second)}
	req, err := http.NewRequest(http.MethodGet, maintURL, nil)
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
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return true, nil
	}
	return false, nil
}

// InstanceID returns the instance ID for the local host.
func InstanceID(ctx types.Context) (*types.InstanceID, error) {

	// UUID can be obtained as descried in
	// https://azure.microsoft.com/en-us/blog/accessing-and-using-azure-vm-unique-id/
	// but this code will use hostname as ID

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &types.InstanceID{
		ID:     hostname,
		Driver: azureud.Name,
	}, nil
}
