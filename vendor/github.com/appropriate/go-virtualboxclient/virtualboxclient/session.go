package virtualboxclient

import "github.com/appropriate/go-virtualboxclient/vboxwebsrv"

type Session struct {
	virtualbox      *VirtualBox
	managedObjectId string
}

func (s *Session) UnlockMachine() error {
	request := vboxwebsrv.ISessionunlockMachine{This: s.managedObjectId}
	_, err := s.virtualbox.ISessionunlockMachine(&request)
	if err != nil {
		return err // TODO: Wrap the error
	}

	// TODO: See if we need to do anything with the response
	return nil
}

func (s *Session) LockMachine(m *Machine, l vboxwebsrv.LockType) error {
	request := vboxwebsrv.IMachinelockMachine{
		This:     m.managedObjectId,
		Session:  s.managedObjectId,
		LockType: &l,
	}
	_, err := s.virtualbox.IMachinelockMachine(&request)
	if err != nil {
		return err // TODO: Wrap the error
	}

	// TODO: See if we need to do anything with the response
	return nil
}

func (s *Session) GetMachine() (*Machine, error) {
	request := vboxwebsrv.ISessiongetMachine{This: s.managedObjectId}
	response, err := s.virtualbox.ISessiongetMachine(&request)
	if err != nil {
		return nil, err // TODO: Wrap the error
	}

	// TODO: See if we need to do anything with the response
	return &Machine{managedObjectId: response.Returnval, virtualbox: s.virtualbox}, nil
}

func (s *Session) Release() error {
	return s.virtualbox.Release(s.managedObjectId)
}
