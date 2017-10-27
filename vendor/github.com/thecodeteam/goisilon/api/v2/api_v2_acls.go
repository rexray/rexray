package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/thecodeteam/goisilon/api"
	"context"
)

// AuthoritativeType is a possible value used with an ACL's Authoritative field.
type AuthoritativeType uint8

const (
	// AuthoritativeTypeUnknown is an unknown AuthoritativeType.
	AuthoritativeTypeUnknown AuthoritativeType = iota

	// AuthoritativeTypeACL sets an ACL's Authoritative field to "acl".
	AuthoritativeTypeACL

	// AuthoritativeTypeMode sets an ACL's Authoritative field to "mode".
	AuthoritativeTypeMode

	authoritativeTypeCount
)

var (
	// PAuthoritativeTypeACL is used to grab a pointer to a const.
	PAuthoritativeTypeACL = AuthoritativeTypeACL

	// PAuthoritativeTypeMode is used to grab a pointer to a const.
	PAuthoritativeTypeMode = AuthoritativeTypeMode
)

const (
	authoritativeTypeUnknownStr = "unknown"
	authoritativeTypeACLStr     = "acl"
	authoritativeTypeModeStr    = "mode"
)

var authoritativeTypesToStrs = [authoritativeTypeCount]string{
	authoritativeTypeUnknownStr,
	authoritativeTypeACLStr,
	authoritativeTypeModeStr,
}

// ParseAuthoritativeType parses an AuthoritativeType from a string.
func ParseAuthoritativeType(text string) AuthoritativeType {
	switch {
	case strings.EqualFold(text, authoritativeTypeACLStr):
		return AuthoritativeTypeACL
	case strings.EqualFold(text, authoritativeTypeModeStr):
		return AuthoritativeTypeMode
	}
	return AuthoritativeTypeUnknown
}

// String returns the string representation of an AuthoritativeType value.
func (p AuthoritativeType) String() string {
	if p < (AuthoritativeTypeUnknown+1) || p >= authoritativeTypeCount {
		return authoritativeTypesToStrs[AuthoritativeTypeUnknown]
	}
	return authoritativeTypesToStrs[p]
}

// MarshalJSON marshals an AuthoritativeType value to JSON.
func (p AuthoritativeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON unmarshals an AuthoritativeType value from JSON.
func (p *AuthoritativeType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = ParseAuthoritativeType(s)
	return nil
}

// ActionType is a possible value used with an ACL's Action field.
type ActionType uint8

const (
	// ActionTypeUnknown is an unknown ActionType.
	ActionTypeUnknown ActionType = iota

	// ActionTypeReplace sets an ACL's Action field to "replace".
	ActionTypeReplace

	// ActionTypeUpdate sets an ACL's Action field to "update".
	ActionTypeUpdate

	ActionTypeCount
)

var (
	// PActionTypeReplace is used to grab a pointer to a const.
	PActionTypeReplace = ActionTypeReplace

	// PActionTypeUpdate is used to grab a pointer to a const.
	PActionTypeUpdate = ActionTypeUpdate
)

const (
	ActionTypeUnknownStr = "unknown"
	ActionTypeReplaceStr = "replace"
	ActionTypeUpdateStr  = "update"
)

var ActionTypesToStrs = [ActionTypeCount]string{
	ActionTypeUnknownStr,
	ActionTypeReplaceStr,
	ActionTypeUpdateStr,
}

// ParseActionType parses an ActionType from a string.
func ParseActionType(text string) ActionType {
	switch {
	case strings.EqualFold(text, ActionTypeReplaceStr):
		return ActionTypeReplace
	case strings.EqualFold(text, ActionTypeUpdateStr):
		return ActionTypeUpdate
	}
	return ActionTypeUnknown
}

// String returns the string representation of an ActionType value.
func (p ActionType) String() string {
	if p < (ActionTypeUnknown+1) || p >= ActionTypeCount {
		return ActionTypesToStrs[ActionTypeUnknown]
	}
	return ActionTypesToStrs[p]
}

// MarshalJSON marshals an ActionType value to JSON.
func (p ActionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON unmarshals an ActionType value from JSON.
func (p *ActionType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = ParseActionType(s)
	return nil
}

// FileMode is an alias for the Go os.FileMode.
type FileMode os.FileMode

// String returns the string representation of an ActionType value.
func (p FileMode) String() string {
	return fmt.Sprintf("%04o", p)
}

// MarshalText marshals a FileMode value to text.
func (p FileMode) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

var invalidFileMode = errors.New("invalid file mode")

// UnmarshalText unmarshals a FileMode value from text.
func (p *FileMode) UnmarshalText(data []byte) error {
	if len(data) < 3 {
		return invalidFileMode
	}
	if len(data) == 3 {
		data = []byte{'0', data[0], data[1], data[2]}
	}
	m, err := strconv.ParseUint(string(data), 8, 32)
	if err != nil {
		return err
	}
	*p = FileMode(m)
	return nil
}

// ParseFileMode parses a string and returns a FileMode.
func ParseFileMode(s string) (FileMode, error) {
	var fm FileMode
	if err := fm.UnmarshalText([]byte(s)); err != nil {
		return 0, err
	}
	return fm, nil
}

// ACL is an Isilon Access Control List used for managing an object's security.
type ACL struct {
	Authoritative *AuthoritativeType `json:"authoritative,omitempty"`
	Action        *ActionType        `json:"action,omitempty"`
	Owner         *Persona           `json:"owner,omitempty"`
	Group         *Persona           `json:"group,omitempty"`
	Mode          *FileMode          `json:"mode,omitempty"`
}

var aclQueryString = api.OrderedValues{{[]byte("acl")}}

// ACLInspect GETs an ACL.
func ACLInspect(
	ctx context.Context,
	client api.Client,
	path string) (*ACL, error) {

	var resp ACL

	if err := client.Get(
		ctx,
		realNamespacePath(client),
		path,
		aclQueryString,
		nil,
		&resp); err != nil {

		return nil, err
	}

	return &resp, nil
}

// ACLUpdate PUTs an ACL.
func ACLUpdate(
	ctx context.Context,
	client api.Client,
	path string,
	acl *ACL) error {

	if err := client.Put(
		ctx,
		realNamespacePath(client),
		path,
		aclQueryString,
		nil,
		acl,
		nil); err != nil {

		return err
	}

	return nil
}
