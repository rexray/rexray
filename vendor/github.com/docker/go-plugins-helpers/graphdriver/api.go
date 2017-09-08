package graphdriver

// See https://github.com/docker/docker/blob/master/experimental/plugins_graphdriver.md

import (
	"io"
	"net/http"

	graphDriver "github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/go-plugins-helpers/sdk"
)

const (
	// DefaultDockerRootDirectory is the default directory where graph drivers will be created.
	DefaultDockerRootDirectory = "/var/lib/docker/graph"

	manifest         = `{"Implements": ["GraphDriver"]}`
	initPath         = "/GraphDriver.Init"
	createPath       = "/GraphDriver.Create"
	createRWPath     = "/GraphDriver.CreateReadWrite"
	removePath       = "/GraphDriver.Remove"
	getPath          = "/GraphDriver.Get"
	putPath          = "/GraphDriver.Put"
	existsPath       = "/GraphDriver.Exists"
	statusPath       = "/GraphDriver.Status"
	getMetadataPath  = "/GraphDriver.GetMetadata"
	cleanupPath      = "/GraphDriver.Cleanup"
	diffPath         = "/GraphDriver.Diff"
	changesPath      = "/GraphDriver.Changes"
	applyDiffPath    = "/GraphDriver.ApplyDiff"
	diffSizePath     = "/GraphDriver.DiffSize"
	capabilitiesPath = "/GraphDriver.Capabilities"
)

// Init

// InitRequest is the structure that docker's init requests are deserialized to.
type InitRequest struct {
	Home    string
	Options []string        `json:"Opts"`
	UIDMaps []idtools.IDMap `json:"UIDMaps"`
	GIDMaps []idtools.IDMap `json:"GIDMaps"`
}

// Create

// CreateRequest is the structure that docker's create requests are deserialized to.
type CreateRequest struct {
	ID         string
	Parent     string
	MountLabel string
	StorageOpt map[string]string
}

// Remove

// RemoveRequest is the structure that docker's remove requests are deserialized to.
type RemoveRequest struct {
	ID string
}

// Get

// GetRequest is the structure that docker's get requests are deserialized to.
type GetRequest struct {
	ID         string
	MountLabel string
}

// GetResponse is the strucutre that docker's remove responses are serialized to.
type GetResponse struct {
	Dir string
}

// Put

// PutRequest is the structure that docker's put requests are deserialized to.
type PutRequest struct {
	ID string
}

// Exists

// ExistsRequest is the structure that docker's exists requests are deserialized to.
type ExistsRequest struct {
	ID string
}

// ExistsResponse is the structure that docker's exists responses are serialized to.
type ExistsResponse struct {
	Exists bool
}

// Status

// StatusRequest is the structure that docker's status requests are deserialized to.
type StatusRequest struct{}

// StatusResponse is the structure that docker's status responses are serialized to.
type StatusResponse struct {
	Status [][2]string
}

// GetMetadata

// GetMetadataRequest is the structure that docker's getMetadata requests are deserialized to.
type GetMetadataRequest struct {
	ID string
}

// GetMetadataResponse is the structure that docker's getMetadata responses are serialized to.
type GetMetadataResponse struct {
	Metadata map[string]string
}

// Cleanup

// CleanupRequest is the structure that docker's cleanup requests are deserialized to.
type CleanupRequest struct{}

// Diff

// DiffRequest is the structure that docker's diff requests are deserialized to.
type DiffRequest struct {
	ID     string
	Parent string
}

// DiffResponse is the structure that docker's diff responses are serialized to.
type DiffResponse struct {
	Stream io.ReadCloser // TAR STREAM
}

// Changes

// ChangesRequest is the structure that docker's changes requests are deserialized to.
type ChangesRequest struct {
	ID     string
	Parent string
}

// ChangesResponse is the structure that docker's changes responses are serialized to.
type ChangesResponse struct {
	Changes []Change
}

// ChangeKind represents the type of change mage
type ChangeKind int

const (
	// Modified is a ChangeKind used when an item has been modified
	Modified ChangeKind = iota
	// Added is a ChangeKind used when an item has been added
	Added
	// Deleted is a ChangeKind used when an item has been deleted
	Deleted
)

// Change is the structure that docker's individual changes are serialized to.
type Change struct {
	Path string
	Kind ChangeKind
}

// ApplyDiff

// ApplyDiffRequest is the structure that docker's applyDiff requests are deserialized to.
type ApplyDiffRequest struct {
	Stream io.Reader // TAR STREAM
	ID     string
	Parent string
}

// ApplyDiffResponse is the structure that docker's applyDiff responses are serialized to.
type ApplyDiffResponse struct {
	Size int64
}

// DiffSize

// DiffSizeRequest is the structure that docker's diffSize requests are deserialized to.
type DiffSizeRequest struct {
	ID     string
	Parent string
}

// DiffSizeResponse is the structure that docker's diffSize responses are serialized to.
type DiffSizeResponse struct {
	Size int64
}

// CapabilitiesRequest is the structure that docker's capabilities requests are deserialized to.
type CapabilitiesRequest struct{}

// CapabilitiesResponse is the structure that docker's capabilities responses are serialized to.
type CapabilitiesResponse struct {
	Capabilities graphDriver.Capabilities
}

