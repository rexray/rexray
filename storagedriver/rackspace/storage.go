package rackspace

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/emccode/rexray/storagedriver"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

var (
	providerName string
)

type Driver struct {
	Provider   *gophercloud.ProviderClient
	Client     *gophercloud.ServiceClient
	Region     string
	InstanceID string
}

func init() {
	storagedriver.Register("rackspace", Init)
	providerName = "RackSpace"
}

func getInstanceID() (string, error) {
	cmdOut, err := exec.Command("/usr/bin/xenstore-read", "name").Output()
	if err != nil {
		return "", fmt.Errorf("%s: %s", storagedriver.ErrDriverInstanceDiscovery, err)
	}

	instanceID := strings.Replace(string(cmdOut), "\n", "", -1)

	validInstanceID := regexp.MustCompile(`^instance-`)
	valid := validInstanceID.MatchString(instanceID)
	if !valid {
		return "", storagedriver.ErrDriverInstanceDiscovery
	}

	instanceID = strings.Replace(instanceID, "instance-", "", 1)
	return instanceID, nil
}

func Init() (storagedriver.Driver, error) {

	instanceID, err := getInstanceID()
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	region, err := getInstanceRegion()
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	region = strings.ToUpper(region)
	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	driver := &Driver{
		Provider:   provider,
		Client:     client,
		Region:     region,
		InstanceID: instanceID,
	}

	return driver, nil
}

func (driver *Driver) getInstance() (*servers.Server, error) {
	server, err := servers.Get(driver.Client, driver.InstanceID).Extract()
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (driver *Driver) GetInstance() (interface{}, error) {
	server, err := driver.getInstance()
	if err != nil {
		return nil, err
	}

	instance := &storagedriver.Instance{
		ProviderName: providerName,
		InstanceID:   driver.InstanceID,
		Region:       driver.Region,
		Name:         server.Name,
	}

	return instance, nil
}

func (driver *Driver) GetBlockDeviceMapping() (interface{}, error) {
	blockDevices, err := driver.getBlockDevices(driver.InstanceID)
	if err != nil {
		return nil, err
	}

	var BlockDevices []*storagedriver.BlockDevice
	for _, blockDevice := range blockDevices {
		sdBlockDevice := &storagedriver.BlockDevice{
			ProviderName: providerName,
			InstanceID:   driver.InstanceID,
			VolumeID:     blockDevice.VolumeID,
			DeviceName:   blockDevice.Device,
			Region:       driver.Region,
			Status:       "",
		}
		BlockDevices = append(BlockDevices, sdBlockDevice)
	}

	return BlockDevices, nil

}

func (driver *Driver) getBlockDevices(instanceID string) ([]*volumeattach.VolumeAttachment, error) {
	volumes := volumeattach.Get(driver.Client, driver.InstanceID, "")

	volumeAttachments, err := volumes.VolumeAttachmentResult.ExtractAll()
	if err != nil {
		return []*volumeattach.VolumeAttachment{}, fmt.Errorf("Error: %v", err)
	}

	return volumeAttachments, nil

}

func getInstanceRegion() (string, error) {
	cmdOut, err := exec.Command("/usr/bin/xenstore-read", "vm-data/provider_data/region").Output()
	if err != nil {
		return "", fmt.Errorf("Error: %v", err)
	}

	region := strings.Replace(string(cmdOut), "\n", "", -1)
	return region, nil
}
