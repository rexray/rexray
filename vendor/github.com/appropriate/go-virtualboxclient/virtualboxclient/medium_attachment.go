package virtualboxclient

import "github.com/appropriate/go-virtualboxclient/vboxwebsrv"

type MediumAttachment struct {
	*vboxwebsrv.IMediumAttachment
	virtualbox      *VirtualBox
	managedObjectId string
}

func (m *MediumAttachment) GetMedium() (*Medium, error) {
	return &Medium{virtualbox: m.virtualbox, managedObjectId: m.Medium}, nil
}
