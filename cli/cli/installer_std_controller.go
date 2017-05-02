// +build !rexray_build_type_client
// +build !rexray_build_type_agent

package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	gofig "github.com/akutz/gofig/types"
	apitypes "github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/rexray/util"
)

func installSelfCert(ctx apitypes.Context, config gofig.Config) {
	certPath := config.GetString(apitypes.ConfigTLSCertFile)
	keyPath := config.GetString(apitypes.ConfigTLSKeyFile)
	host := "127.0.0.1"

	fmt.Println("Generating server self-signed certificate...")
	if err := util.CreateSelfCert(ctx, certPath, keyPath, host); err != nil {
		log.Fatalf("cert generation failed: %v\n", err)
	}

	fmt.Printf("Created cert file %s, key %s\n\n", certPath, keyPath)
}
