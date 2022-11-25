package utils

import (
	"fmt"

	"github.com/rexray/rexray/libstorage/api/types"
	"github.com/rexray/rexray/libstorage/drivers/storage/ebs"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
)

// IsEC2Instance returns a flag indicating whether the executing host is an EC2
// instance based on whether or not the metadata URL can be accessed.
func IsEC2Instance(ctx types.Context) (bool, error) {
	sess, err := session.NewSession()
        if err != nil {
            return false, err
        }
	_, err = ec2metadata.New(sess).GetMetadata("instance-id")
	if err != nil {
		return false, err
	}
	return true, nil
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

	sess, err := session.NewSession()
	if err != nil {
	    return nil, err
	}
	iid, err := ec2metadata.New(sess).GetInstanceIdentityDocument()
	if err != nil {
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

	sess, err := session.NewSession()
	if err != nil {
	    return nil, err
	}
	buf, err := ec2metadata.New(sess).GetMetadata("block-device-mapping/")
	if err != nil {
		return nil, err
	}
	return []byte(buf), nil
}

// BlockDeviceName returns the name of the provided EBS device.
func BlockDeviceName(
	ctx types.Context,
	device string) ([]byte, error) {

	sess, err := session.NewSession()
	if err != nil {
	    return nil, err
	}
	buf, err := ec2metadata.New(sess).GetMetadata(fmt.Sprintf("block-device-mapping/%s", device))
	if err != nil {
		return nil, err
	}
	return []byte(buf), nil
}
