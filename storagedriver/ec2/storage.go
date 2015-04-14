package ec2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/emccode/rexray/storagedriver"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
)

var (
	providerName string
)

type Driver struct {
	InstanceDocument *instanceIdentityDocument
	EC2Instance      *ec2.EC2
}

func init() {
	providerName = "ec2"
	storagedriver.Register("ec2", Init)
}

func Init() (storagedriver.Driver, error) {
	instanceDocument, err := getInstanceIdendityDocument()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", storagedriver.ErrDriverInstanceDiscovery, err)
	}

	auth := aws.Auth{AccessKey: os.Getenv("AWS_ACCESS_KEY"), SecretKey: os.Getenv("AWS_SECRET_KEY")}
	ec2Instance := ec2.New(
		auth,
		aws.Regions[instanceDocument.Region],
	)

	driver := &Driver{
		EC2Instance:      ec2Instance,
		InstanceDocument: instanceDocument,
	}

	return driver, nil
}

type instanceIdentityDocument struct {
	InstanceID         string      `json:"instanceId"`
	BillingProducts    interface{} `json:"billingProducts"`
	AccountID          string      `json:"accountId"`
	ImageID            string      `json:"imageId"`
	InstanceType       string      `json:"instanceType"`
	KernelID           string      `json:"kernelId"`
	RamdiskID          string      `json:"ramdiskId"`
	PendingTime        string      `json:"pendingTime"`
	Architecture       string      `json:"architecture"`
	Region             string      `json:"region"`
	Version            string      `json:"version"`
	AvailabilityZone   string      `json:"availabilityZone"`
	DevpayproductCodes interface{} `json:"devpayProductCodes"`
	PrivateIP          string      `json:"privateIp"`
}

func (driver *Driver) GetBlockDeviceMapping() (interface{}, error) {
	blockDevices, err := driver.getBlockDevices(driver.InstanceDocument.InstanceID)
	if err != nil {
		return nil, err
	}

	var BlockDevices []*storagedriver.BlockDevice
	for _, blockDevice := range blockDevices {
		sdBlockDevice := &storagedriver.BlockDevice{
			ProviderName: providerName,
			InstanceID:   driver.InstanceDocument.InstanceID,
			Region:       driver.InstanceDocument.Region,
			DeviceName:   blockDevice.DeviceName,
			VolumeID:     blockDevice.EBS.VolumeId,
			Status:       blockDevice.EBS.Status,
		}
		BlockDevices = append(BlockDevices, sdBlockDevice)
	}

	return BlockDevices, nil
}

func getInstanceIdendityDocument() (*instanceIdentityDocument, error) {
	conn, err := net.DialTimeout("tcp", "169.254.169.254:80", 50*time.Millisecond)
	if err != nil {
		return &instanceIdentityDocument{}, fmt.Errorf("Error: %v\n", err)
	}
	defer conn.Close()

	url := "http://169.254.169.254/latest/dynamic/instance-identity/document"
	resp, err := http.Get(url)
	if err != nil {
		return &instanceIdentityDocument{}, fmt.Errorf("Error: %v\n", err)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &instanceIdentityDocument{}, fmt.Errorf("Error: %v\n", err)
	}

	var document instanceIdentityDocument
	err = json.Unmarshal(data, &document)
	if err != nil {
		return &instanceIdentityDocument{}, fmt.Errorf("Error: %v\n", err)
	}

	return &document, nil
}

func (driver *Driver) getBlockDevices(instanceID string) ([]ec2.BlockDevice, error) {

	instance, err := driver.getInstance()
	if err != nil {
		return []ec2.BlockDevice{}, err
	}

	return instance.BlockDevices, nil

}

func getInstanceName(server ec2.Instance) string {
	return getTag(server, "Name")
}

func getTag(server ec2.Instance, key string) string {
	for _, tag := range server.Tags {
		if tag.Key == key {
			return tag.Value
		}
	}
	return ""
}

func (driver *Driver) GetInstance() (interface{}, error) {

	server, err := driver.getInstance()
	if err != nil {
		return storagedriver.Instance{}, err
	}

	instance := &storagedriver.Instance{
		ProviderName: providerName,
		InstanceID:   driver.InstanceDocument.InstanceID,
		Region:       driver.InstanceDocument.Region,
		Name:         getInstanceName(server),
	}

	return instance, nil
}

func (driver *Driver) getInstance() (ec2.Instance, error) {

	resp, err := driver.EC2Instance.DescribeInstances([]string{driver.InstanceDocument.InstanceID}, &ec2.Filter{})
	if err != nil {
		return ec2.Instance{}, err
	}

	return resp.Reservations[0].Instances[0], nil

}
