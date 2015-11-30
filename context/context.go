package context

import (
	"net/http"

	"github.com/akutz/goof"
	gcontext "github.com/gorilla/context"
	"golang.org/x/net/context"

	"github.com/emccode/libstorage/api"
)

// Context is the libStorage implementation of the golang context pattern as
// described in the blog post at https://blog.golang.org/context.
type Context interface {
	context.Context

	// InstanceID gets the context's instance ID.
	InstanceID() *api.InstanceID

	// Profile gets the context's profile.
	Profile() string

	// The *http.Request associated with this context, if any.
	HTTPRequest() *http.Request
}

type libstorContext struct {
	context.Context
	req *http.Request
}

// Background initializes a new, empty context.
func Background() Context {
	return NewContext(context.Background(), nil)
}

// NewContext initializes a new libStorage context.
func NewContext(parent context.Context, r *http.Request) Context {
	return &libstorContext{parent, r}
}

// WithInstanceID returns a context with the provided instance ID.
func WithInstanceID(
	parent context.Context,
	instanceID *api.InstanceID) Context {
	return WithValue(parent, "instanceID", instanceID)
}

// WithProfile returns a context with the provided profile. This context type
// is valid on the service-side only and will be ignored on the client-side.
func WithProfile(
	parent context.Context,
	profile string) Context {
	return WithValue(parent, "profile", profile)
}

// WithValue returns a context with the provided value.
func WithValue(
	parent context.Context,
	key interface{},
	val interface{}) Context {

	switch tp := parent.(type) {
	case Context:
		return NewContext(context.WithValue(tp, key, val), tp.HTTPRequest())
	default:
		return NewContext(context.WithValue(parent, key, val), nil)
	}
}

// InstanceID gets the context's instance ID.
func (ctx *libstorContext) InstanceID() *api.InstanceID {
	v, ok := ctx.Value("instanceID").(*api.InstanceID)
	if !ok {
		panic(goof.New("invalid instanceID"))
	}
	return v
}

// Profile gets the context's profile.
func (ctx *libstorContext) Profile() string {
	v, ok := ctx.Value("profile").(string)
	if !ok {
		panic(goof.New("invalid profile"))
	}
	return v
}

// HTTPRequest returns the *http.Request associated with ctx using
// NewRequestContext, if any.
func (ctx *libstorContext) HTTPRequest() *http.Request {
	if req, ok := HTTPRequest(ctx); ok {
		return req
	}
	return nil
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

// HTTPRequest returns the *http.Request associated with ctx using
// NewRequestContext, if any.
func HTTPRequest(ctx context.Context) (*http.Request, bool) {
	// We cannot use ctx.(*wrapper).req to get the request because ctx may
	// be a Context derived from a *wrapper. Instead, we use Value to
	// access the request if it is anywhere up the Context tree.
	req, ok := ctx.Value(reqKey).(*http.Request)
	return req, ok
}
