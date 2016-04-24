package context

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	gcontext "github.com/gorilla/context"
	"golang.org/x/net/context"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
)

const (
	// ContextKeyInstanceID is a context key.
	ContextKeyInstanceID = types.ContextKeyInstanceID
	// ContextKeyInstanceIDsByService is a context key.
	ContextKeyInstanceIDsByService = types.ContextKeyInstanceIDsByService
	// ContextKeyProfile is a context key.
	ContextKeyProfile = types.ContextKeyProfile
	// ContextKeyRoute is a context key.
	ContextKeyRoute = types.ContextKeyRoute
	// ContextKeyContextID is a context key.
	ContextKeyContextID = types.ContextKeyContextID
	// ContextKeyService is a context key.
	ContextKeyService = types.ContextKeyService
	// ContextKeyServiceName is a context key.
	ContextKeyServiceName = types.ContextKeyServiceName
	// ContextKeyDriver is a context key.
	ContextKeyDriver = types.ContextKeyDriver
	// ContextKeyDriverName is a context key.
	ContextKeyDriverName = types.ContextKeyDriverName
	// ContextKeyLocalDevices is a context key.
	ContextKeyLocalDevices = types.ContextKeyLocalDevices
	// ContextKeyLocalDevicesByService is a context key.
	ContextKeyLocalDevicesByService = types.ContextKeyLocalDevicesByService
)

var (
	instanceIDTypeName = utils.GetTypePkgPathAndName(
		&types.InstanceID{})

	instanceIDsByServiceTypeName = utils.GetTypePkgPathAndName(
		map[string]*types.InstanceID{})

	localDevicesTypeName = utils.GetTypePkgPathAndName(
		map[string]string{})

	localDevicesByServiceTypeName = utils.GetTypePkgPathAndName(
		map[string]map[string]string{})

	loggerTypeName = utils.GetTypePkgPathAndName(
		&log.Logger{})
)

// Context is a libStorage context.
type Context types.Context

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

// WithInstanceIDsByService returns a context with the provided instance ID map.
func WithInstanceIDsByService(
	parent context.Context, val map[string]*types.InstanceID) Context {
	return WithValue(parent, ContextKeyInstanceIDsByService, val)
}

// WithInstanceID returns a context with the provided instance ID.
func WithInstanceID(
	parent context.Context, val *types.InstanceID) Context {
	return WithValue(parent, ContextKeyInstanceID, val)
}

// WithLocalDevicesByService returns a context with the provided service to
//  instance ID map.
func WithLocalDevicesByService(
	parent context.Context, val map[string]map[string]string) Context {
	return WithValue(parent, ContextKeyLocalDevicesByService, val)
}

// WithLocalDevices returns a context with the provided local devices map.
func WithLocalDevices(
	parent context.Context, val map[string]string) Context {
	return WithValue(parent, ContextKeyLocalDevices, val)
}

// WithProfile returns a context with the provided profile.
func WithProfile(
	parent context.Context, val string) Context {
	return WithValue(parent, ContextKeyProfile, val)
}

// WithRoute returns a contex with the provided route name.
func WithRoute(parent context.Context, val string) Context {
	return WithValue(parent, ContextKeyRoute, val)
}

// WithContextID returns a context with the provided context ID information.
// The context ID is often used with logging to identify a log statement's
// origin.
func WithContextID(
	parent context.Context, id, val string) Context {

	contextID := map[string]string{id: val}
	parentContextID, ok := parent.Value(contextID).(map[string]string)
	if ok {
		for k, v := range parentContextID {
			contextID[k] = v
		}
	}
	return WithValue(parent, ContextKeyContextID, contextID)
}

// WithValue returns a context with the provided value.
func WithValue(
	parent context.Context,
	key interface{},
	val interface{}) Context {
	return NewContext(context.WithValue(parent, key, val), HTTPRequest(parent))
}

