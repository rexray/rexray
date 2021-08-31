package utils

import (
	"io/ioutil"
	"net/http"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/drivers/storage/dobs"
)

const (
	metadataBase   = "169.254.169.254"
	metadataURL    = "http://" + metadataBase + "/metadata/v1"
	metadataID     = metadataURL + "/id"
	metadataRegion = metadataURL + "/region"
	metadataName   = metadataURL + "/hostname"
)

// InstanceID gets the instance information from the droplet
func InstanceID(ctx types.Context) (*types.InstanceID, error) {

	id, err := getURL(ctx, metadataID)
	if err != nil {
		return nil, err
	}

	region, err := getURL(ctx, metadataRegion)
	if err != nil {
		return nil, err
	}

	name, err := getURL(ctx, metadataName)
	if err != nil {
		return nil, err
	}

	return &types.InstanceID{
		ID:     id,
		Driver: dobs.Name,
		Fields: map[string]string{
			dobs.InstanceIDFieldRegion: region,
			dobs.InstanceIDFieldName:   name,
		},
	}, nil
}

// IsDroplet is a simple check to see if code is being executed on a DigitalOcean droplet or not
func IsDroplet(ctx types.Context) (bool, error) {
	_, err := getURL(ctx, metadataURL)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func getURL(ctx types.Context, url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := doRequest(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	id, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(id), nil
}
