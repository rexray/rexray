package handlers

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
)

// postArgsHandler is an HTTP filter for injecting the store with the POST
// object's fields and additional options
type postArgsHandler struct {
	handler types.APIFunc
	config  gofig.Config
}

// NewPostArgsHandler returns a new filter for injecting the store with the
// POST object's fields and additional options.
func NewPostArgsHandler(config gofig.Config) types.Middleware {
	return &postArgsHandler{config: config}
}

func (h *postArgsHandler) Name() string {
	return "post-args-handler"
}

func (h *postArgsHandler) Handler(m types.APIFunc) types.APIFunc {
	return (&postArgsHandler{m, h.config}).Handle
}

// Handle is the type's Handler function.
func (h *postArgsHandler) Handle(
	ctx types.Context,
	w http.ResponseWriter,
	req *http.Request,
	store types.Store) error {

	reqObj := ctx.Value("reqObj")
	if reqObj == nil {
		return fmt.Errorf("missing request object")
	}

	v := reflect.ValueOf(reqObj).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i).Interface()

		switch tfv := fv.(type) {
		case nil:
			// do nothing
		case map[string]interface{}:
			store.Set(getFieldName(ft), utils.NewStoreWithData(tfv))
		default:
			// add it to the store
			store.Set(getFieldName(ft), fv)
		}
	}

	if h.config.GetBool(types.ConfigServerParseRequestOpts) {
		ctx.Debug("parsing req opts enabled")
		if store.IsSet("opts") {
			ctx.Debug("parsing req opts: is set")
			if opts, ok := store.Get("opts").(types.Store); ok {
				ctx.Debug("parsing req opts: valid type")
				for _, k := range opts.Keys() {
					store.Set(k, opts.Get(k))
					ctx.WithField("optsKey", k).Debug(
						"parsing req opts: set key")
				}
			}
		}
	}

	return h.handler(ctx, w, req, store)
}

func getFieldName(ft reflect.StructField) string {
	fn := ft.Name
	if tag := ft.Tag.Get("json"); tag != "" {
		if tag != "-" {
			tagParts := strings.Split(tag, ",")
			if tagParts[0] != "" {
				fn = tagParts[0]
			}
		}
	}
	return fn
}
