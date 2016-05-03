package types

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/gofig"
	"golang.org/x/net/context"
)

// Context is a libStorage context.
type Context interface {
	context.Context
	gofig.Config
	log.FieldLogger

	// Join joins this context with another, such that value lookups will first
	// check this context, and if no such value exist, a lookup will be
	// performed against the joined context.
	Join(rightSide Context) Context

	// Log returns the underlying logger instance.
	Log() *log.Logger

	// ServerName gets the server name.
	ServerName() string

	// TransactionID gets the transaction ID.
	TransactionID() string

	// TransactionCreated gets the timestamp of when the transaction was
	// created.
	TransactionCreated() time.Time

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

	// ServiceName returns the name of the context's service.
	ServiceName() string

	// Client returns this context's client.
	Client() Client

	// WithConfig returns a context with the provided config.
	WithConfig(config gofig.Config) Context

	// WithHTTPRequest returns a context with the provided HTTP request.
	WithHTTPRequest(req *http.Request) Context

	// WithValue returns a context with the provided value.
	WithValue(key interface{}, val interface{}) Context

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

	// WithServiceName returns a contex with the provided service name.
	WithServiceName(serviceName string) Context

	// WithContextID returns a context with the provided context ID information.
	// The context ID is often used with logging to identify a log statement's
	// origin.
	WithContextID(id, value string) Context

	// WithContextSID is the same as the WithContextID function except this
	// variant only accepts fmt.Stringer values for its id argument.
	WithContextSID(id fmt.Stringer, value string) Context

	// WithTransactionID returns a context with the provided transaction ID.
	WithTransactionID(transactionID string) Context

	// WithTransactionCreated returns a context with the provided transaction
	// created timestamp.
	WithTransactionCreated(timestamp time.Time) Context

	// WithClient returns a context with the provided client.
	WithClient(client Client) Context
}

// ContextKey is the type used as a context key.
type ContextKey int

// String returns the string-representation of the context key.
func (ck ContextKey) String() string {
	if v, ok := ctxIDKeys[ck]; ok {
		return v
	}
	return ""
}

const (
	// ContextHTTPRequest is a context key.
	ContextHTTPRequest ContextKey = 5000 + iota

	// ContextConfig is a context key.
	ContextConfig

	// ContextLogger is a context key.
	ContextLogger

	// ContextInstanceID is a context key.
	ContextInstanceID

	// ContextInstanceIDsByService is a context key.
	ContextInstanceIDsByService

	// ContextProfile is a context key.
	ContextProfile

	// ContextRoute is a context key.
	ContextRoute

	// ContextContextID is a context key.
	ContextContextID

	// ContextService is a context key.
	ContextService

	// ContextServiceName is a context key.
	ContextServiceName

	// ContextDriver is a context key.
	ContextDriver

	// ContextDriverName is a context key.
	ContextDriverName

	// ContextLocalDevices is a context key.
	ContextLocalDevices

	// ContextLocalDevicesByService is a context key.
	ContextLocalDevicesByService

	// ContextServerName is a context key.
	ContextServerName

	// ContextTransactionID is a context key.
	ContextTransactionID

	// ContextTransactionCreated is a context key.
	ContextTransactionCreated

	// ContextClient is a context key.
	ContextClient

	// ContextOSDriver is a context key.
	ContextOSDriver

	// ContextStorageDriver is a context key.
	ContextStorageDriver

	// ContextIntegrationDriver is a context key.
	ContextIntegrationDriver

	// ContexUser is a context key.
	ContextUser

	// ContextHost is a context key.
	ContextHost
)

var (
	ctxIDKeys = map[ContextKey]string{
		ContextLogger:             "logger",
		ContextInstanceID:         "instanceID",
		ContextProfile:            "profile",
		ContextRoute:              "route",
		ContextService:            "service",
		ContextServiceName:        "service",
		ContextDriver:             "driver",
		ContextDriverName:         "driver",
		ContextServerName:         "server",
		ContextTransactionID:      "txID",
		ContextTransactionCreated: "txCR",
		ContextOSDriver:           "osDriver",
		ContextStorageDriver:      "storageDriver",
		ContextIntegrationDriver:  "integrationDriver",
		ContextUser:               "user",
		ContextHost:               "host",
	}
)
