package types

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

// Context is a libStorage context.
type Context interface {
	context.Context

	// InstanceID returns the context's instance ID.
	InstanceID() *InstanceID

	// Profile returns the context's profile name.
	Profile() string

	// Route returns the name of context's route.
	Route() string

	// Log returns the context's logger.
	Log() *log.Logger

	// WithInstanceID returns a context with the provided instance ID.
	WithInstanceID(instanceID *InstanceID) Context

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
