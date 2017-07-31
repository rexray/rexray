package virtualboxclient

import (
	"github.com/appropriate/go-virtualboxclient/vboxwebsrv"
)

type SystemProperties struct {
	virtualbox      *VirtualBox
	managedObjectId string
}

func (sp *SystemProperties) GetMaxNetworkAdapters(chipset *vboxwebsrv.ChipsetType) (uint32, error) {
	request := vboxwebsrv.ISystemPropertiesgetMaxNetworkAdapters{This: sp.managedObjectId, Chipset: chipset}

	response, err := sp.virtualbox.ISystemPropertiesgetMaxNetworkAdapters(&request)
	if err != nil {
		return 0, err // TODO: Wrap the error
	}

	return response.Returnval, nil
}

func (sp *SystemProperties) GetMaxDevicesPerPortForStorageBus(bus vboxwebsrv.StorageBus) (uint32, error) {
	request := vboxwebsrv.ISystemPropertiesgetMaxDevicesPerPortForStorageBus{This: sp.managedObjectId, Bus: &bus}
	response, err := sp.virtualbox.ISystemPropertiesgetMaxDevicesPerPortForStorageBus(&request)
	if err != nil {
		return 0, err // TODO: Wrap the error
	}

	return response.Returnval, nil
}

func (sp *SystemProperties) GetMinPortCountForStorageBus(bus vboxwebsrv.StorageBus) (uint32, error) {
	request := vboxwebsrv.ISystemPropertiesgetMinPortCountForStorageBus{This: sp.managedObjectId, Bus: &bus}
	response, err := sp.virtualbox.ISystemPropertiesgetMinPortCountForStorageBus(&request)
	if err != nil {
		return 0, err // TODO: Wrap the error
	}

	return response.Returnval, nil
}

func (sp *SystemProperties) Release() error {
	return sp.virtualbox.Release(sp.managedObjectId)
}
