package gofig

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/akutz/gofig/types"
	"github.com/akutz/goof"
	log "github.com/sirupsen/logrus"
)

type configReg struct {
	name string
	yaml string
	keys []types.ConfigRegistrationKey
}

type configRegKey struct {
	keyType    types.ConfigKeyTypes
	defVal     interface{}
	short      string
	desc       string
	keyName    string
	flagName   string
	envVarName string
}

// NewRegistration creates a new registration with the given name.
func NewRegistration(name string) types.ConfigRegistration {
	return newRegistration(name)
}

func newRegistration(name string) *configReg {
	return &configReg{name: name, keys: []types.ConfigRegistrationKey{}}
}

func (r *configReg) Name() string {
	return r.name
}

func (r *configReg) Keys() <-chan types.ConfigRegistrationKey {
	c := make(chan types.ConfigRegistrationKey)
	go func() {
		for _, k := range r.keys {
			c <- k
		}
		close(c)
	}()
	return c
}

func (r *configReg) YAML() string     { return r.yaml }
func (r *configReg) SetYAML(y string) { r.yaml = y }

func (r *configReg) Key(
	keyType types.ConfigKeyTypes,
	short string,
	defVal interface{},
	description string,
	keys ...interface{}) {

	lk := len(keys)
	if lk == 0 {
		panic(goof.New("keys is empty"))
	}

	rk := &configRegKey{
		keyType: keyType,
		short:   short,
		desc:    description,
		defVal:  defVal,
		keyName: toString(keys[0]),
	}

	if keyType == types.SecureString {
		secureKey(rk)
	}

	if lk < 2 {
		kp := strings.Split(rk.keyName, ".")
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
	} else {
		rk.flagName = toString(keys[1])
	}

	if lk < 3 {
		kp := strings.Split(rk.keyName, ".")
		for x, s := range kp {
			kp[x] = strings.ToUpper(s)
		}
		rk.envVarName = strings.Join(kp, "_")
	} else {
		rk.envVarName = toString(keys[2])
	}

	r.keys = append(r.keys, rk)
}

func (k *configRegKey) KeyType() types.ConfigKeyTypes { return k.keyType }
func (k *configRegKey) DefaultValue() interface{}     { return k.defVal }
func (k *configRegKey) Short() string                 { return k.short }
func (k *configRegKey) Description() string           { return k.desc }
func (k *configRegKey) KeyName() string               { return k.keyName }
func (k *configRegKey) FlagName() string              { return k.flagName }
func (k *configRegKey) EnvVarName() string            { return k.envVarName }

func secureKey(k *configRegKey) {
	secureKeysRWL.Lock()
	defer secureKeysRWL.Unlock()
	kn := strings.ToLower(k.keyName)
	if LogSecureKey {
		log.WithField("keyName", kn).Debug("securing key")
	}
	secureKeys[kn] = k
}