// ErrorResponse is a formatted error message that docker can understand
type ErrorResponse struct {
	Err string
}

// NewErrorResponse creates an ErrorResponse with the provided message
func NewErrorResponse(msg string) *ErrorResponse {
	return &ErrorResponse{Err: msg}
}

// Driver represent the interface a driver must fulfill.
type Driver interface {
	Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) error
	Create(id, parent, mountlabel string, storageOpt map[string]string) error
	CreateReadWrite(id, parent, mountlabel string, storageOpt map[string]string) error
	Remove(id string) error
	Get(id, mountLabel string) (string, error)
	Put(id string) error
	Exists(id string) bool
	Status() [][2]string
	GetMetadata(id string) (map[string]string, error)
	Cleanup() error
	Diff(id, parent string) io.ReadCloser
	Changes(id, parent string) ([]Change, error)
	ApplyDiff(id, parent string, archive io.Reader) (int64, error)
	DiffSize(id, parent string) (int64, error)
	Capabilities() graphDriver.Capabilities
}

// Handler forwards requests and responses between the docker daemon and the plugin.
type Handler struct {
	driver Driver
	sdk.Handler
}

// NewHandler initializes the request handler with a driver implementation.
func NewHandler(driver Driver) *Handler {
	h := &Handler{driver, sdk.NewHandler(manifest)}
	h.initMux()
	return h
}

func (h *Handler) initMux() {
	h.HandleFunc(initPath, func(w http.ResponseWriter, r *http.Request) {
		req := InitRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		err = h.driver.Init(req.Home, req.Options, req.UIDMaps, req.GIDMaps)
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, struct{}{}, false)
	})
	h.HandleFunc(createPath, func(w http.ResponseWriter, r *http.Request) {
		req := CreateRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		err = h.driver.Create(req.ID, req.Parent, req.MountLabel, req.StorageOpt)
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, struct{}{}, false)
	})
	h.HandleFunc(createRWPath, func(w http.ResponseWriter, r *http.Request) {
		req := CreateRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		err = h.driver.CreateReadWrite(req.ID, req.Parent, req.MountLabel, req.StorageOpt)
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, struct{}{}, false)
	})
	h.HandleFunc(removePath, func(w http.ResponseWriter, r *http.Request) {
		req := RemoveRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		err = h.driver.Remove(req.ID)
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, struct{}{}, false)

	})
	h.HandleFunc(getPath, func(w http.ResponseWriter, r *http.Request) {
		req := GetRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		dir, err := h.driver.Get(req.ID, req.MountLabel)
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, &GetResponse{Dir: dir}, false)
	})
	h.HandleFunc(putPath, func(w http.ResponseWriter, r *http.Request) {
		req := PutRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		err = h.driver.Put(req.ID)
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, struct{}{}, false)
	})
	h.HandleFunc(existsPath, func(w http.ResponseWriter, r *http.Request) {
		req := ExistsRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		exists := h.driver.Exists(req.ID)
		sdk.EncodeResponse(w, &ExistsResponse{Exists: exists}, false)
	})
	h.HandleFunc(statusPath, func(w http.ResponseWriter, r *http.Request) {
		status := h.driver.Status()
		sdk.EncodeResponse(w, &StatusResponse{Status: status}, false)
	})
	h.HandleFunc(getMetadataPath, func(w http.ResponseWriter, r *http.Request) {
		req := GetMetadataRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		metadata, err := h.driver.GetMetadata(req.ID)
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, &GetMetadataResponse{Metadata: metadata}, false)
	})
	h.HandleFunc(cleanupPath, func(w http.ResponseWriter, r *http.Request) {
		err := h.driver.Cleanup()
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, struct{}{}, false)
	})
	h.HandleFunc(diffPath, func(w http.ResponseWriter, r *http.Request) {
		req := DiffRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		stream := h.driver.Diff(req.ID, req.Parent)
		sdk.StreamResponse(w, stream)
	})
	h.HandleFunc(changesPath, func(w http.ResponseWriter, r *http.Request) {
		req := ChangesRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		changes, err := h.driver.Changes(req.ID, req.Parent)
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, &ChangesResponse{Changes: changes}, false)
	})
	h.HandleFunc(applyDiffPath, func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		id := params.Get("id")
		parent := params.Get("parent")
		size, err := h.driver.ApplyDiff(id, parent, r.Body)
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, &ApplyDiffResponse{Size: size}, false)
	})
	h.HandleFunc(diffSizePath, func(w http.ResponseWriter, r *http.Request) {
		req := DiffRequest{}
		err := sdk.DecodeRequest(w, r, &req)
		if err != nil {
			return
		}
		size, err := h.driver.DiffSize(req.ID, req.Parent)
		if err != nil {
			sdk.EncodeResponse(w, NewErrorResponse(err.Error()), true)
			return
		}
		sdk.EncodeResponse(w, &DiffSizeResponse{Size: size}, false)
	})
	h.HandleFunc(capabilitiesPath, func(w http.ResponseWriter, r *http.Request) {
		caps := h.driver.Capabilities()
		sdk.EncodeResponse(w, &CapabilitiesResponse{Capabilities: caps}, false)
	})
}
