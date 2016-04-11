package volume

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
	config gofig.Config
	routes []apihttp.Route
}

func (r *router) Name() string {
	return "volume-router"
}

func (r *router) Init(config gofig.Config) {
	r.config = config
	r.initRoutes()
}

// Routes returns the available routes.
func (r *router) Routes() []apihttp.Route {
	return r.routes
}

func (r *router) initRoutes() {
	r.routes = []apihttp.Route{
		// GET

		// get all volumes (and possibly attachments) from all services
		httputils.NewGetRoute(
			"volumesAndAttachments",
			"/volumes",
			newVolumesRoute(r.config, true).volumes,
			handlers.NewInstanceIDValidator(false),
			handlers.NewSchemaValidator(nil, schema.ServiceVolumeMapSchema, nil),
		).Queries("attachments"),

		// get all volumes from all services
		httputils.NewGetRoute(
			"volumes",
			"/volumes",
			newVolumesRoute(r.config, false).volumes,
			handlers.NewSchemaValidator(nil, schema.ServiceVolumeMapSchema, nil),
		),

		// get all volumes (and possibly attachments) from a specific service
		httputils.NewGetRoute(
			"volumesAndAttachmentsForService",
			"/volumes/{service}",
			newVolumesRoute(r.config, true).volumesForService,
			handlers.NewServiceValidator(),
			handlers.NewInstanceIDValidator(false),
			handlers.NewSchemaValidator(nil, schema.VolumeMapSchema, nil),
		).Queries("attachments"),

		// get all volumes from a specific service
		httputils.NewGetRoute(
			"volumesForService",
			"/volumes/{service}",
			newVolumesRoute(r.config, false).volumesForService,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(nil, schema.VolumeMapSchema, nil),
		),

		// get a specific volume (and possibly attachments) from a specific
		// service
		httputils.NewGetRoute(
			"volumeAndAttachmentsInspect",
			"/volumes/{service}/{volumeID}",
			r.volumeInspect,
			handlers.NewServiceValidator(),
			handlers.NewInstanceIDValidator(false),
			handlers.NewSchemaValidator(nil, schema.VolumeSchema, nil),
		).Queries("attachments"),

		// get a specific volume from a specific service
		httputils.NewGetRoute(
			"volumeInspect",
			"/volumes/{service}/{volumeID}",
			r.volumeInspect,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(nil, schema.VolumeSchema, nil),
		),

		// POST

		// detach all volumes for a service
		httputils.NewPostRoute(
			"volumesDetachForService",
			"/volumes/{service}",
			r.volumeDetachAllForService,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeDetachRequestSchema,
				schema.VolumeMapSchema,
				func() interface{} { return &apihttp.VolumeDetachRequest{} }),
		).Queries("detach"),

		// create a new volume
		httputils.NewPostRoute(
			"volumeCreate",
			"/volumes/{service}",
			r.volumeCreate,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeCreateRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &apihttp.VolumeCreateRequest{} }),
			handlers.NewPostArgsHandler(),
		),

		// create a new volume using an existing volume as the baseline
		httputils.NewPostRoute(
			"volumeCopy",
			"/volumes/{service}/{volumeID}",
			r.volumeCopy,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeCopyRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &apihttp.VolumeCopyRequest{} }),
			handlers.NewPostArgsHandler(),
		).Queries("copy"),

		// snapshot an existing volume
		httputils.NewPostRoute(
			"volumeSnapshot",
			"/volumes/{service}/{volumeID}",
			r.volumeSnapshot,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeSnapshotRequestSchema,
				schema.SnapshotSchema,
				func() interface{} { return &apihttp.VolumeSnapshotRequest{} }),
		).Queries("snapshot"),

		// attach an existing volume
		httputils.NewPostRoute(
			"volumeAttach",
			"/volumes/{service}/{volumeID}",
			r.volumeAttach,
			handlers.NewServiceValidator(),
			handlers.NewInstanceIDValidator(true),
			handlers.NewSchemaValidator(
				schema.VolumeAttachRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &apihttp.VolumeAttachRequest{} }),
		).Queries("attach"),

		// detach all volumes for all services
		httputils.NewPostRoute(
			"volumesDetachAll",
			"/volumes",
			r.volumeDetachAll,
			handlers.NewSchemaValidator(
				schema.VolumeDetachRequestSchema,
				schema.ServiceVolumeMapSchema,
				func() interface{} { return &apihttp.VolumeDetachRequest{} }),
		).Queries("detach"),

		// detach an individual volume
		httputils.NewPostRoute(
			"volumeDetach",
			"/volumes/{service}/{volumeID}",
			r.volumeDetach,
			handlers.NewServiceValidator(),
			handlers.NewSchemaValidator(
				schema.VolumeDetachRequestSchema,
				schema.VolumeSchema,
				func() interface{} { return &apihttp.VolumeDetachRequest{} }),
		).Queries("detach"),

		// DELETE
		httputils.NewDeleteRoute(
			"volumeRemove",
			"/volumes/{service}/{volumeID}",
			r.volumeRemove,
			handlers.NewServiceValidator(),
		),
	}
}
