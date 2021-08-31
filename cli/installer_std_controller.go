// +build !client
// +build !agent

package cli

import (
	"fmt"
	"os"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"

	"github.com/AVENTER-UG/rexray/libstorage/api/context"
	apitypes "github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/util"
)

func installSelfCert(
	ctx apitypes.Context,
	config gofig.Config) error {

	certPath := config.GetString(apitypes.ConfigTLSCertFile)
	keyPath := config.GetString(apitypes.ConfigTLSKeyFile)

	if gotil.FileExists(certPath) && gotil.FileExists(keyPath) {
		ctx.WithFields(map[string]interface{}{
			"certFile": certPath,
			"keyFile":  keyPath,
		}).Debug("not creating certs; files already exist")
		return nil
	}

	pathConfig := context.MustPathConfig(ctx)
	certPath = pathConfig.DefaultTLSCertFile
	keyPath = pathConfig.DefaultTLSKeyFile

	if gotil.FileExists(certPath) && gotil.FileExists(keyPath) {
		ctx.WithFields(map[string]interface{}{
			"certFile": certPath,
			"keyFile":  keyPath,
		}).Debug("not creating certs; files already exist")
		return nil
	}

	host, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("get hostname for cert failed: %v", err)
	}

	ctx.WithFields(map[string]interface{}{
		"host":     host,
		"certFile": certPath,
		"keyFile":  keyPath,
	}).Debug("creating certs")

	fmt.Println("generating self-signed certificate...")
	if err := util.CreateSelfCert(ctx, certPath, keyPath, host); err != nil {
		return fmt.Errorf("cert generation failed: %v", err)
	}

	fmt.Printf("  %s\n  %s\n", certPath, keyPath)
	return nil
}
