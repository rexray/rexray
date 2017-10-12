package csi_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	gofigCore "github.com/akutz/gofig"
	gofig "github.com/akutz/gofig/types"
)

func init() {
	r := gofigCore.NewRegistration("MapTest")
	r.Key(gofig.String,
		"", "", "",
		"nfs.volumes")
	gofigCore.Register(r)
}

func TestGofigMap(t *testing.T) {
	config := gofigCore.NewConfig(false, false, "config", "yaml")
	config.ReadConfig(strings.NewReader(`
nfs:
  volumes:
    - name1=uri1
    - name2=uri2
    - name3=uri3
`))
	asSlice := config.GetStringSlice("nfs.volumes")
	fmt.Printf("GetStringSlice: len=%d, %v\n", len(asSlice), asSlice)

	os.Setenv("NFS_VOLUMES", "name4=uri4 name5=uri5")
	asSlice = config.GetStringSlice("nfs.volumes")
	fmt.Printf("GetStringSlice: len=%d, %v\n", len(asSlice), asSlice)
}
