// +build !rexray_build_type_client
// +build !rexray_build_type_agent

package cli

import (
	"fmt"
	"os"

	gofig "github.com/akutz/gofig/types"
	"github.com/akutz/gotil"

	"github.com/codedellemc/rexray/libstorage/api/context"
	apitypes "github.com/codedellemc/rexray/libstorage/api/types"
	"github.com/codedellemc/rexray/util"
)

func installSelfCert(ctx apitypes.Context, config gofig.Config) {

	certPath := config.GetString(apitypes.ConfigTLSCertFile)
	keyPath := config.GetString(apitypes.ConfigTLSKeyFile)

	if gotil.FileExists(certPath) && gotil.FileExists(keyPath) {
		ctx.WithFields(map[string]interface{}{
			"certFile": certPath,
			"keyFile":  keyPath,
		}).Debug("not creating certs; files already exist")
		return
	}

	pathConfig := context.MustPathConfig(ctx)
	certPath = pathConfig.DefaultTLSCertFile
	keyPath = pathConfig.DefaultTLSKeyFile

	if gotil.FileExists(certPath) && gotil.FileExists(keyPath) {
		ctx.WithFields(map[string]interface{}{
			"certFile": certPath,
			"keyFile":  keyPath,
		}).Debug("not creating certs; files already exist")
		return
	}

	host, err := os.Hostname()
	if err != nil {
		ctx.Fatalf("failed to get hostname for cert")
	}

	ctx.WithFields(map[string]interface{}{
		"host":     host,
		"certFile": certPath,
		"keyFile":  keyPath,
	}).Debug("creating certs")

	fmt.Println("Generating server self-signed certificate...")
	if err := util.CreateSelfCert(ctx, certPath, keyPath, host); err != nil {
		ctx.WithError(err).Fatal("cert generation failed")
	}

	fmt.Printf("Created cert file %s, key %s\n\n", certPath, keyPath)
}
