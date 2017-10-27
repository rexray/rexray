package gocsi_test

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
)

var _ = Describe("Identity", func() {
	var (
		err      error
		stopMock func()
		ctx      context.Context
		gclient  *grpc.ClientConn
		client   csi.IdentityClient
	)
	BeforeEach(func() {
		ctx = context.Background()
		gclient, stopMock, err = startMockServer(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		client = csi.NewIdentityClient(gclient)
	})
	AfterEach(func() {
		ctx = nil
		gclient.Close()
		gclient = nil
		client = nil
		stopMock()
	})

	Describe("GetPluginInfo", func() {
		var (
			res     *csi.GetPluginInfoResponse_Result
			version gocsi.Version
		)
		BeforeEach(func() {
			version, err = gocsi.ParseVersion(CTest().ComponentTexts[3])
			Ω(err).ShouldNot(HaveOccurred())
			res, err = gocsi.GetPluginInfo(
				ctx,
				client,
				&csi.Version{
					Major: version.GetMajor(),
					Minor: version.GetMinor(),
					Patch: version.GetPatch(),
				})
		})
		shouldBeValid := func() {
			Ω(err).ShouldNot(HaveOccurred())
			Ω(res).ShouldNot(BeNil())
			Ω(res.Name).Should(Equal(pluginName))
			Ω(res.VendorVersion).Should(Equal(CTest().ComponentTexts[3]))
		}
		shouldNotBeValid := func() {
			Ω(res).Should(BeNil())
			Ω(err).Should(HaveOccurred())
			Ω(err).Should(Σ(&gocsi.Error{
				FullMethod: "/csi.Identity/GetPluginInfo",
				Code:       2,
				Description: fmt.Sprintf("unsupported request version: %s",
					CTest().ComponentTexts[3]),
			}))
		}
		Context("With Request Version", func() {
			Context("0.0.0", func() {
				It("Should Not Be Valid", shouldNotBeValid)
			})
			Context("0.1.0", func() {
				It("Should Be Valid", shouldBeValid)
			})
			Context("0.2.0", func() {
				It("Should Be Valid", shouldBeValid)
			})
			Context("1.0.0", func() {
				It("Should Be Valid", shouldBeValid)
			})
			Context("1.1.0", func() {
				It("Should Be Valid", shouldBeValid)
			})
			Context("1.2.0", func() {
				It("Should Not Be Valid", shouldNotBeValid)
			})
		})
	})

	Describe("GetSupportedVersions", func() {
		It("Should Be Valid", func() {
			res, err := gocsi.GetSupportedVersions(
				ctx,
				client)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(res).ShouldNot(BeNil())
			Ω(res).Should(HaveLen(len(mockSupportedVersions)))
			Ω(res).Should(Equal(mockSupportedVersions))
		})
	})
})
