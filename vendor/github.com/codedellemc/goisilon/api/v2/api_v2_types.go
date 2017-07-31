package v2

import (
	"encoding/json"
	"fmt"
	"strings"
)

// UserMapping maps to the ISI <user-mapping> type.
type UserMapping struct {
	Enabled        *bool      `json:"enabled,omitempty"`
	User           *Persona   `json:"user,omitempty"`
	PrimaryGroup   *Persona   `json:"primary_group,omitempty"`
	SecondaryGroup []*Persona `json:"secondary_group,omitempty"`
}

type userMapping struct {
	Enabled        *bool      `json:"enabled,omitempty"`
	User           *Persona   `json:"user,omitempty"`
	PrimaryGroup   *Persona   `json:"primary_group,omitempty"`
	SecondaryGroup []*Persona `json:"secondary_group,omitempty"`
}

func isNilPersona(p *Persona) bool {
	return p == nil || (p.ID == nil && p.Name == nil && p.Type == nil)
}

// UnmarshalJSON unmarshals a UserMapping from JSON.
func (um *UserMapping) UnmarshalJSON(data []byte) error {

	if isEmptyJSON(&data) {
		return nil
	}

	var pum userMapping
	if err := json.Unmarshal(data, &pum); err != nil {
		return nil
	}

	if pum.Enabled != nil {
		um.Enabled = pum.Enabled
	}
	if !isNilPersona(pum.User) {
		um.User = pum.User
	}
	if !isNilPersona(pum.PrimaryGroup) {
		um.PrimaryGroup = pum.PrimaryGroup
	}
	if len(pum.SecondaryGroup) > 0 {
		um.SecondaryGroup = pum.SecondaryGroup
	}

	return nil
}

func isEmptyJSON(data *[]byte) bool {
	d := *data
	return len(d) == 2 && d[0] == '{' && d[1] == '}'
}

// Persona maps to the ISI <persona> type.
type Persona struct {
	ID   *PersonaID   `json:"id,omitempty"`
	Type *PersonaType `json:"type,omitempty"`
	Name *string      `json:"name,omitempty"`
}

type persona struct {
	ID   *PersonaID   `json:"id,omitempty"`
	Type *PersonaType `json:"type,omitempty"`
	Name *string      `json:"name,omitempty"`
}

type personaWithID struct {
	ID *PersonaID `json:"id,omitempty"`
}

// MarshalJSON marshals a Persona to JSON.
func (p *Persona) MarshalJSON() ([]byte, error) {
	if p.ID != nil {
		return json.Marshal(personaWithID{p.ID})
	} else if p.Type != nil && p.Name != nil {
		return json.Marshal(fmt.Sprintf("%s:%s", *p.Type, *p.Name))
	} else if p.Name != nil {
		return json.Marshal(*p.Name)
	}
	return nil, fmt.Errorf("persona cannot be marshaled to json: %+v", p)
}

