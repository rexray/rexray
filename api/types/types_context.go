package types

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	// ContextKeyInstanceID is a context key.
	ContextKeyInstanceID = "instanceID"
	// ContextKeyInstanceIDsByService is a context key.
	ContextKeyInstanceIDsByService = "InstanceIDsByService"
	// ContextKeyProfile is a context key.
	ContextKeyProfile = "profile"
	// ContextKeyRoute is a context key.
	ContextKeyRoute = "route"
	// ContextKeyContextID is a context key.
	ContextKeyContextID = "contextID"
	// ContextKeyService is a context key.
	ContextKeyService = "service"
	// ContextKeyServiceName is a context key.
	ContextKeyServiceName = "serviceName"
	// ContextKeyDriver is a context key.
	ContextKeyDriver = "driver"
	// ContextKeyDriverName is a context key.
	ContextKeyDriverName = "driverName"
	// ContextKeyLocalDevices is a context key.
	ContextKeyLocalDevices = "localDevices"
	// ContextKeyLocalDevicesByService is a context key.
	ContextKeyLocalDevicesByService = "localDevicesByService"
)

// Context is a libStorage context.
type Context interface {
	context.Context

	// InstanceIDsByService returns the context's service to instance ID map.
	InstanceIDsByService() map[string]*InstanceID

	// InstanceID returns the context's instance ID.
	InstanceID() *InstanceID

	// LocalDevicesByService returns the context's service to local devices map.
	LocalDevicesByService() map[string]map[string]string

	// LocalDevices returns the context's local devices map.
	LocalDevices() map[string]string

	// Profile returns the context's profile name.
	Profile() string

	// Route returns the name of context's route.
	Route() string

	// Log returns the context's logger.
	Log() *log.Logger

	// WithInstanceIDsByService returns a context with the provided service to
	// instance ID map.
	WithInstanceIDsByService(val map[string]*InstanceID) Context

	// WithInstanceID returns a context with the provided instance ID.
	WithInstanceID(val *InstanceID) Context

	// WithLocalDevicesByService returns a context with the provided service to
	// local devices map.
	WithLocalDevicesByService(val map[string]map[string]string) Context

	// WithLocalDevices returns a context with the provided local devices map.
	WithLocalDevices(val map[string]string) Context

	// WithProfile returns a context with the provided profile.
	WithProfile(profile string) Context

	// WithRoute returns a contex with the provided route name.
	WithRoute(routeName string) Context

	// WithContextID returns a context with the provided context ID information.
	// The context ID is often used with logging to identify a log statement's
	// origin.
	WithContextID(id, value string) Context

	// WithValue returns a context with the provided value.
	WithValue(key interface{}, val interface{}) Context
}
