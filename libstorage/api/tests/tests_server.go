package tests

import (
	"fmt"

	"github.com/akutz/gotil"

	apiserver "github.com/AVENTER-UG/rexray/libstorage/api/server"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
)

func (t *testRunner) initTCPSocket() {
	t.proto = "tcp"
	t.laddr = fmt.Sprintf("127.0.0.1:%d", gotil.RandomTCPPort())
}
func (t *testRunner) initUNIXSocket() {
	t.proto = protoUnix
	t.laddr = utils.GetTempSockFile(t.ctx)
}

func (t *testRunner) initServer() {
	var err error
	t.server, t.srvErr, err = apiserver.Serve(t.ctx, t.config)
	Ω(err).ToNot(HaveOccurred())
	go func() {
		defer GinkgoRecover()
		if err = <-t.srvErr; err != nil {
			Fail(err.Error())
		}
	}()
}
