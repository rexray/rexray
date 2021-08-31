package snapshot

import (
	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/handlers"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/httputils"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils/schema"
)

func init() {
	registry.RegisterRouter(&router{})
}

type router struct {
	config gofig.Config
	routes []types.Route
}

func (r *router) Name() string {
	return "snapshot-router"
}

func (r *router) Init(config gofig.Config) {
	r.config = config
	r.initRoutes()
}

// Routes returns the available routes.
func (r *router) Routes() []types.Route {
	return r.routes
}

func (r *router) initRoutes() {
	r.routes = []types.Route{

		// GET

		// get all snapshots from all services
		httputils.NewGetRoute(
			"snapshots",
			"/snapshots",
			r.snapshots,
			handlers.NewAuthAllSvcsHandler(),
			handlers.NewSchemaValidator(
				nil, schema.ServiceSnapshotMapSchema, nil),
		),

		// get all snapshots from a specific service
		httputils.NewGetRoute(
			"snapshotsForService",
			"/snapshots/{service}",
			r.snapshotsForService,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(
				nil, schema.SnapshotMapSchema, nil),
		),

		// get a specific snapshot from a specific service
		httputils.NewGetRoute(
			"snapshotInspect",
			"/snapshots/{service}/{snapshotID}",
			r.snapshotInspect,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(nil, schema.SnapshotSchema, nil),
		),

		// POST

		// create volume from snapshot
		httputils.NewPostRoute(
			"snapshotCreate",
			"/snapshots/{service}/{snapshotID}",
			r.volumeCreate,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(
				schema.VolumeCreateRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &types.VolumeCreateRequest{} }),
			handlers.NewPostArgsHandler(r.config),
		).Queries("create"),

		// copy snapshot
		httputils.NewPostRoute(
			"snapshotCopy",
			"/snapshots/{service}/{snapshotID}",
			r.snapshotCopy,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(
				schema.SnapshotCopyRequestSchema,
				schema.SnapshotSchema,
				func() interface{} {
					return &types.SnapshotCopyRequest{}
				}),
			handlers.NewPostArgsHandler(r.config),
		).Queries("copy"),

		// DELETE
		httputils.NewDeleteRoute(
			"snapshotRemove",
			"/snapshots/{service}/{snapshotID}",
			r.snapshotRemove,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
		),
	}
}
