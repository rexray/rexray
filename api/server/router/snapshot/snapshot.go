package snapshot

import (
	"github.com/akutz/gofig"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/server/handlers"
	"github.com/emccode/libstorage/api/server/httputils"
	httptypes "github.com/emccode/libstorage/api/types/http"
	"github.com/emccode/libstorage/api/utils/schema"
)

func init() {
	registry.RegisterRouter(&router{})
}

type router struct {
	config   gofig.Config
	services map[string]httputils.Service
	routes   []httputils.Route
}

func (r *router) Name() string {
	return "snapshot-router"
}

func (r *router) Init(
	config gofig.Config, services map[string]httputils.Service) {
	r.config = config
	r.services = services
	r.initRoutes()
}

// Routes returns the available routes.
func (r *router) Routes() []httputils.Route {
	return r.routes
}

func (r *router) initRoutes() {
	r.routes = []httputils.Route{

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
			handlers.NewServiceValidator(r.services),
			handlers.NewSchemaValidator(
				nil, schema.SnapshotMapSchema, nil),
		),

		// get a specific snapshot from a specific service
		httputils.NewGetRoute(
			"snapshotInspect",
			"/snapshots/{service}/{snapshotID}",
			r.snapshotInspect,
			handlers.NewServiceValidator(r.services),
			handlers.NewSnapshotValidator(),
			handlers.NewSchemaValidator(nil, schema.SnapshotSchema, nil),
		),

		// POST

		// create volume from snapshot
		httputils.NewPostRoute(
			"snapshotCreate",
			"/snapshots/{service}/{snapshotID}",
			r.volumeCreate,
			handlers.NewServiceValidator(r.services),
			handlers.NewSnapshotValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeCreateFromSnapshotRequestSchema,
				schema.VolumeSchema,
				func() interface{} {
					return &httptypes.VolumeCreateFromSnapshotRequest{}
				}),
			handlers.NewPostArgsHandler(),
		).Queries("create"),

		// copy snapshot
		httputils.NewPostRoute(
			"snapshotCopy",
			"/snapshots/{service}/{snapshotID}",
			r.snapshotCopy,
			handlers.NewServiceValidator(r.services),
			handlers.NewSnapshotValidator(),
			handlers.NewSchemaValidator(
				schema.SnapshotCopyRequestSchema,
				schema.SnapshotSchema,
				func() interface{} {
					return &httptypes.SnapshotCopyRequest{}
				}),
			handlers.NewPostArgsHandler(),
		).Queries("copy"),

		// DELETE
		httputils.NewDeleteRoute(
			"snapshotRemove",
			"/snapshots/{service}/{snapshotID}",
			r.snapshotRemove),
	}
}
