package config

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/emccode/rexray/core/errors"
)

// Registration is used to register configuration information with the config
// package.
type Registration struct {
	name string
	yaml string
	keys []*regKey
}

// KeyType is a config registration key type.
type KeyType int

const (
	// String is a key with a string value
	String KeyType = iota

	// Int is a key with an integer value
	Int

	// Bool is a key with a boolean value
	Bool
)

type regKey struct {
	keyType    KeyType
	defVal     interface{}
	short      string
	desc       string
	keyName    string
	flagName   string
	envVarName string
}

// NewRegistration creates a new registration with the given name.
func NewRegistration(name string) *Registration {
	return &Registration{
		name: name,
		keys: []*regKey{},
	}
}

// Yaml sets the registration's default yaml configuration.
func (r *Registration) Yaml(y string) {
	r.yaml = y
}

// Key adds a key to the registration.
//
// The first vararg argument is the yaml name of the key, using a '.' as
// the nested separator. If the second two arguments are omitted they will be
// generated from the first argument. The second argument is the explicit name
// of the flag bound to this key. The third argument is the explicit name of
// the environment variable bound to thie key.
func (r *Registration) Key(
	keyType KeyType,
	short string,
	defVal interface{},
	description string,
	keys ...string) {

	if keys == nil {
		panic(errors.New("keys is nil"))
	}

	lk := len(keys)

	if lk == 0 {
		panic(errors.New("len(keys) == 0"))
	}

	kn := keys[0]

	rk := &regKey{
		keyType: keyType,
		short:   short,
		desc:    description,
		defVal:  defVal,
		keyName: keys[0],
	}

	if lk < 2 {
		kp := strings.Split(kn, ".")
		for x, s := range kp {
			if x == 0 {
				var buff []byte
				b := bytes.NewBuffer(buff)
				for y, r := range s {
					if y == 0 {
						b.WriteRune(unicode.ToLower(r))
					} else {
						b.WriteRune(r)
					}
				}
				kp[x] = b.String()
			} else {
				kp[x] = strings.Title(s)
			}
		}
		rk.flagName = strings.Join(kp, "")
	}

	if lk < 3 {
		kp := strings.Split(kn, ".")
		for x, s := range kp {
			kp[x] = strings.ToUpper(s)
		}
		rk.envVarName = strings.Join(kp, "_")
	}

	r.keys = append(r.keys, rk)
}
