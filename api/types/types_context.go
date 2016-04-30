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

	// StorageDriver returns this context's storage driver.
	StorageDriver() StorageDriver

	// OSDriver returns this context's OS driver.
	OSDriver() OSDriver

	// IntegrationDriver returns this context's integration driver.
	IntegrationDriver() IntegrationDriver

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

	// WithStorageDriver returns a context with the provided storage driver.
	WithStorageDriver(driver StorageDriver) Context

	// WithOSDriver returns a context with the provided OS driver.
	WithOSDriver(driver OSDriver) Context

	// WithIntegrationDriver returns a context with the provided integration
	// driver.
	WithIntegrationDriver(driver IntegrationDriver) Context
}

// ContextKey is the type used as a context key.
type ContextKey int

// String returns the string-representation of the context key.
func (ck ContextKey) String() string {
	return ctxIDKeys[ck]
}

const (

	// CtxKeyHTTPRequest is a context key.
	CtxKeyHTTPRequest ContextKey = iota

	// CtxKeyConfig is a context key.
	CtxKeyConfig

	// CtxKeyLogger is a context key.
	CtxKeyLogger

	// CtxKeyInstanceID is a context key.
	CtxKeyInstanceID

	// CtxKeyInstanceIDsByService is a context key.
	CtxKeyInstanceIDsByService

	// CtxKeyProfile is a context key.
	CtxKeyProfile

	// CtxKeyRoute is a context key.
	CtxKeyRoute

	// CtxKeyContextID is a context key.
	CtxKeyContextID

	// CtxKeyService is a context key.
	CtxKeyService

	// CtxKeyServiceName is a context key.
	CtxKeyServiceName

	// CtxKeyDriver is a context key.
	CtxKeyDriver

	// CtxKeyDriverName is a context key.
	CtxKeyDriverName

	// CtxKeyLocalDevices is a context key.
	CtxKeyLocalDevices

	// CtxKeyLocalDevicesByService is a context key.
	CtxKeyLocalDevicesByService

	// CtxKeyServerName is a context key.
	CtxKeyServerName

	// CtxKeyTransactionID is a context key.
	CtxKeyTransactionID

	// CtxKeyTransactionCreated is a context key.
	CtxKeyTransactionCreated

	// CtxKeyOSDriver is a context key.
	CtxKeyOSDriver

	// CtxKeyStorageDriver is a context key.
	CtxKeyStorageDriver

	// CtxKeyIntegrationDriver is a context key.
	CtxKeyIntegrationDriver
)

var (
	ctxIDKeys = map[ContextKey]string{
		CtxKeyHTTPRequest:           "httpRequest",
		CtxKeyConfig:                "config",
		CtxKeyLogger:                "logger",
		CtxKeyInstanceID:            "instanceID",
		CtxKeyInstanceIDsByService:  "instanceIDsByService",
		CtxKeyProfile:               "profile",
		CtxKeyRoute:                 "route",
		CtxKeyContextID:             "contextID",
		CtxKeyService:               "service",
		CtxKeyServiceName:           "serviceName",
		CtxKeyDriver:                "driver",
		CtxKeyDriverName:            "driverName",
		CtxKeyLocalDevices:          "localDevices",
		CtxKeyLocalDevicesByService: "localDevicesByService",
		CtxKeyServerName:            "serverName",
		CtxKeyTransactionID:         "txID",
		CtxKeyTransactionCreated:    "txCR",
		CtxKeyOSDriver:              "osDriver",
		CtxKeyStorageDriver:         "storageDriver",
		CtxKeyIntegrationDriver:     "integrationDriver",
	}
)
