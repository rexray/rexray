package context

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	gcontext "github.com/gorilla/context"
	"golang.org/x/net/context"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

var (
	instanceIDTypeName = utils.GetTypePkgPathAndName(&types.InstanceID{})
	loggerTypeName     = utils.GetTypePkgPathAndName(&log.Logger{})
)

// Context is a libStorage context.
type Context interface {
	context.Context

	// InstanceID returns the context's instance ID.
	InstanceID() *types.InstanceID

	// Profile returns the context's profile name.
	Profile() string

	// Route returns the name of context's route.
	Route() string

	// Log returns the context's logger.
	Log() *log.Logger

	// WithInstanceID returns a context with the provided instance ID.
	WithInstanceID(instanceID *types.InstanceID) Context

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

type libstorContext struct {
	context.Context
	req        *http.Request
	logger     *log.Logger
	contextIDs map[string]string
}

// Background initializes a new, empty context.
func Background() Context {
	return NewContext(context.Background(), nil)
}

// NewContext initializes a new libStorage context.
func NewContext(parent context.Context, r *http.Request) Context {

	var parentLogger *log.Logger
	if parentCtx, ok := parent.(Context); ok {
		parentLogger = parentCtx.Log()
	} else {
		parentLogger = log.StandardLogger()
	}

	ctx := &libstorContext{
		Context: parent,
		req:     r,
	}

	logger := &log.Logger{
		Formatter: &fieldFormatter{parentLogger.Formatter, ctx},
		Out:       parentLogger.Out,
		Hooks:     parentLogger.Hooks,
		Level:     parentLogger.Level,
	}
	ctx.logger = logger

	return ctx
}

// WithInstanceID returns a context with the provided instance ID.
func WithInstanceID(
	parent context.Context,
	instanceID *types.InstanceID) Context {
	return WithValue(parent, "instanceID", instanceID)
}

// WithProfile returns a context with the provided profile.
func WithProfile(
	parent context.Context,
	profile string) Context {
	return WithValue(parent, "profile", profile)
}

// WithRoute returns a contex with the provided route name.
func WithRoute(parent context.Context, routeName string) Context {
	return WithValue(parent, "route", routeName)
}

// WithContextID returns a context with the provided context ID information.
// The context ID is often used with logging to identify a log statement's
// origin.
func WithContextID(
	parent context.Context,
	id, value string) Context {

	contextID := map[string]string{id: value}
	parentContextID, ok := parent.Value("contextID").(map[string]string)
	if ok {
		for k, v := range parentContextID {
			contextID[k] = v
		}
	}
	return WithValue(parent, "contextID", contextID)
}

// WithValue returns a context with the provided value.
func WithValue(
	parent context.Context,
	key interface{},
	val interface{}) Context {
	return NewContext(context.WithValue(parent, key, val), HTTPRequest(parent))
}

// InstanceID gets the context's instance ID.
func InstanceID(ctx context.Context) (*types.InstanceID, error) {
	obj := ctx.Value("instanceID")
	if obj == nil {
		return nil, utils.NewContextKeyErr("instanceID")
	}
	typedObj, ok := obj.(*types.InstanceID)
	if !ok {
		return nil, utils.NewContextTypeErr(
			"instanceID", instanceIDTypeName, utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// Profile gets the context's profile.
func Profile(ctx context.Context) (string, error) {
	obj := ctx.Value("profile")
	if obj == nil {
		return "", utils.NewContextKeyErr("profile")
	}
	typedObj, ok := obj.(string)
	if !ok {
		return "", utils.NewContextTypeErr(
			"profile", "string", utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// Route gets the context's route name.
func Route(ctx context.Context) (string, error) {
	obj := ctx.Value("route")
	if obj == nil {
		return "", utils.NewContextKeyErr("route")
	}
	typedObj, ok := obj.(string)
	if !ok {
		return "", utils.NewContextTypeErr(
			"route", "string", utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// HTTPRequest returns the *http.Request associated with ctx using
// NewRequestContext, if any.
func HTTPRequest(ctx context.Context) *http.Request {
	req, ok := ctx.Value(reqKey).(*http.Request)
	if !ok {
		return nil
	}
	return req
}

type key int

const reqKey key = 0

// Value returns Gorilla's context package's value for this Context's request
// and key. It delegates to the parent Context if there is no such value.
func (ctx *libstorContext) Value(key interface{}) interface{} {
	if ctx.req == nil {
		ctx.Context.Value(key)
	}

	if key == reqKey {
		return ctx.req
	}
	if val, ok := gcontext.GetOk(ctx.req, key); ok {
		return val
	}

	return ctx.Context.Value(key)
}

func (ctx *libstorContext) InstanceID() *types.InstanceID {
	v, _ := InstanceID(ctx)
	return v
}

func (ctx *libstorContext) Profile() string {
	v, _ := Profile(ctx)
	return v
}

func (ctx *libstorContext) Route() string {
	v, _ := Route(ctx)
	return v
}

func (ctx *libstorContext) Log() *log.Logger {
	return ctx.logger
}

func (ctx *libstorContext) WithInstanceID(
	instanceID *types.InstanceID) Context {
	return WithInstanceID(ctx, instanceID)
}

func (ctx *libstorContext) WithProfile(profile string) Context {
	return WithProfile(ctx, profile)
}

func (ctx *libstorContext) WithRoute(routeName string) Context {
	return WithRoute(ctx, routeName)
}

func (ctx *libstorContext) WithContextID(id, value string) Context {
	return WithContextID(ctx, id, value)
}

func (ctx *libstorContext) WithValue(
	key interface{}, value interface{}) Context {
	return WithValue(ctx, key, value)
}

type fieldFormatter struct {
	f   log.Formatter
	ctx Context
}

func (f *fieldFormatter) Format(entry *log.Entry) ([]byte, error) {
	contextID, ok := f.ctx.Value("contextID").(map[string]string)
	if !ok {
		return f.f.Format(entry)
	}
	if entry.Data == nil {
		entry.Data = log.Fields{}
	}
	for k, v := range contextID {
		entry.Data[k] = v
	}
	return f.f.Format(entry)
}
