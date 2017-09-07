package tests

import (
	"fmt"
	"testing"

	// load the ginko test package
	_ "github.com/onsi/ginkgo"
)

// RunSuite executes the complete test suite for a driver.
func RunSuite(t *testing.T, driver string) {

	// the driver's tests using a tcp server
	Describe(newSuiteRunner(t, driver, "tcp"))

	// the driver's tests using a unix socket server
	Describe(newSuiteRunner(t, driver, "unix"))

	RegisterFailHandler(Fail)
	RunSpecs(t, fmt.Sprintf("%s Suite", driver))
}

func (sr *suiteRunner) Describe() {

	var t *testRunner

	BeforeEach(func() {
		t = newTestRunner(sr.driverName)
		t.beforeEach()
	})
	AfterEach(func() {
		t.afterEach()
		t = nil
	})
	JustBeforeEach(func() {
		t.justBeforeEach()
	})

	Describe("w server", func() {
		BeforeEach(func() {
			switch sr.proto {
			case "tcp":
				t.initTCPSocket()
			case "unix":
				t.initUNIXSocket()
			}
			t.initConfigData()
		})
		JustBeforeEach(func() {
			t.initServer()
		})

		Context("w client", func() {
			JustBeforeEach(func() {
				t.justBeforeEachClientSpec()
			})

			It("list root resources", func() {
				t.itClientSpecListRootResources()
			})
			It("list all volumes for all services", func() {
				t.itClientSpecListVolumes()
			})

			Context("w service", func() {
				BeforeEach(func() {
					t.beforeEachClientServiceSpec()
				})

				It("preserves local devices with empty values", func() {
					t.itPreservesLocalDevicesEmptyVals()
				})
				It("list volumes", func() {
					t.itClientSvcSpecListVolumes()
				})
				It("inspect a missing volume", func() {
					t.itClientSvcSpecInspectMissingVolume()
				})
				It("create a new volume", func() {
					t.itClientSvcSpecCreateVolume()
				})

				Context("w volume", func() {
					AfterEach(func() {
						t.afterEachVolumeSpec()
					})
					JustBeforeEach(func() {
						t.justBeforeEachVolumeSpec()
					})

					It("inspect the volume", func() {
						t.itVolumeSpecInspect()
					})
					It("delete the volume", func() {
						t.itVolumeSpecDelete()
					})
					It("attach the volume", func() {
						t.itVolumeSpecAttach()
					})

					Context("that is attached", func() {
						JustBeforeEach(func() {
							t.justBeforeEachAttVolumeSpec()
						})

						It("inspect w att req", func() {
							t.itAttVolumeSpecInspect()
						})
						It("inspect w att req - avai", func() {
							t.itAttVolumeSpecInspectAvai()
						})
						It("inspect w att req - att or avai", func() {
							t.itAttVolumeSpecInspectAttOrAvai()
						})
						It("deatch", func() {
							t.itAttVolumeSpecDetach()
						})
					}) // Context w att volume

				}) // Context w volume

			}) // Context w service

		}) // Context w client

		// withClients is used to validate the following tls connections
		withClients := func() {
			Context("w client", func() {
				JustBeforeEach(func() {
					t.justBeforeEachClientSpec()
				})
				It("list root resources", func() {
					t.itClientSpecListRootResources()
				})
			}) // Context w client
			Context("w non-tls client", func() {
				JustBeforeEach(func() {
					t.justBeforeEachTLSClientNoTLS()
				})
				It("client should fail", func() {
					t.itTLSClientError()
				})
			}) // Context w non-tls client
		}

		Describe("w tls", func() {
			BeforeEach(func() {
				t.beforeEachTLS()
			})
			withClients()
		}) // Describe w tls

		Describe("w tls auto config", func() {
			BeforeEach(func() {
				t.beforeEachTLSAuto()
			})
			withClients()
		}) // Describe w tls auto config

		Describe("w tls insecure config", func() {
			BeforeEach(func() {
				t.beforeEachTLSInsecure()
			})
			withClients()
		}) // Describe w tls insecure config

		// withRemovedKnownHostsFile is used to validate the below
		// tests that involve configuring TLS using a known hosts file.
		withRemovedKnownHostsFile := func() {
			Context("w removed known_hosts file", func() {
				JustBeforeEach(func() {
					t.justBeforeEachTLSRemoveKnownHosts()
				})
				Specify("client should fail", func() {
					t.itTLSClientError()
				})
			}) // Context w removed known_hosts file
		}

		Describe("w tls known_hosts", func() {
			BeforeEach(func() {
				t.beforeEachTLSKnownHosts()
			})
			withClients()
			withRemovedKnownHostsFile()
		}) // Describe w tls known_hosts

		Describe("w tls auto known_hosts", func() {
			BeforeEach(func() {
				t.beforeEachTLSKnownHostsAuto()
			})
			withClients()
			withRemovedKnownHostsFile()
		}) // Describe w tls auto known_hosts

		Describe("w tls auto known_hosts & auto config", func() {
			BeforeEach(func() {
				t.beforeEachTLSAutoKnownHostsAuto()
			})
			withClients()
			withRemovedKnownHostsFile()
		}) // Describe w tls auto known_hosts & auto config

	}) // Describe w server
}
