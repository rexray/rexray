package types

import (
	"io"

	"github.com/spf13/pflag"
)

// Config is the interface that enables retrieving configuration information.
// The variations of the Get function, the Set, IsSet, and Scope functions
// all take an interface{} as their first parameter. However, the param must be
// either a string or a fmt.Stringer, otherwise the function will panic.
type Config interface {

	// DisableEnvVarSubstitution is the same as the global flag,
	// DisableEnvVarSubstitution.
	DisableEnvVarSubstitution(disable bool)

	// Parent gets the configuration's parent (if set).
	Parent() Config

	// FlagSets gets the config's flag sets.
	FlagSets() map[string]*pflag.FlagSet

	// Scope returns a scoped view of the configuration. The specified scope
	// string will be used to prefix all property retrievals via the Get
	// and Set functions. Please note that the other functions will still
	// operate as they would for the non-scoped configuration instance. This
	// includes the AllSettings and AllKeys functions as well; they are *not*
	// scoped.
	Scope(scope interface{}) Config

	// GetScope returns the config's current scope (if any).
	GetScope() string

	// GetString returns the value associated with the key as a string
	GetString(k interface{}) string

	// GetBool returns the value associated with the key as a bool
	GetBool(k interface{}) bool

	// GetStringSlice returns the value associated with the key as a string
	// slice.
	GetStringSlice(k interface{}) []string

	// GetInt returns the value associated with the key as an int
	GetInt(k interface{}) int

	// Get returns the value associated with the key
	Get(k interface{}) interface{}

	// Set sets an override value
	Set(k interface{}, v interface{})

	// IsSet returns a flag indicating whether or not a key is set.
	IsSet(k interface{}) bool

	// Copy creates a copy of this Config instance
	Copy() (Config, error)

	// ToJSON exports this Config instance to a JSON string
	ToJSON() (string, error)

	// ToJSONCompact exports this Config instance to a compact JSON string
	ToJSONCompact() (string, error)

	// MarshalJSON implements the encoding/json.Marshaller interface. It allows
	// this type to provide its own marshalling routine.
	MarshalJSON() ([]byte, error)

	// ReadConfig reads a configuration stream into the current config instance
	ReadConfig(in io.Reader) error

	// ReadConfigFile reads a configuration files into the current config
	// instance
	ReadConfigFile(filePath string) error

	// EnvVars returns an array of the initialized configuration keys as
	// key=value strings where the key is configuration key's environment
	// variable key and the value is the current value for that key.
	EnvVars() []string

	// AllKeys gets a list of all the keys present in this configuration.
	AllKeys() []string

	// AllSettings gets a map of this configuration's settings.
	AllSettings() map[string]interface{}
}
