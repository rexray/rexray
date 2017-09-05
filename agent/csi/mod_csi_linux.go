package csi

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"plugin"
	"strings"
	"sync"

	"github.com/codedellemc/goioc"
)

var loadGoPluginsOnce sync.Once

func init() {
	loadGoPluginsFunc = LoadGoPlugins
}

// LoadGoPlugins reads the value of the environment variable
// CSI_GOPLUGINS, a CSV with each element being the absolute path to
// a shared object file -- a Go plug-in.
func LoadGoPlugins(ctx context.Context, filePaths ...string) (err error) {
	loadGoPluginsOnce.Do(func() { err = loadGoPlugins(ctx, filePaths...) })
	return err
}

func loadGoPlugins(ctx context.Context, filePaths ...string) error {

	if len(filePaths) == 0 {
		var err error
		// read the paths of the go plug-in files
		rdr := csv.NewReader(strings.NewReader(os.Getenv("CSI_GOPLUGINS")))
		if filePaths, err = rdr.Read(); err != nil && err != io.EOF {
			return err
		}
	}
	if len(filePaths) == 0 {
		return nil
	}

	// iterate the shared object files and load them one at a time
	for _, so := range filePaths {

		// attempt to open the plug-in
		p, err := plugin.Open(so)
		if err != nil {
			return err
		}
		log.Printf("loaded plug-in: %s\n", so)

		spSym, err := p.Lookup("ServiceProviders")
		if err != nil {
			return err
		}
		sps, ok := spSym.(*map[string]func() interface{})
		if !ok {
			return fmt.Errorf(
				"error: invalid ServiceProviders field: %T", spSym)
		}

		if sps == nil {
			return fmt.Errorf("error: nil ServiceProviders")
		}

		// record the service provider names and constructors
		for k, v := range *sps {
			goioc.Register(k, v)
			log.Printf("registered service provider: %s\n", k)
		}
	}

	return nil
}
