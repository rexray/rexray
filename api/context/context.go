package context

import (
	"fmt"
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

	clientTypeName = utils.GetTypePkgPathAndName(
		(*types.Client)(nil))
)

type lsc struct {
	context.Context
	*log.Logger
	req       *http.Request
	rightSide context.Context
}

// Background initializes a new, empty context.
func Background() types.Context {
	return newContext(context.Background(), nil)
}

func newContext(parent context.Context, r *http.Request) types.Context {

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

// WithHTTPRequest returns a context with the provided HTTP request.
func WithHTTPRequest(ctx types.Context, val *http.Request) types.Context {
	return newContext(ctx, val)
}

// WithInstanceIDsByService returns a context with the provided instance ID map.
func WithInstanceIDsByService(
	parent context.Context, val map[string]*types.InstanceID) types.Context {
	return WithValue(parent, types.ContextInstanceIDsByService, val)
}

// WithInstanceID returns a context with the provided instance ID.
func WithInstanceID(
	parent context.Context, val *types.InstanceID) types.Context {
	return WithValue(parent, types.ContextInstanceID, val)
}

// WithLocalDevicesByService returns a context with the provided service to
//  instance ID map.
func WithLocalDevicesByService(
	parent context.Context, val map[string]map[string]string) types.Context {
	return WithValue(parent, types.ContextLocalDevicesByService, val)
}

// WithLocalDevices returns a context with the provided local devices map.
func WithLocalDevices(
	parent context.Context, val map[string]string) types.Context {
	return WithValue(parent, types.ContextLocalDevices, val)
}

// WithProfile returns a context with the provided profile.
func WithProfile(
	parent context.Context, val string) types.Context {
	return WithValue(parent, types.ContextProfile, val)
}

// WithRoute returns a contex with the provided route name.
func WithRoute(parent context.Context, val string) types.Context {
	return WithValue(parent, types.ContextRoute, val)
}

// WithServiceName returns a contex with the provided service name.
func WithServiceName(parent context.Context, val string) types.Context {
	return WithValue(parent, types.ContextServiceName, val)
}

// WithContextID returns a context with the provided context ID information.
// The context ID is often used with logging to identify a log statement's
// origin.
func WithContextID(
	parent context.Context, id, val string) types.Context {

	contextID := map[string]string{id: val}
	parentContextID, ok := parent.Value(
		types.ContextContextID).(map[string]string)
	if ok {
		for k, v := range parentContextID {
			contextID[k] = v
		}
	}
	return WithValue(parent, types.ContextContextID, contextID)
}

// WithContextSID is the same as the WithContextID function except this
// variant only accepts fmt.Stringer values for its id argument.
func WithContextSID(
	parent context.Context, id fmt.Stringer, value string) types.Context {
	return WithContextID(parent, id.String(), value)
}

// WithTransactionID returns a context with the provided transaction ID.
func WithTransactionID(parent context.Context, val string) types.Context {
	return WithValue(parent, types.ContextTransactionID, val)
}

// WithTransactionCreated returns a context with the provided transaction
// created timestamp.
func WithTransactionCreated(
	parent context.Context, val time.Time) types.Context {
	return WithValue(parent, types.ContextTransactionCreated, val)
}

// WithClient returns a context with the provided client.
func WithClient(
	parent context.Context, val types.Client) types.Context {
	return WithValue(parent, types.ContextClient, val)
}

// WithValue returns a context with the provided value.
func WithValue(
	parent context.Context,
	key interface{},
	val interface{}) types.Context {
	return newContext(
		context.WithValue(parent, key, val),
		HTTPRequest(parent))
}

func value(
	ctx context.Context,
	key types.ContextKey) (interface{}, types.ContextKey, error) {
	if val := ctx.Value(key); val != nil {
		return val, key, nil
	}
	return nil, 0, utils.NewContextKeyErr(key)
}

// ServerName returns the context's server name.
func ServerName(ctx context.Context) (string, error) {
	val, key, err := value(ctx, types.ContextServerName)
	if err != nil {
		return "", err
	}
	if tval, ok := val.(string); ok {
		return tval, nil
	}
	return "", utils.NewContextTypeErr(
		key, "string", utils.GetTypePkgPathAndName(val))
}

// InstanceIDsByService gets the context's service to instance IDs map.
func InstanceIDsByService(
	ctx context.Context) (map[string]*types.InstanceID, error) {
	val, key, err := value(ctx, types.ContextInstanceIDsByService)
	if err != nil {
		return nil, err
	}
	if tval, ok := val.(map[string]*types.InstanceID); ok {
		return tval, nil
	}
	return nil, utils.NewContextTypeErr(
		key, instanceIDsByServiceTypeName,
		utils.GetTypePkgPathAndName(val))
}

// InstanceID gets the context's instance ID.
func InstanceID(ctx context.Context) (*types.InstanceID, error) {
	val, key, err := value(ctx, types.ContextInstanceID)
	if err != nil {
		return nil, err
	}
	if tval, ok := val.(*types.InstanceID); ok {
		return tval, nil
	}
	return nil, utils.NewContextTypeErr(
		key, instanceIDTypeName,
		utils.GetTypePkgPathAndName(val))
}

// LocalDevicesByService gets the context's service to local devices map.
func LocalDevicesByService(
	ctx context.Context) (map[string]map[string]string, error) {
	val, key, err := value(ctx, types.ContextLocalDevicesByService)
	if err != nil {
		return nil, err
	}
	if tval, ok := val.(map[string]map[string]string); ok {
		return tval, nil
	}
	return nil, utils.NewContextTypeErr(
		key, localDevicesByServiceTypeName,
		utils.GetTypePkgPathAndName(val))
}

// LocalDevices gets the context's local devices map.
func LocalDevices(ctx context.Context) (tval map[string]string, err error) {
	val, key, err := value(ctx, types.ContextLocalDevices)
	if err != nil {
		return nil, err
	}
	if tval, ok := val.(map[string]string); ok {
		return tval, nil
	}
	return nil, utils.NewContextTypeErr(
		key, localDevicesTypeName,
		utils.GetTypePkgPathAndName(val))
}

// Profile gets the context's profile.
func Profile(ctx context.Context) (string, error) {
	val, key, err := value(ctx, types.ContextProfile)
	if err != nil {
		return "", err
	}
	if tval, ok := val.(string); ok {
		return tval, nil
	}
	return "", utils.NewContextTypeErr(
		key, "string",
		utils.GetTypePkgPathAndName(val))
}

// Route gets the context's route name.
func Route(ctx context.Context) (string, error) {
	val, key, err := value(ctx, types.ContextRoute)
	if err != nil {
		return "", err
	}
	if tval, ok := val.(string); ok {
		return tval, nil
	}
	return "", utils.NewContextTypeErr(
		key, "string",
		utils.GetTypePkgPathAndName(val))
}

// ServiceName returns the name of the context's service.
func ServiceName(ctx context.Context) (string, error) {
	val, key, err := value(ctx, types.ContextServiceName)
	if err != nil {
		return "", err
	}
	if tval, ok := val.(string); ok {
		return tval, nil
	}
	return "", utils.NewContextTypeErr(
		key, "string",
		utils.GetTypePkgPathAndName(val))
}

// TransactionID gets the context's transaction ID.
func TransactionID(ctx context.Context) (string, error) {
	val, key, err := value(ctx, types.ContextTransactionID)
	if err != nil {
		return "", err
	}
	if tval, ok := val.(string); ok {
		return tval, nil
	}
	return "", utils.NewContextTypeErr(
		key, "string",
		utils.GetTypePkgPathAndName(val))
}

// TransactionCreated gets the context's transaction created timstamp.
func TransactionCreated(ctx context.Context) (time.Time, error) {
	val, key, err := value(ctx, types.ContextTransactionCreated)
	if err != nil {
		return time.Time{}, err
	}
	if tval, ok := val.(time.Time); ok {
		return tval, nil
	}
	return time.Time{}, utils.NewContextTypeErr(
		key, timestampTypeName,
		utils.GetTypePkgPathAndName(val))
}

// HTTPRequest returns the *http.Request associated with ctx using NewContext.
func HTTPRequest(ctx context.Context) *http.Request {
	val, _, err := value(ctx, types.ContextHTTPRequest)
	if err != nil {
		return nil
	}
	if tval, ok := val.(*http.Request); ok {
		return tval
	}
	return nil
}

// Client returns this context's storage driver.
func Client(ctx context.Context) (types.Client, error) {
	val, key, err := value(ctx, types.ContextClient)
	if err != nil {
		return nil, err
	}
	if tval, ok := val.(types.Client); ok {
		return tval, nil
	}
	return nil, utils.NewContextTypeErr(
		key, clientTypeName,
		utils.GetTypePkgPathAndName(val))
}

// Value returns Gorilla's context package's value for this Context's request
// and key. It delegates to the parent types.Context if there is no such value.
func (ctx *lsc) Value(key interface{}) interface{} {

	var val interface{}

	switch key {
	case types.ContextHTTPRequest:
		val = ctx.req
	case types.ContextLogger:
		val = ctx.Logger
	case ctx.req != nil:
		if reqVal, ok := gcontext.GetOk(ctx.req, key); ok {
			val = reqVal
		}
	}

	if val != nil {
		return val
	}

	if val = ctx.Context.Value(key); val != nil {
		return val
	}

	if ctx.rightSide != nil {
		return ctx.rightSide.Value(key)
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

func (ctx *lsc) Client() types.Client {
	v, _ := Client(ctx)
	return v
}

func (ctx *lsc) WithHTTPRequest(val *http.Request) types.Context {
	return WithHTTPRequest(ctx, val)
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

func (ctx *lsc) WithContextSID(id fmt.Stringer, value string) types.Context {
	return WithContextSID(ctx, id, value)
}

func (ctx *lsc) WithTransactionID(val string) types.Context {
	return WithTransactionID(ctx, val)
}

func (ctx *lsc) WithTransactionCreated(val time.Time) types.Context {
	return WithTransactionCreated(ctx, val)
}

func (ctx *lsc) WithClient(val types.Client) types.Context {
	return WithClient(ctx, val)
}

func (ctx *lsc) WithValue(
	key interface{}, val interface{}) types.Context {
	return WithValue(ctx, key, val)
}

// Join joins this context with another, such that value lookups will first
// check this context, and if no such value exist, a lookup will be performed
// against the joined context.
func (ctx *lsc) Join(rightSide types.Context) types.Context {

	if rightSide == nil {
		return ctx
	}

	if lscRightSide, ok := rightSide.(*lsc); ok && ctx == lscRightSide {
		panic(goof.New("context.Join with same contexts"))
	}

	newCtx := &lsc{
		Context:   ctx,
		rightSide: rightSide,
		Logger:    ctx.Logger,
		req:       ctx.req,
	}

	newCtxID := map[string]string{}

	rsCtxID, ok := rightSide.Value(types.ContextContextID).(map[string]string)
	if ok {
		for k, v := range rsCtxID {
			newCtxID[k] = v
		}
	}

	lsCtxID, ok := ctx.Value(types.ContextContextID).(map[string]string)
	if ok {
		for k, v := range lsCtxID {
			newCtxID[k] = v
		}
	}

	newerCtx := newCtx.WithValue(types.ContextContextID, newCtxID)
	return newerCtx
}

type fieldFormatter struct {
	f   log.Formatter
	ctx types.Context
}

func (f *fieldFormatter) Format(entry *log.Entry) ([]byte, error) {
	contextID, ok := f.ctx.Value(types.ContextContextID).(map[string]string)
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
