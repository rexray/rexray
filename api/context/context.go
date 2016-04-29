package context

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/akutz/goof"
	gcontext "github.com/gorilla/context"
	"golang.org/x/net/context"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/utils"
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

	timestampTypeName = utils.GetTypePkgPathAndName(time.Time{})

	loggerTypeName = utils.GetTypePkgPathAndName(
		&log.Logger{})
)

type lsc struct {
	context.Context
	*log.Logger
	req *http.Request
}

// Background initializes a new, empty context.
func Background() types.Context {
	return NewContext(context.Background(), nil)
}

// NewContext initializes a new libStorage context.
func NewContext(parent context.Context, r *http.Request) types.Context {
	var parentLogger *log.Logger
	if parentCtx, ok := parent.(*lsc); ok {
		parentLogger = parentCtx.Logger
	} else {
		parentLogger = log.StandardLogger()
	}

	ctx := &lsc{
		Context: parent,
		req:     r,
	}

	ctx.Logger = &log.Logger{
		Formatter: &fieldFormatter{parentLogger.Formatter, ctx},
		Out:       parentLogger.Out,
		Hooks:     parentLogger.Hooks,
		Level:     parentLogger.Level,
	}

	return ctx
}

// WithInstanceIDsByService returns a context with the provided instance ID map.
func WithInstanceIDsByService(
	parent context.Context, val map[string]*types.InstanceID) types.Context {
	return WithValue(parent, types.ContextKeyInstanceIDsByService, val)
}

// WithInstanceID returns a context with the provided instance ID.
func WithInstanceID(
	parent context.Context, val *types.InstanceID) types.Context {
	return WithValue(parent, types.ContextKeyInstanceID, val)
}

// WithLocalDevicesByService returns a context with the provided service to
//  instance ID map.
func WithLocalDevicesByService(
	parent context.Context, val map[string]map[string]string) types.Context {
	return WithValue(parent, types.ContextKeyLocalDevicesByService, val)
}

// WithLocalDevices returns a context with the provided local devices map.
func WithLocalDevices(
	parent context.Context, val map[string]string) types.Context {
	return WithValue(parent, types.ContextKeyLocalDevices, val)
}

// WithProfile returns a context with the provided profile.
func WithProfile(
	parent context.Context, val string) types.Context {
	return WithValue(parent, types.ContextKeyProfile, val)
}

// WithRoute returns a contex with the provided route name.
func WithRoute(parent context.Context, val string) types.Context {
	return WithValue(parent, types.ContextKeyRoute, val)
}

// WithServiceName returns a contex with the provided service name.
func WithServiceName(parent context.Context, val string) types.Context {
	return WithValue(parent, types.ContextKeyServiceName, val)
}

// WithContextID returns a context with the provided context ID information.
// The context ID is often used with logging to identify a log statement's
// origin.
func WithContextID(
	parent context.Context, id, val string) types.Context {

	contextID := map[string]string{id: val}
	parentContextID, ok := parent.Value(
		types.ContextKeyContextID).(map[string]string)
	if ok {
		for k, v := range parentContextID {
			contextID[k] = v
		}
	}
	return WithValue(parent, types.ContextKeyContextID, contextID)
}

// WithTransactionID returns a context with the provided transaction ID.
func WithTransactionID(parent context.Context, val string) types.Context {
	return WithValue(parent, types.ContextKeyTransactionID, val)
}

// WithTransactionCreated returns a context with the provided transaction
// created timestamp.
func WithTransactionCreated(
	parent context.Context, val time.Time) types.Context {
	return WithValue(parent, types.ContextKeyTransactionCreated, val)
}

// WithValue returns a context with the provided value.
func WithValue(
	parent context.Context,
	key interface{},
	val interface{}) types.Context {
	return NewContext(context.WithValue(parent, key, val), HTTPRequest(parent))
}

