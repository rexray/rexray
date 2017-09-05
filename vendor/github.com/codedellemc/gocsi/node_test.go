package gocsi_test

import (
	"context"

	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
)

var _ = Describe("Node", func() {
	var (
		err      error
		stopMock func()
		ctx      context.Context
		gclient  *grpc.ClientConn
		client   csi.NodeClient
	)
	BeforeEach(func() {
		ctx = context.Background()
		gclient, stopMock, err = startMockServer(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		client = csi.NewNodeClient(gclient)
	})
	AfterEach(func() {
		ctx = nil
		gclient.Close()
		gclient = nil
		client = nil
		stopMock()
	})

	Describe("GetNodeID", func() {
		var nodeID *csi.NodeID
		BeforeEach(func() {
			nodeID, err = gocsi.GetNodeID(
				ctx,
				client,
				mockSupportedVersions[0])
		})
		It("Should Be Valid", func() {
			Ω(err).ShouldNot(HaveOccurred())
			Ω(nodeID).ShouldNot(BeNil())
			Ω(nodeID.Values).Should(HaveLen(1))
			Ω(nodeID.Values["id"]).Should(Equal(pluginName))
		})
	})
})
