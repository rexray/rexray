package gofig

import "github.com/akutz/gofig/types"

func testReg3() *configReg {
	r := newRegistration("Mock Provider")
	r.SetYAML(`mockProvider:
    userName: admin
    useCerts: true
    docker:
        MinVolSize: 16
`)
	r.Key(types.String, "", "admin", "", "mockProvider.userName")
	r.Key(types.String, "", "", "", "mockProvider.password")
	r.Key(types.Bool, "", false, "", "mockProvider.useCerts")
	r.Key(types.Int, "", 16, "", "mockProvider.docker.minVolSize")
	r.Key(types.Bool, "i", true, "", "mockProvider.insecure")
	r.Key(types.Int, "m", 256, "", "mockProvider.docker.maxVolSize")
	return r
}
