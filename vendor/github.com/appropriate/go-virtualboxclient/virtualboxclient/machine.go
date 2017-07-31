package virtualboxclient

import (
	"errors"

	"github.com/appropriate/go-virtualboxclient/vboxwebsrv"
)

type Machine struct {
	virtualbox      *VirtualBox
	managedObjectId string
	ID              string
	Name            string
}

func (m *Machine) GetChipsetType() (*vboxwebsrv.ChipsetType, error) {
	request := vboxwebsrv.IMachinegetChipsetType{This: m.managedObjectId}

	response, err := m.virtualbox.IMachinegetChipsetType(&request)
	if err != nil {
		return nil, err // TODO: Wrap the error
	}

	return response.Returnval, nil
}

func (m *Machine) GetMediumAttachments() ([]*vboxwebsrv.IMediumAttachment, error) {
	request := vboxwebsrv.IMachinegetMediumAttachments{This: m.managedObjectId}

	response, err := m.virtualbox.IMachinegetMediumAttachments(&request)
	if err != nil {
		return nil, err // TODO: Wrap the error
	}

	ret := response.Returnval
	return ret, nil
}

func (m *Machine) GetMediumAttachmentsOfController(cName string) ([]*vboxwebsrv.IMediumAttachment, error) {
	request := vboxwebsrv.IMachinegetMediumAttachmentsOfController{This: m.managedObjectId, Name: cName}

	response, err := m.virtualbox.IMachinegetMediumAttachmentsOfController(&request)
	if err != nil {
		return nil, err // TODO: Wrap the error
	}

	return response.Returnval, nil
}

func (m *Machine) GetNetworkAdapter(slot uint32) (*NetworkAdapter, error) {
	request := vboxwebsrv.IMachinegetNetworkAdapter{This: m.managedObjectId, Slot: slot}

	response, err := m.virtualbox.IMachinegetNetworkAdapter(&request)
	if err != nil {
		return nil, err // TODO: Wrap the error
	}

	return &NetworkAdapter{m.virtualbox, response.Returnval}, nil
}

func (m *Machine) GetSettingsFilePath() (string, error) {
	request := vboxwebsrv.IMachinegetSettingsFilePath{This: m.managedObjectId}

	response, err := m.virtualbox.IMachinegetSettingsFilePath(&request)
	if err != nil {
		return "", err // TODO: Wrap the error
	}

	return response.Returnval, nil
}

func (m *Machine) SaveSettings() error {
	request := vboxwebsrv.IMachinesaveSettings{This: m.managedObjectId}

	_, err := m.virtualbox.IMachinesaveSettings(&request)
	if err != nil {
		defer m.DiscardSettings()
		return err // TODO: Wrap the error
	}

	return nil
}

func (m *Machine) DiscardSettings() error {
	request := vboxwebsrv.IMachinediscardSettings{This: m.managedObjectId}

	_, err := m.virtualbox.IMachinediscardSettings(&request)
	if err != nil {
		return err // TODO: Wrap the error
	}

	return nil
}

func (m *Machine) GetStorageControllers() ([]*StorageController, error) {
	request := vboxwebsrv.IMachinegetStorageControllers{This: m.managedObjectId}

	response, err := m.virtualbox.IMachinegetStorageControllers(&request)
	if err != nil {
		return nil, err // TODO: Wrap the error
	}

	storageControllers := make([]*StorageController, len(response.Returnval))
	for i, oid := range response.Returnval {
		storageControllers[i] = &StorageController{virtualbox: m.virtualbox, managedObjectId: oid}
	}

	return storageControllers, nil
}

func (m *Machine) GetStorageController(name string) (*StorageController, error) {
	if name == "" {
		return nil, errors.New("storage controller name not specified")
	}
	scs, err := m.GetStorageControllers()
	if err != nil {
		return nil, err
	}

	for _, sc := range scs {
		scName, err := sc.GetName()
		if err != nil {
			return nil, err
		}
		if scName == name {
			sc.Name = scName
			return sc, nil
		}
	}
	return nil, errors.New("storage controller not found")
}

