package virtualboxclient

import (
	"github.com/appropriate/go-virtualboxclient/vboxwebsrv"
)

type NetworkAdapter struct {
	virtualbox      *VirtualBox
	managedObjectId string
}

func (na *NetworkAdapter) GetMACAddress() (string, error) {
	request := vboxwebsrv.INetworkAdaptergetMACAddress{This: na.managedObjectId}

	response, err := na.virtualbox.INetworkAdaptergetMACAddress(&request)
	if err != nil {
		return "", err // TODO: Wrap the error
	}

	return response.Returnval, nil
}
