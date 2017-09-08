package shim

import (
	"errors"
	"io"
	"log"

	graphDriver "github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/idtools"
	graphPlugin "github.com/docker/go-plugins-helpers/graphdriver"
)

type shimDriver struct {
	driver graphDriver.Driver
	init   graphDriver.InitFunc
}

// NewHandlerFromGraphDriver creates a plugin handler from an existing graph
// driver. This could be used, for instance, by the `overlayfs` graph driver
// built-in to Docker Engine and it would create a plugin from it that maps
// plugin API calls directly to any volume driver that satifies the
// graphdriver.Driver interface from Docker Engine.
func NewHandlerFromGraphDriver(init graphDriver.InitFunc) *graphPlugin.Handler {
	return graphPlugin.NewHandler(&shimDriver{driver: nil, init: init})
}

func (d *shimDriver) Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) error {
	driver, err := d.init(home, options, uidMaps, gidMaps)
	if err != nil {
		return err
	}
	d.driver = driver
	return nil
}

var errNotInitialized = errors.New("Not initialized")

func (d *shimDriver) Create(id, parent, mountLabel string, storageOpt map[string]string) error {
	if d == nil {
		return errNotInitialized
	}
	opts := graphDriver.CreateOpts{
		MountLabel: mountLabel,
		StorageOpt: storageOpt,
	}
	return d.driver.Create(id, parent, &opts)
}

func (d *shimDriver) CreateReadWrite(id, parent, mountLabel string, storageOpt map[string]string) error {
	if d == nil {
		return errNotInitialized
	}
	opts := graphDriver.CreateOpts{
		MountLabel: mountLabel,
		StorageOpt: storageOpt,
	}
	return d.driver.CreateReadWrite(id, parent, &opts)
}

func (d *shimDriver) Remove(id string) error {
	if d == nil {
		return errNotInitialized
	}
	return d.driver.Remove(id)
}

func (d *shimDriver) Get(id, mountLabel string) (string, error) {
	if d == nil {
		return "", errNotInitialized
	}
	return d.driver.Get(id, mountLabel)
}

func (d *shimDriver) Put(id string) error {
	if d == nil {
		return errNotInitialized
	}
	return d.driver.Put(id)
}

func (d *shimDriver) Exists(id string) bool {
	if d == nil {
		return false
	}
	return d.driver.Exists(id)
}

func (d *shimDriver) Status() [][2]string {
	if d == nil {
		return nil
	}
	return d.driver.Status()
}

func (d *shimDriver) GetMetadata(id string) (map[string]string, error) {
	if d == nil {
		return nil, errNotInitialized
	}
	return d.driver.GetMetadata(id)
}

func (d *shimDriver) Cleanup() error {
	if d == nil {
		return errNotInitialized
	}
	return d.driver.Cleanup()
}

func (d *shimDriver) Diff(id, parent string) io.ReadCloser {
	if d == nil {
		return nil
	}
	// FIXME(samoht): how do we pass the error to the driver?
	archive, err := d.driver.Diff(id, parent)
	if err != nil {
		log.Fatalf("Diff: error in stream %v", err)
	}
	return archive
}

func changeKind(c archive.ChangeType) graphPlugin.ChangeKind {
	switch c {
	case archive.ChangeModify:
		return graphPlugin.Modified
	case archive.ChangeAdd:
		return graphPlugin.Added
	case archive.ChangeDelete:
		return graphPlugin.Deleted
	}
	return 0
}

func (d *shimDriver) Changes(id, parent string) ([]graphPlugin.Change, error) {
	if d == nil {
		return nil, errNotInitialized
	}
	cs, err := d.driver.Changes(id, parent)
	if err != nil {
		return nil, err
	}
	changes := make([]graphPlugin.Change, len(cs))
	for _, c := range cs {
		change := graphPlugin.Change{
			Path: c.Path,
			Kind: changeKind(c.Kind),
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func (d *shimDriver) ApplyDiff(id, parent string, archive io.Reader) (int64, error) {
	if d == nil {
		return 0, errNotInitialized
	}
	return d.driver.ApplyDiff(id, parent, archive)
}

func (d *shimDriver) DiffSize(id, parent string) (int64, error) {
	if d == nil {
		return 0, errNotInitialized
	}
	return d.driver.DiffSize(id, parent)
}

func (d *shimDriver) Capabilities() graphDriver.Capabilities {
	if d == nil {
		return graphDriver.Capabilities{}
	}
	if capDriver, ok := d.driver.(graphDriver.CapabilityDriver); ok {
		return capDriver.Capabilities()
	}
	return graphDriver.Capabilities{}
}
