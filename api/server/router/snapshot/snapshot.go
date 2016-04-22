package snapshot

import (
	"github.com/akutz/gofig"
	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/server/handlers"
	"github.com/emccode/libstorage/api/server/httputils"
	apihttp "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils/schema"
)

func init() {
	registry.RegisterRouter(&router{})
}

type router struct {
	routes []apihttp.Route
}

func (r *router) Name() string {
	return "snapshot-router"
}

func (r *router) Init(config gofig.Config) {
	r.initRoutes()
}

// Routes returns the available routes.
func (r *router) Routes() []apihttp.Route {
	return r.routes
}

func (r *router) initRoutes() {
	r.routes = []apihttp.Route{

		// GET

		// get all snapshots from all services
		httputils.NewGetRoute(
			"snapshots",
			"/snapshots",
			r.snapshots,
			handlers.NewSchemaValidator(
				nil, schema.ServiceSnapshotMapSchema, nil),
		),

		// get all snapshots from a specific service
		httputils.NewGetRoute(
			"snapshotsForService",
			"/snapshots/{service}",
			r.snapshotsForService,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(
				nil, schema.SnapshotMapSchema, nil),
		),

		// get a specific snapshot from a specific service
		httputils.NewGetRoute(
			"snapshotInspect",
			"/snapshots/{service}/{snapshotID}",
			r.snapshotInspect,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(nil, schema.SnapshotSchema, nil),
		),

		// POST

		// create volume from snapshot
		httputils.NewPostRoute(
			"snapshotCreate",
			"/snapshots/{service}/{snapshotID}",
			r.volumeCreate,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeCreateRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &apihttp.VolumeCreateRequest{} }),
			handlers.NewPostArgsHandler(),
		).Queries("create"),

		// copy snapshot
		httputils.NewPostRoute(
			"snapshotCopy",
			"/snapshots/{service}/{snapshotID}",
			r.snapshotCopy,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(
				schema.SnapshotCopyRequestSchema,
				schema.SnapshotSchema,
				func() interface{} {
					return &apihttp.SnapshotCopyRequest{}
				}),
			handlers.NewPostArgsHandler(),
		).Queries("copy"),

		// DELETE
		httputils.NewDeleteRoute(
			"snapshotRemove",
			"/snapshots/{service}/{snapshotID}",
			r.snapshotRemove,
			handlers.NewServiceValidator(),
		),
	}
}