// ServerName returns the context's server name.
func ServerName(ctx context.Context) (string, error) {

	obj := ctx.Value(types.ContextKeyServerName)
	if obj == nil {
		return "", utils.NewContextKeyErr(types.ContextKeyServerName)
	}
	typedObj, ok := obj.(string)
	if !ok {
		return "", utils.NewContextTypeErr(
			types.ContextKeyServerName,
			"string", utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// InstanceIDsByService gets the context's service to instance IDs map.
func InstanceIDsByService(
	ctx context.Context) (map[string]*types.InstanceID, error) {

	obj := ctx.Value(types.ContextKeyInstanceIDsByService)
	if obj == nil {
		return nil, utils.NewContextKeyErr(
			types.ContextKeyInstanceIDsByService)
	}
	typedObj, ok := obj.(map[string]*types.InstanceID)
	if !ok {
		return nil, utils.NewContextTypeErr(
			types.ContextKeyInstanceIDsByService,
			instanceIDsByServiceTypeName, utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// InstanceID gets the context's instance ID.
func InstanceID(ctx context.Context) (*types.InstanceID, error) {
	obj := ctx.Value(types.ContextKeyInstanceID)
	if obj == nil {
		return nil, utils.NewContextKeyErr(types.ContextKeyInstanceID)
	}
	typedObj, ok := obj.(*types.InstanceID)
	if !ok {
		return nil, utils.NewContextTypeErr(
			types.ContextKeyInstanceID,
			instanceIDTypeName, utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// LocalDevicesByService gets the context's service to local devices map.
func LocalDevicesByService(
	ctx context.Context) (map[string]map[string]string, error) {

	obj := ctx.Value(types.ContextKeyLocalDevicesByService)
	if obj == nil {
		return nil, utils.NewContextKeyErr(types.ContextKeyLocalDevicesByService)
	}
	typedObj, ok := obj.(map[string]map[string]string)
	if !ok {
		return nil, utils.NewContextTypeErr(
			types.ContextKeyLocalDevicesByService,
			localDevicesByServiceTypeName, utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// LocalDevices gets the context's local devices map.
func LocalDevices(ctx context.Context) (map[string]string, error) {
	obj := ctx.Value(types.ContextKeyLocalDevices)
	if obj == nil {
		return nil, utils.NewContextKeyErr(types.ContextKeyLocalDevices)
	}
	typedObj, ok := obj.(map[string]string)
	if !ok {
		return nil, utils.NewContextTypeErr(
			types.ContextKeyLocalDevices,
			localDevicesTypeName, utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// Profile gets the context's profile.
func Profile(ctx context.Context) (string, error) {
	obj := ctx.Value(types.ContextKeyProfile)
	if obj == nil {
		return "", utils.NewContextKeyErr(types.ContextKeyProfile)
	}
	typedObj, ok := obj.(string)
	if !ok {
		return "", utils.NewContextTypeErr(
			types.ContextKeyProfile,
			"string", utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// Route gets the context's route name.
func Route(ctx context.Context) (string, error) {
	obj := ctx.Value(types.ContextKeyRoute)
	if obj == nil {
		return "", utils.NewContextKeyErr(types.ContextKeyRoute)
	}
	typedObj, ok := obj.(string)
	if !ok {
		return "", utils.NewContextTypeErr(
			types.ContextKeyRoute,
			"string", utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// ServiceName returns the name of the context's service.
func ServiceName(ctx context.Context) (string, error) {
	obj := ctx.Value(types.ContextKeyServiceName)
	if obj == nil {
		return "", utils.NewContextKeyErr(types.ContextKeyServiceName)
	}
	typedObj, ok := obj.(string)
	if !ok {
		return "", utils.NewContextTypeErr(
			types.ContextKeyServiceName,
			"string", utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// TransactionID gets the context's transaction ID.
func TransactionID(ctx context.Context) (string, error) {
	obj := ctx.Value(types.ContextKeyTransactionID)
	if obj == nil {
		return "", utils.NewContextKeyErr(types.ContextKeyTransactionID)
	}
	typedObj, ok := obj.(string)
	if !ok {
		return "", utils.NewContextTypeErr(
			types.ContextKeyTransactionID,
			"string", utils.GetTypePkgPathAndName(obj))
	}
	return typedObj, nil
}

// TransactionCreated gets the context's transaction created timstamp.
func TransactionCreated(ctx context.Context) (time.Time, error) {
	obj := ctx.Value(types.ContextKeyTransactionCreated)
	if obj == nil {
		return time.Time{}, utils.NewContextKeyErr(types.ContextKeyTransactionCreated)
	}
	typedObj, ok := obj.(time.Time)
	if !ok {
		return time.Time{}, utils.NewContextTypeErr(
			types.ContextKeyTransactionCreated,
			timestampTypeName, utils.GetTypePkgPathAndName(obj))
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
func (ctx *lsc) Value(key interface{}) interface{} {
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

func (ctx *lsc) Log() *log.Logger {
	return ctx.Logger
}

func (ctx *lsc) ServerName() string {
	v, _ := ServerName(ctx)
	return v
}

func (ctx *lsc) InstanceIDsByService() map[string]*types.InstanceID {
	v, _ := InstanceIDsByService(ctx)
	return v
}

func (ctx *lsc) InstanceID() *types.InstanceID {
	v, _ := InstanceID(ctx)
	return v
}

func (ctx *lsc) LocalDevicesByService() map[string]map[string]string {
	v, _ := LocalDevicesByService(ctx)
	return v
}

func (ctx *lsc) LocalDevices() map[string]string {
	v, _ := LocalDevices(ctx)
	return v
}

func (ctx *lsc) Profile() string {
	v, _ := Profile(ctx)
	return v
}

func (ctx *lsc) Route() string {
	v, _ := Route(ctx)
	return v
}

func (ctx *lsc) ServiceName() string {
	v, _ := ServiceName(ctx)
	return v
}

func (ctx *lsc) TransactionID() string {
	v, _ := TransactionID(ctx)
	return v
}

func (ctx *lsc) TransactionCreated() time.Time {
	v, _ := TransactionCreated(ctx)
	return v
}

func (ctx *lsc) StorageDriver() types.StorageDriver {
	return nil
}

func (ctx *lsc) OSDriver() types.OSDriver {
	return nil
}

func (ctx *lsc) IntegrationDriver() types.IntegrationDriver {
	return nil
}

func (ctx *lsc) WithInstanceIDsByService(
	val map[string]*types.InstanceID) types.Context {
	return WithInstanceIDsByService(ctx, val)
}

func (ctx *lsc) WithInstanceID(
	val *types.InstanceID) types.Context {
	return WithInstanceID(ctx, val)
}

func (ctx *lsc) WithLocalDevicesByService(
	val map[string]map[string]string) types.Context {
	return WithLocalDevicesByService(ctx, val)
}

func (ctx *lsc) WithLocalDevices(
	val map[string]string) types.Context {
	return WithLocalDevices(ctx, val)
}

func (ctx *lsc) WithProfile(val string) types.Context {
	return WithProfile(ctx, val)
}

func (ctx *lsc) WithRoute(val string) types.Context {
	return WithRoute(ctx, val)
}

func (ctx *lsc) WithServiceName(val string) types.Context {
	return WithServiceName(ctx, val)
}

func (ctx *lsc) WithContextID(id, val string) types.Context {
	return WithContextID(ctx, id, val)
}

func (ctx *lsc) WithTransactionID(val string) types.Context {
	return WithTransactionID(ctx, val)
}

func (ctx *lsc) WithTransactionCreated(val time.Time) types.Context {
	return WithTransactionCreated(ctx, val)
}

// WithStorageDriver returns a context with the provided storage driver.
func (ctx *lsc) WithStorageDriver(driver types.StorageDriver) types.Context {
	return nil
}

// WithOSDriver returns a context with the provided OS driver.
func (ctx *lsc) WithOSDriver(driver types.OSDriver) types.Context {
	return nil
}

// WithIntegrationDriver sreturns a context with the provided integration
// driver.
func (ctx *lsc) WithIntegrationDriver(
	driver types.IntegrationDriver) types.Context {
	return nil
}

func (ctx *lsc) WithValue(
	key interface{}, val interface{}) types.Context {
	return WithValue(ctx, key, val)
}

// WithParent creates a copy of this context with a new parent.
func (ctx *lsc) WithParent(parent types.Context) types.Context {

	if lscParent, ok := parent.(*lsc); ok && ctx == lscParent {
		panic(goof.New("context.WithParent with same contexts"))
	}

	newLSCtx := &lsc{
		Context: parent,
		Logger:  ctx.Logger,
		req:     ctx.req,
	}

	contextID := map[string]string{}
	parentContextID, ok := parent.Value(
		types.ContextKeyContextID).(map[string]string)
	if ok {
		for k, v := range parentContextID {
			contextID[k] = v
		}
	}

	childContextID, ok := ctx.Value(
		types.ContextKeyContextID).(map[string]string)
	if ok {
		for k, v := range childContextID {
			contextID[k] = v
		}
	}

	newCtx := newLSCtx.WithValue(types.ContextKeyContextID, contextID)

	return newCtx
}

type fieldFormatter struct {
	f   log.Formatter
	ctx types.Context
}

func (f *fieldFormatter) Format(entry *log.Entry) ([]byte, error) {
	contextID, ok := f.ctx.Value(types.ContextKeyContextID).(map[string]string)
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