// InstanceIDsByService gets the context's service to instance IDs map.
func InstanceIDsByService(
	ctx context.Context) (map[string]*types.InstanceID, error) {

	obj := ctx.Value(ContextKeyInstanceIDsByService)
	if obj == nil {
		return nil, utils.NewContextKeyErr(ContextKeyInstanceIDsByService)
	}
	typedObj, ok := obj.(map[string]*types.InstanceID)
	if !ok {
		return nil, utils.NewContextTypeErr(
			ContextKeyInstanceIDsByService,
			instanceIDsByServiceTypeName, utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// InstanceID gets the context's instance ID.
func InstanceID(ctx context.Context) (*types.InstanceID, error) {
	obj := ctx.Value(ContextKeyInstanceID)
	if obj == nil {
		return nil, utils.NewContextKeyErr(ContextKeyInstanceID)
	}
	typedObj, ok := obj.(*types.InstanceID)
	if !ok {
		return nil, utils.NewContextTypeErr(
			ContextKeyInstanceID,
			instanceIDTypeName, utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// LocalDevicesByService gets the context's service to local devices map.
func LocalDevicesByService(
	ctx context.Context) (map[string]map[string]string, error) {

	obj := ctx.Value(ContextKeyLocalDevicesByService)
	if obj == nil {
		return nil, utils.NewContextKeyErr(ContextKeyLocalDevicesByService)
	}
	typedObj, ok := obj.(map[string]map[string]string)
	if !ok {
		return nil, utils.NewContextTypeErr(
			ContextKeyLocalDevicesByService,
			localDevicesByServiceTypeName, utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// LocalDevices gets the context's local devices map.
func LocalDevices(ctx context.Context) (map[string]string, error) {
	obj := ctx.Value(ContextKeyLocalDevices)
	if obj == nil {
		return nil, utils.NewContextKeyErr(ContextKeyLocalDevices)
	}
	typedObj, ok := obj.(map[string]string)
	if !ok {
		return nil, utils.NewContextTypeErr(
			ContextKeyLocalDevices,
			localDevicesTypeName, utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// Profile gets the context's profile.
func Profile(ctx context.Context) (string, error) {
	obj := ctx.Value(ContextKeyProfile)
	if obj == nil {
		return "", utils.NewContextKeyErr(ContextKeyProfile)
	}
	typedObj, ok := obj.(string)
	if !ok {
		return "", utils.NewContextTypeErr(
			ContextKeyProfile,
			"string", utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// Route gets the context's route name.
func Route(ctx context.Context) (string, error) {
	obj := ctx.Value(ContextKeyRoute)
	if obj == nil {
		return "", utils.NewContextKeyErr(ContextKeyRoute)
	}
	typedObj, ok := obj.(string)
	if !ok {
		return "", utils.NewContextTypeErr(
			ContextKeyRoute,
			"string", utils.GetTypePkgPathAndName(obj))
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
// and key. It delegates to the parent types.Context if there is no such value.
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

func (ctx *libstorContext) InstanceIDsByService() map[string]*types.InstanceID {
	v, _ := InstanceIDsByService(ctx)
	return v
}

func (ctx *libstorContext) InstanceID() *types.InstanceID {
	v, _ := InstanceID(ctx)
	return v
}

func (ctx *libstorContext) LocalDevicesByService() map[string]map[string]string {
	v, _ := LocalDevicesByService(ctx)
	return v
}

func (ctx *libstorContext) LocalDevices() map[string]string {
	v, _ := LocalDevices(ctx)
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

func (ctx *libstorContext) WithInstanceIDsByService(
	val map[string]*types.InstanceID) types.Context {
	return WithInstanceIDsByService(ctx, val)
}

func (ctx *libstorContext) WithInstanceID(
	val *types.InstanceID) types.Context {
	return WithInstanceID(ctx, val)
}

func (ctx *libstorContext) WithLocalDevicesByService(
	val map[string]map[string]string) types.Context {
	return WithLocalDevicesByService(ctx, val)
}

func (ctx *libstorContext) WithLocalDevices(
	val map[string]string) types.Context {
	return WithLocalDevices(ctx, val)
}

func (ctx *libstorContext) WithProfile(val string) types.Context {
	return WithProfile(ctx, val)
}

func (ctx *libstorContext) WithRoute(val string) types.Context {
	return WithRoute(ctx, val)
}

func (ctx *libstorContext) WithContextID(id, val string) types.Context {
	return WithContextID(ctx, id, val)
}

func (ctx *libstorContext) WithValue(
	key interface{}, val interface{}) types.Context {
	return WithValue(ctx, key, val)
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
