package types

// ConfigKeyTypes is a type of configuration key.
type ConfigKeyTypes int

const (
	// String is a key with a string value
	String ConfigKeyTypes = iota // 0

	// Int is a key with an integer value
	Int // 1

	// Bool is a key with a boolean value
	Bool // 2

	// SecureString is a key with a string value that is not included when the
	// configuration is marshaled to JSON.
	SecureString // 3
)

// ConfigRegistration is an interface that describes a configuration
// registration object.
type ConfigRegistration interface {

	// Name returns the name of the config registration.
	Name() string

	// YAML returns the registration's default yaml configuration.
	YAML() string

	// SetYAML sets the registration's default yaml configuration.
	SetYAML(y string)

	// Key adds a key to the registration.
	//
	// The first vararg argument is the yaml name of the key, using a '.' as
	// the nested separator. If the second two arguments are omitted they will
	// be generated from the first argument. The second argument is the explicit
	// name of the flag bound to this key. The third argument is the explicit
	// name of the environment variable bound to thie key.
	Key(
		keyType ConfigKeyTypes,
		short string,
		defVal interface{},
		description string,
		keys ...interface{})

	// Keys returns a channel on which a listener can receive the config
	// registration's keys.
	Keys() <-chan ConfigRegistrationKey
}

// ConfigRegistrationKey is an interfact that describes a cofniguration
// registration key object.
type ConfigRegistrationKey interface {
	KeyType() ConfigKeyTypes
	DefaultValue() interface{}
	Short() string
	Description() string
	KeyName() string
	FlagName() string
	EnvVarName() string
}
