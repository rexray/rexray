package handlers

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	//log "github.com/Sirupsen/logrus"

	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/api/types/context"
	apihttp "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils"
)

// postArgsHandler is an HTTP filter for injecting the store with the POST
// object's fields and additional options
type postArgsHandler struct {
	handler apihttp.APIFunc
}

// NewPostArgsHandler returns a new filter for injecting the store with the
// POST object's fields and additional options.
func NewPostArgsHandler() apihttp.Middleware {
	return &postArgsHandler{}
}

func (h *postArgsHandler) Name() string {
	return "post-args-handler"
}

func (h *postArgsHandler) Handler(m apihttp.APIFunc) apihttp.APIFunc {
	return (&postArgsHandler{m}).Handle
}

// Handle is the type's Handler function.
func (h *postArgsHandler) Handle(
	ctx context.Context,
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
