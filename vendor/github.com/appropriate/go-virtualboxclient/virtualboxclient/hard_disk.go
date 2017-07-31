package virtualboxclient

type HardDisk struct {
	virtualbox      *VirtualBox
	managedObjectId string
}

type HardDisks struct {
	disks []*HardDisk
}

func (h *HardDisk) getMedium() *Medium {
	return &Medium{virtualbox: h.virtualbox, managedObjectId: h.managedObjectId}
}

func isSet(value string) bool {
	return value != ""
}

func (hs *HardDisks) GetMedium(objectID, name string) ([]*Medium, error) {
	var ms []*Medium
	for _, hardDisk := range hs.disks {
		om := hardDisk.getMedium()
		var m *Medium
		if isSet(name) || isSet(objectID) {
			var err error
			m, err = om.GetIDName()
			if err != nil {
				return nil, err
			}
		}

		if isSet(name) && m.Name != name {
			continue
		}

		if isSet(objectID) && m.ID != objectID {
			continue
		}

		medium, err := om.Get()
		if err != nil {
			return nil, err
		}
		ms = append(ms, medium)
	}

	return ms, nil
}