// UnmarshalJSON unmarshals a Persona from JSON.
func (p *Persona) UnmarshalJSON(data []byte) error {

	if isEmptyJSON(&data) {
		return nil
	}

	var pp persona
	if err := json.Unmarshal(data, &pp); err == nil {
		p.ID = pp.ID
		p.Name = pp.Name
		p.Type = pp.Type
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 1 {
		p.Name = &parts[0]
		return nil
	}

	pt := ParsePersonaType(parts[0])
	p.Type = &pt
	p.Name = &parts[1]
	return nil
}

// PersonaID maps to the ISI <persona-id> type.
type PersonaID struct {
	ID   string
	Type PersonaIDType
}

// MarshalJSON marshals a PersonaID to JSON.
func (p *PersonaID) MarshalJSON() ([]byte, error) {
	if p.Type == PersonaIDTypeUnknown {
		return json.Marshal(p.ID)
	}
	return json.Marshal(fmt.Sprintf("%s:%s", p.Type, p.ID))
}

// UnmarshalJSON unmarshals a PersonaID from JSON.
func (p *PersonaID) UnmarshalJSON(data []byte) error {

	if isEmptyJSON(&data) {
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 1 {
		p.ID = parts[0]
		return nil
	}

	p.Type = ParsePersonaIDType(parts[0])
	p.ID = parts[1]
	return nil
}

// PersonaIDType is a valid Persona ID type.
type PersonaIDType uint8

const (
	// PersonaIDTypeUnknown is an unknown PersonaID type.
	PersonaIDTypeUnknown PersonaIDType = iota

	// PersonaIDTypeUser is a PersonaID user type.
	PersonaIDTypeUser

	// PersonaIDTypeGroup is a PersonaID group type.
	PersonaIDTypeGroup

	// PersonaIDTypeSID is a PersonaID SID type.
	PersonaIDTypeSID

	// PersonaIDTypeUID is a PersonaID UID type.
	PersonaIDTypeUID

	// PersonaIDTypeGID is a PersonaID GID type.
	PersonaIDTypeGID

	personaIDTypeCount
)

const (
	personaIDTypeUnknownStr = "unknown"
	personaIDTypeUserStr    = "user"
	personaIDTypeGroupStr   = "group"
	personaIDTypeSIDStr     = "SID"
	personaIDTypeUIDStr     = "UID"
	personaIDTypeGIDStr     = "GID"
)

var personaIDTypesToStrs = [personaIDTypeCount]string{
	personaIDTypeUnknownStr,
	personaIDTypeUserStr,
	personaIDTypeGroupStr,
	personaIDTypeSIDStr,
	personaIDTypeUIDStr,
	personaIDTypeGIDStr,
}

// ParsePersonaIDType parses a PersonaIDType from a string.
func ParsePersonaIDType(text string) PersonaIDType {
	switch {
	case strings.EqualFold(text, personaIDTypeUserStr):
		return PersonaIDTypeUser
	case strings.EqualFold(text, personaIDTypeGroupStr):
		return PersonaIDTypeGroup
	case strings.EqualFold(text, personaIDTypeSIDStr):
		return PersonaIDTypeSID
	case strings.EqualFold(text, personaIDTypeUIDStr):
		return PersonaIDTypeUID
	case strings.EqualFold(text, personaIDTypeGIDStr):
		return PersonaIDTypeGID
	}
	return PersonaIDTypeUnknown
}

// String returns the string representation of a PersonaIDType value.
func (p PersonaIDType) String() string {
	if p < (PersonaIDTypeUnknown+1) || p >= personaIDTypeCount {
		return personaIDTypesToStrs[PersonaIDTypeUnknown]
	}
	return personaIDTypesToStrs[p]
}

// MarshalJSON marshals a PersonaIDType value to JSON.
func (p PersonaIDType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON unmarshals a PersonaIDType value from JSON.
func (p *PersonaIDType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = ParsePersonaIDType(s)
	return nil
}

// PersonaType is a valid Persona type.
type PersonaType uint8

const (
	// PersonaTypeUnknown is an unknown Persona type.
	PersonaTypeUnknown PersonaType = iota

	// PersonaIDTypeUser is a Persona user type.
	PersonaTypeUser

	// PersonaTypeGroup is a Persona group type.
	PersonaTypeGroup

	// PersonaTypeWellKnown is a Persona wellknown type.
	PersonaTypeWellKnown

	personaTypeCount
)

var (
	// PPersonaIDTypeUnknown is used to get adddress of the constant.
	PPersonaTypeUnknown = PersonaTypeUnknown

	// PPersonaTypeUser is used to get adddress of the constant.
	PPersonaTypeUser = PersonaTypeUser

	// PPersonaTypeGroup is used to get adddress of the constant.
	PPersonaTypeGroup = PersonaTypeGroup

	// PPersonaTypeWellKnown is used to get adddress of the constant.
	PPersonaTypeWellKnown = PersonaTypeWellKnown
)

const (
	personaTypeUnknownStr   = "unknown"
	personaTypeUserStr      = "user"
	personaTypeGroupStr     = "group"
	personaTypeWellKnownStr = "wellknown"
)

var personaTypesToStrs = [personaTypeCount]string{
	personaTypeUnknownStr,
	personaTypeUserStr,
	personaTypeGroupStr,
	personaTypeWellKnownStr,
}

// ParsePersonaType parses a PersonaType from a string.
func ParsePersonaType(text string) PersonaType {
	switch {
	case strings.EqualFold(text, personaTypeUserStr):
		return PersonaTypeUser
	case strings.EqualFold(text, personaTypeGroupStr):
		return PersonaTypeGroup
	case strings.EqualFold(text, personaTypeWellKnownStr):
		return PersonaTypeWellKnown
	}
	return PersonaTypeUnknown
}

// String returns the string representation of a PersonaType value.
func (p PersonaType) String() string {
	if p < (PersonaTypeUnknown+1) || p >= personaTypeCount {
		return personaTypesToStrs[PersonaTypeUnknown]
	}
	return personaTypesToStrs[p]
}

// MarshalJSON marshals a PersonaType value to JSON.
func (p PersonaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON marshals a PersonaType value from JSON.
func (p *PersonaType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = ParsePersonaType(s)
	return nil
}