func (m *Machine) AttachDevice(medium *Medium) error {
	session, err := m.virtualbox.GetSession()
	if err != nil {
		return err
	}
	// defer session.Release()

	if err := m.Lock(session, vboxwebsrv.LockTypeShared); err != nil {
		return err
	}
	defer m.Unlock(session)

	sm, err := session.GetMachine()
	if err != nil {
		return err
	}
	defer sm.Release()

	if m.virtualbox.controllerName == "" {
		return errors.New("missing controllerName")
	}

	sc, err := sm.GetStorageController(m.virtualbox.controllerName)
	if err != nil {
		return err
	}

	pn, err := sc.GetNextAvailablePort(m)
	if err != nil {
		return err
	}

	request := vboxwebsrv.IMachineattachDevice{
		This:           sm.managedObjectId,
		Name:           sc.Name,
		ControllerPort: pn,
		Device:         0,
		Type_:          &medium.DeviceType,
		Medium:         medium.managedObjectId,
	}

	_, err = m.virtualbox.IMachineattachDevice(&request)
	if err != nil {
		return err
	}

	if err := sm.SaveSettings(); err != nil {
		return err
	}

	return nil
}

func (m *Machine) DetachDevice(medium *Medium) error {

	session, err := m.virtualbox.GetSession()
	if err != nil {
		return err
	}
	// defer session.Release()

	if err := m.Lock(session, vboxwebsrv.LockTypeShared); err != nil {
		return err
	}
	defer m.Unlock(session)

	sm, err := session.GetMachine()
	if err != nil {
		return err
	}
	defer sm.Release()

	mediumAttachments, err := m.GetMediumAttachments()
	if err != nil {
		return err
	}

	var request *vboxwebsrv.IMachinedetachDevice
	for _, ma := range mediumAttachments {
		am := &Medium{virtualbox: m.virtualbox, managedObjectId: ma.Medium}
		defer am.Release()
		amID, err := am.GetID()
		if err != nil {
			return err
		}

		if amID != medium.ID {
			continue
		}
		request = &vboxwebsrv.IMachinedetachDevice{
			This:           sm.managedObjectId,
			Name:           ma.Controller,
			ControllerPort: ma.Port,
			Device:         0,
		}
	}
	if request == nil {
		return errors.New("couldn't find attached medium")
	}

	_, err = m.virtualbox.IMachinedetachDevice(request)
	if err != nil {
		return err
	}

	if err := sm.SaveSettings(); err != nil {
		return err
	}

	return nil
}

func (m *Machine) Unlock(session *Session) error {
	if err := session.UnlockMachine(); err != nil {
		return err
	}
	return nil
}

func (m *Machine) Lock(session *Session, lockType vboxwebsrv.LockType) error {
	if err := session.LockMachine(m, lockType); err != nil {
		return err
	}
	return nil
}

func (m *Machine) GetID() (string, error) {
	request := vboxwebsrv.IMachinegetId{This: m.managedObjectId}

	response, err := m.virtualbox.IMachinegetId(&request)
	if err != nil {
		return "", err // TODO: Wrap the error
	}

	// TODO: See if we need to do anything with the response
	return response.Returnval, nil
}

func (m *Machine) GetName() (string, error) {
	request := vboxwebsrv.IMachinegetName{This: m.managedObjectId}

	response, err := m.virtualbox.IMachinegetName(&request)
	if err != nil {
		return "", err // TODO: Wrap the error
	}

	// TODO: See if we need to do anything with the response
	return response.Returnval, nil
}

func (m *Machine) Release() error {
	return m.virtualbox.Release(m.managedObjectId)
}

func (m *Machine) Refresh() error {
	if mr, err := m.virtualbox.FindMachine(m.ID); err != nil {
		return err
	} else {
		m.managedObjectId = mr.managedObjectId
	}
	return nil
}
