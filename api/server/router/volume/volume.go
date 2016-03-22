package volume

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
	return "volume-router"
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

		// get all volumes (and possibly attachments) from all services
		httputils.NewGetRoute(
			"volumesAndAttachments",
			"/volumes",
			newVolumesRoute(r.services, true).volumes,
			handlers.NewInstanceIDValidator(false),
			handlers.NewSchemaValidator(nil, schema.ServiceVolumeMapSchema, nil),
		).Queries("attachments"),

		// get all volumes from all services
		httputils.NewGetRoute(
			"volumes",
			"/volumes",
			newVolumesRoute(r.services, false).volumes,
			handlers.NewSchemaValidator(nil, schema.ServiceVolumeMapSchema, nil),
		),

		// get all volumes (and possibly attachments) from a specific service
		httputils.NewGetRoute(
			"volumesAndAttachmentsForService",
			"/volumes/{service}",
			newVolumesRoute(r.services, true).volumesForService,
			handlers.NewServiceValidator(r.services),
			handlers.NewInstanceIDValidator(false),
			handlers.NewSchemaValidator(nil, schema.VolumeMapSchema, nil),
		).Queries("attachments"),

		// get all volumes from a specific service
		httputils.NewGetRoute(
			"volumesForService",
			"/volumes/{service}",
			newVolumesRoute(r.services, false).volumesForService,
			handlers.NewServiceValidator(r.services),
			handlers.NewSchemaValidator(nil, schema.VolumeMapSchema, nil),
		),

		// get a specific volume (and possibly attachments) from a specific
		// service
		httputils.NewGetRoute(
			"volumeAndAttachmentsInspect",
			"/volumes/{service}/{volumeID}",
			r.volumeInspect,
			handlers.NewServiceValidator(r.services),
			handlers.NewInstanceIDValidator(false),
			handlers.NewVolumeValidator(),
			handlers.NewSchemaValidator(nil, schema.VolumeSchema, nil),
		).Queries("attachments"),

		// get a specific volume from a specific service
		httputils.NewGetRoute(
			"volumeInspect",
			"/volumes/{service}/{volumeID}",
			r.volumeInspect,
			handlers.NewServiceValidator(r.services),
			handlers.NewVolumeValidator(),
			handlers.NewSchemaValidator(nil, schema.VolumeSchema, nil),
		),

		// POST

		// detach all volumes for a service
		httputils.NewPostRoute(
			"volumesDetachForService",
			"/volumes/{service}",
			r.volumeDetachAllForService,
			handlers.NewServiceValidator(r.services),
			handlers.NewSchemaValidator(
				schema.VolumeDetachRequestSchema,
				schema.VolumeMapSchema,
				func() interface{} { return &httptypes.VolumeDetachRequest{} }),
		).Queries("detach"),

		// create a new volume
		httputils.NewPostRoute(
			"volumeCreate",
			"/volumes/{service}",
			r.volumeCreate,
			handlers.NewServiceValidator(r.services),
			handlers.NewSchemaValidator(
				schema.VolumeCreateRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &httptypes.VolumeCreateRequest{} }),
			handlers.NewPostArgsHandler(),
		),

		// create a new volume using an existing volume as the baseline
		httputils.NewPostRoute(
			"volumeCopy",
			"/volumes/{service}/{volumeID}",
			r.volumeCopy,
			handlers.NewServiceValidator(r.services),
			handlers.NewVolumeValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeCopyRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &httptypes.VolumeCopyRequest{} }),
			handlers.NewPostArgsHandler(),
		).Queries("copy"),

		// snapshot an existing volume
		httputils.NewPostRoute(
			"volumeSnapshot",
			"/volumes/{service}/{volumeID}",
			r.volumeSnapshot,
			handlers.NewServiceValidator(r.services),
			handlers.NewVolumeValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeSnapshotRequestSchema,
				schema.SnapshotSchema,
				func() interface{} { return &httptypes.VolumeSnapshotRequest{} }),
		).Queries("snapshot"),

		// attach an existing volume
		httputils.NewPostRoute(
			"volumeAttach",
			"/volumes/{service}/{volumeID}",
			r.volumeAttach,
			handlers.NewServiceValidator(r.services),
			handlers.NewInstanceIDValidator(true),
			handlers.NewVolumeValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeAttachRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &httptypes.VolumeAttachRequest{} }),
		).Queries("attach"),

		// detach all volumes for all services
		httputils.NewPostRoute(
			"volumesDetachAll",
			"/volumes",
			r.volumeDetachAll,
			handlers.NewSchemaValidator(
				schema.VolumeDetachRequestSchema,
				schema.ServiceVolumeMapSchema,
				func() interface{} { return &httptypes.VolumeDetachRequest{} }),
		).Queries("detach"),

		// detach an individual volume
		httputils.NewPostRoute(
			"volumeDetach",
			"/volumes/{service}/{volumeID}",
			r.volumeDetach,
			handlers.NewServiceValidator(r.services),
			handlers.NewVolumeValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeDetachRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &httptypes.VolumeDetachRequest{} }),
		).Queries("detach"),

		// DELETE
		httputils.NewDeleteRoute(
			"volumeRemove",
			"/volumes/{service}/{volumeID}",
			r.volumeRemove,
			handlers.NewServiceValidator(r.services),
			handlers.NewVolumeValidator(),
		),
	}
}
