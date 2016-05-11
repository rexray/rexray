package types

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/akutz/goof"
)

// InstanceIDMap is a map of InstanceID objects.
type InstanceIDMap map[string]*InstanceID

// InstanceID identifies a host to a remote storage platform.
type InstanceID struct {
	// ID is the simple part of the InstanceID.
	ID string `json:"id"`

	// Driver is the name of the StorageExecutor that created the InstanceID
	// as well as the name of the StorageDriver for which the InstanceID is
	// valid.
	Driver string `json:"driver"`

	// Formatted is a flag indicating whether or not the InstanceID has
	// been formatted by an instance inspection.
	Formatted bool `json:"formatted,omitempty"`

	// Metadata is any extra information about the InstanceID.
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// String returns the string representation of an InstanceID object.
func (i *InstanceID) String() string {
	buf, err := i.MarshalText()
	if err != nil {
		panic(err)
	}
	return string(buf)
}

// AddMetadata encodes the provided data to JSON and assigns the result to
// the InstanceID's Metadata field.
func (i *InstanceID) AddMetadata(data interface{}) error {

	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	i.Metadata = buf
	return nil
}

// MarshalText marshals InstanceID to a text string that adheres to the format
// `DRIVER=ID[,FORMATTED][,METADATA]`. The formatted flag may be omitted if
// it is false or if there is no metadata; otherwise the flag is included
// regardless of its value. If metadata is present it is encoded as a base64
// string.
func (i *InstanceID) MarshalText() ([]byte, error) {

	t := &bytes.Buffer{}
	fmt.Fprintf(t, "%s=%s", i.Driver, i.ID)

	hasMetadata := len(i.Metadata) > 0

	if i.Formatted || hasMetadata {
		fmt.Fprintf(t, ",%v", i.Formatted)
	}

	if !hasMetadata {
		return t.Bytes(), nil
	}

	t.WriteByte(',')

	enc := b64.NewEncoder(b64.StdEncoding, t)
	if _, err := enc.Write(i.Metadata); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return t.Bytes(), nil
}

var iidRX = regexp.MustCompile(
	`(?i)^([^=]+)=([^,]*)(?:,(true|false|0|1)(?:,(.+))?)?$`)

// UnmarshalText unmarshals the data into a an InstanceID provided the data
// adheres to the format described in the MarshalText function.
func (i *InstanceID) UnmarshalText(value []byte) error {

	m := iidRX.FindSubmatch(value)
	lm := len(m)

	if lm < 3 {
		return goof.WithField("value", string(value), "invalid InstanceID")
	}

	i.Driver = string(m[1])
	i.ID = string(m[2])

	if lm > 3 && len(m[3]) > 0 {
		i.Formatted, _ = strconv.ParseBool(string(m[3]))
	}

	if lm > 4 && len(m[4]) > 0 {
		enc := m[4]
		dec := b64.NewDecoder(b64.StdEncoding, bytes.NewReader(enc))
		i.Metadata = make([]byte, b64.StdEncoding.DecodedLen(len(enc)))
		n, err := dec.Read(i.Metadata)
		if err != nil {
			return err
		}
		i.Metadata = i.Metadata[:n]
	}

	return nil
}

// MarshalJSON marshals the InstanceID to JSON.
func (i *InstanceID) MarshalJSON() ([]byte, error) {

	return json.Marshal(&struct {
		ID        string          `json:"id"`
		Driver    string          `json:"driver"`
		Formatted bool            `json:"formatted,omitempty"`
		Metadata  json.RawMessage `json:"metadata,omitempty"`
	}{i.ID, i.Driver, i.Formatted, i.Metadata})

}

// UnmarshalJSON marshals the InstanceID to JSON.
func (i *InstanceID) UnmarshalJSON(data []byte) error {

	iid := &struct {
		ID        string          `json:"id"`
		Driver    string          `json:"driver"`
		Formatted bool            `json:"formatted,omitempty"`
		Metadata  json.RawMessage `json:"metadata,omitempty"`
	}{}

	if err := json.Unmarshal(data, iid); err != nil {
		return err
	}

	i.ID = iid.ID
	i.Driver = iid.Driver
	i.Formatted = iid.Formatted
	i.Metadata = iid.Metadata

	return nil
}
