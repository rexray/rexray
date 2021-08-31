package volume

import (
	"net/http"

	gofig "github.com/akutz/gofig/types"

	"github.com/AVENTER-UG/rexray/libstorage/api/registry"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/handlers"
	"github.com/AVENTER-UG/rexray/libstorage/api/server/httputils"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils/schema"
)

// OnVolume is a handler to which an external provider can attach that is
// invoked for every Volume object produced prior to it being written to
// the response writer.
//
// If a false value is returned the volume will not be provided to the
// response writer.
var OnVolume func(
	ctx types.Context,
	req *http.Request,
	store types.Store,
	volume *types.Volume) (bool, error)

func init() {
	registry.RegisterRouter(&router{})
}

type router struct {
	config gofig.Config
	routes []types.Route
}

func (r *router) Name() string {
	return "volume-router"
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

		// get all volumes from all services
		httputils.NewGetRoute(
			"volumes",
			"/volumes",
			r.volumes,
			handlers.NewAuthAllSvcsHandler(),
			handlers.NewSchemaValidator(nil, schema.ServiceVolumeMapSchema, nil),
		),

		// get all volumes from a specific service
		httputils.NewGetRoute(
			"volumesForService",
			"/volumes/{service}",
			r.volumesForService,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(nil, schema.VolumeMapSchema, nil),
		),

		// get a specific volume from a specific service
		httputils.NewGetRoute(
			"volumeInspect",
			"/volumes/{service}/{volumeID}",
			r.volumeInspect,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(nil, schema.VolumeSchema, nil),
		),

		// POST

		// detach all volumes for a service
		httputils.NewPostRoute(
			"volumesDetachForService",
			"/volumes/{service}",
			r.volumeDetachAllForService,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(
				schema.VolumeDetachRequestSchema,
				schema.VolumeMapSchema,
				func() interface{} { return &types.VolumeDetachRequest{} }),
			handlers.NewPostArgsHandler(r.config),
		).Queries("detach"),

		// create a new volume
		httputils.NewPostRoute(
			"volumeCreate",
			"/volumes/{service}",
			r.volumeCreate,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(
				schema.VolumeCreateRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &types.VolumeCreateRequest{} }),
			handlers.NewPostArgsHandler(r.config),
		),

		// create a new volume using an existing volume as the baseline
		httputils.NewPostRoute(
			"volumeCopy",
			"/volumes/{service}/{volumeID}",
			r.volumeCopy,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(
				schema.VolumeCopyRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &types.VolumeCopyRequest{} }),
			handlers.NewPostArgsHandler(r.config),
		).Queries("copy"),

		// snapshot an existing volume
		httputils.NewPostRoute(
			"volumeSnapshot",
			"/volumes/{service}/{volumeID}",
			r.volumeSnapshot,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(
				schema.VolumeSnapshotRequestSchema,
				schema.SnapshotSchema,
				func() interface{} { return &types.VolumeSnapshotRequest{} }),
			handlers.NewPostArgsHandler(r.config),
		).Queries("snapshot"),

		// attach an existing volume
		httputils.NewPostRoute(
			"volumeAttach",
			"/volumes/{service}/{volumeID}",
			r.volumeAttach,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(
				schema.VolumeAttachRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &types.VolumeAttachRequest{} }),
			handlers.NewPostArgsHandler(r.config),
		).Queries("attach"),

		// detach all volumes for all services
		httputils.NewPostRoute(
			"volumesDetachAll",
			"/volumes",
			r.volumeDetachAll,
			handlers.NewAuthAllSvcsHandler(),
			handlers.NewSchemaValidator(
				schema.VolumeDetachRequestSchema,
				schema.ServiceVolumeMapSchema,
				func() interface{} { return &types.VolumeDetachRequest{} }),
			handlers.NewPostArgsHandler(r.config),
		).Queries("detach"),

		// detach an individual volume
		httputils.NewPostRoute(
			"volumeDetach",
			"/volumes/{service}/{volumeID}",
			r.volumeDetach,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
			handlers.NewSchemaValidator(
				schema.VolumeDetachRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &types.VolumeDetachRequest{} }),
			handlers.NewPostArgsHandler(r.config),
		).Queries("detach"),

		// DELETE
		httputils.NewDeleteRoute(
			"volumeRemove",
			"/volumes/{service}/{volumeID}",
			r.volumeRemove,
			handlers.NewServiceValidator(),
			handlers.NewAuthSvcHandler(),
			handlers.NewStorageSessionHandler(),
		),
	}
}
