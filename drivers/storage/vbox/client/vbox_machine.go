package client

// Machine represents an installed virtual machine in vbox.
type Machine struct {
	mobref      string
	id          string
	name        string
	attachments []*MediumAttachment
	vb          *VirtualBox
}

//MediumAttachment represents attached devices to machine
type MediumAttachment struct {
	Medium     string
	Controller string
	Port       int32
	Device     int32
	Type       string
}

// NewMachine returns a pointer to a Machine value
func NewMachine(vb *VirtualBox, id string) *Machine {
	return &Machine{vb: vb, id: id}
}

// GetID returns the ID last populated for this machine
func (m *Machine) GetID() string {
	return m.id
}

// GetName returns the Name last populated for this machine
func (m *Machine) GetName() string {
	return m.name
}

// GetMediumAttachments returns the attached media to machine
func (m *Machine) GetMediumAttachments() []*MediumAttachment {
	return m.attachments
}
