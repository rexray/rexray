package gocsi_test

import (
	"os"

	"github.com/codedellemc/gocsi"
)

var _ = Describe("ParseVersion", func() {
	shouldParse := func() gocsi.Version {
		v, err := gocsi.ParseVersion(
			CurrentGinkgoTestDescription().ComponentTexts[1])
		Ω(err).ShouldNot(HaveOccurred())
		Ω(v).ShouldNot(BeNil())
		return v
	}
	Context("0.0.0", func() {
		It("Should Parse", func() {
			v := shouldParse()
			Ω(v.GetMajor()).Should(Equal(uint32(0)))
			Ω(v.GetMinor()).Should(Equal(uint32(0)))
			Ω(v.GetPatch()).Should(Equal(uint32(0)))
		})
	})
	Context("0.1.0", func() {
		It("Should Parse", func() {
			v := shouldParse()
			Ω(v.GetMajor()).Should(Equal(uint32(0)))
			Ω(v.GetMinor()).Should(Equal(uint32(1)))
			Ω(v.GetPatch()).Should(Equal(uint32(0)))
		})
	})
	Context("1.1.0", func() {
		It("Should Parse", func() {
			v := shouldParse()
			Ω(v.GetMajor()).Should(Equal(uint32(1)))
			Ω(v.GetMinor()).Should(Equal(uint32(1)))
			Ω(v.GetPatch()).Should(Equal(uint32(0)))
		})
	})
})

var _ = Describe("GetCSIEndpoint", func() {
	var (
		err         error
		proto       string
		addr        string
		expEndpoint string
		expProto    string
		expAddr     string
	)
	BeforeEach(func() {
		expEndpoint = CurrentGinkgoTestDescription().ComponentTexts[2]
		os.Setenv(gocsi.CSIEndpoint, expEndpoint)
	})
	AfterEach(func() {
		proto = ""
		addr = ""
		expEndpoint = ""
		expProto = ""
		expAddr = ""
		os.Unsetenv(gocsi.CSIEndpoint)
	})
	JustBeforeEach(func() {
		proto, addr, err = gocsi.GetCSIEndpoint()
	})

	Context("Valid Endpoint", func() {
		shouldBeValid := func() {
			Ω(os.Getenv(gocsi.CSIEndpoint)).Should(Equal(expEndpoint))
			Ω(proto).Should(Equal(expProto))
			Ω(addr).Should(Equal(expAddr))
		}
		Context("unix://path/to/sock.sock", func() {
			BeforeEach(func() {
				expProto = "unix"
				expAddr = "path/to/sock.sock"
			})
			It("Should Be Valid", shouldBeValid)
		})
		Context("unix:///path/to/sock.sock", func() {
			BeforeEach(func() {
				expProto = "unix"
				expAddr = "/path/to/sock.sock"
			})
			It("Should Be Valid", shouldBeValid)
		})
	})

	Context("Missing Endpoint", func() {
		Context("", func() {
			It("Should Be Missing", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(gocsi.ErrMissingCSIEndpoint))
			})
		})
	})

	Context("Invalid Endpoint", func() {
		shouldBeInvalid := func() {
			Ω(err).Should(HaveOccurred())
			Ω(err).Should(Equal(gocsi.ErrInvalidCSIEndpoint))
		}
		Context("tcp5://localhost:5000", func() {
			It("Should Be An Invalid Endpoint", shouldBeInvalid)
		})
		Context("unixpcket://path/to/sock.sock", func() {
			It("Should Be An Invalid Endpoint", shouldBeInvalid)
		})
	})
})
