package utils_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/akutz/gotil"
	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
	gomegaTypes "github.com/onsi/gomega/types"
)

var _ = Describe("Paths", func() {
	var homeDir string
	var userHome = gotil.HomeDir()

	BeforeSuite(func() {
		homeDir = gotil.HomeDir()
	})

	Describe("Creating a new paths config", func() {

		var tmpDir string

		BeforeEach(func() {
			td, err := ioutil.TempDir("", "")
			if err != nil {
				panic(err)
			}
			tmpDir = td
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		Context("With empty args", func() {

			var pathConfig *types.PathConfig

			BeforeEach(func() {
				pathConfig = utils.NewPathConfig()
			})
			AfterEach(func() {
				pathConfig = nil
			})

			It("Home should be $HOME/.libstorage", func() {
				Ω(pathConfig.Home).To(Σ(homeDir, ".libstorage"))
			})
			It("Etc should be $HOME/.libstorage/etc", func() {
				Ω(pathConfig.Etc).To(Σ(homeDir, ".libstorage", "etc"))
			})
			It("TLS should be $HOME/.libstorage/etc/tls", func() {
				Ω(pathConfig.TLS).To(Σ(homeDir, ".libstorage", "etc", "tls"))
			})
			It("Lib should be $HOME/.libstorage/var/lib", func() {
				Ω(pathConfig.Lib).To(Σ(homeDir, ".libstorage", "var", "lib"))
			})
			It("Log should be $HOME/.libstorage/var/log", func() {
				Ω(pathConfig.Log).To(Σ(homeDir, ".libstorage", "var", "log"))
			})
			It("Run should be $HOME/.libstorage/var/run", func() {
				Ω(pathConfig.Run).To(Σ(homeDir, ".libstorage", "var", "run"))
			})
			It("tls crt should be $HOME/.libstorage/etc/tls/libstorage.crt",
				func() {
					Ω(pathConfig.DefaultTLSCertFile).To(Σ(
						homeDir, ".libstorage", "etc", "tls", "libstorage.crt"))
				})
			It("tls key should be $HOME/.libstorage/etc/tls/libstorage.key",
				func() {
					Ω(pathConfig.DefaultTLSKeyFile).To(Σ(
						homeDir, ".libstorage", "etc", "tls", "libstorage.key"))
				})
			It("tls cacerts should be $HOME/.libstorage/etc/tls/cacerts",
				func() {
					Ω(pathConfig.DefaultTLSTrustedRootsFile).To(Σ(
						homeDir, ".libstorage", "etc", "tls", "cacerts"))
				})
			It("tls known_hosts should be "+
				"$HOME/.libstorage/etc/tls/known_hosts",
				func() {
					Ω(pathConfig.DefaultTLSKnownHosts).To(Σ(
						homeDir, ".libstorage", "etc", "tls", "known_hosts"))
				})
			It("usr home should be $HOME", func() {
				Ω(pathConfig.UserHome).To(Σ(userHome, ".libstorage"))
			})
			It("usr known_hosts should be $HOME/.libstorage/known_hosts",
				func() {
					Ω(pathConfig.UserDefaultTLSKnownHosts).To(
						Σ(userHome, ".libstorage", "known_hosts"))
				})
		})

		Context("With home", func() {

			var pathConfig *types.PathConfig

			BeforeEach(func() {
				pathConfig = utils.NewPathConfig(tmpDir, "")
			})
			AfterEach(func() {
				pathConfig = nil
			})

			It("Home should be tmpDir", func() {
				Ω(pathConfig.Home).To(Σ(tmpDir))
			})
			It("Etc should be tmpDir/etc", func() {
				Ω(pathConfig.Etc).To(Σ(tmpDir, "etc"))
			})
			It("TLS should be tmpDir/etc/tls", func() {
				Ω(pathConfig.TLS).To(Σ(tmpDir, "etc", "tls"))
			})
			It("Lib should be tmpDir/var/lib", func() {
				Ω(pathConfig.Lib).To(Σ(tmpDir, "var", "lib"))
			})
			It("Log should be tmpDir/var/log", func() {
				Ω(pathConfig.Log).To(Σ(tmpDir, "var", "log"))
			})
			It("Run should be tmpDir/var/run", func() {
				Ω(pathConfig.Run).To(Σ(tmpDir, "var", "run"))
			})
			It("tls crt should be tmpDir/etc/tls/libstorage.crt",
				func() {
					Ω(pathConfig.DefaultTLSCertFile).To(Σ(
						tmpDir, "etc", "tls", "libstorage.crt"))
				})
			It("tls key should be tmpDir/etc/tls/libstorage.key",
				func() {
					Ω(pathConfig.DefaultTLSKeyFile).To(Σ(
						tmpDir, "etc", "tls", "libstorage.key"))
				})
			It("tls cacerts should be tmpDir/etc/tls/cacerts",
				func() {
					Ω(pathConfig.DefaultTLSTrustedRootsFile).To(Σ(
						tmpDir, "etc", "tls", "cacerts"))
				})
			It("tls known_hosts should be "+
				"tmpDir/etc/tls/known_hosts",
				func() {
					Ω(pathConfig.DefaultTLSKnownHosts).To(Σ(
						tmpDir, "etc", "tls", "known_hosts"))
				})
			It("usr home should be $HOME", func() {
				Ω(pathConfig.UserHome).To(Σ(userHome, ".libstorage"))
			})
			It("usr known_hosts should be $HOME/.libstorage/known_hosts",
				func() {
					Ω(pathConfig.UserDefaultTLSKnownHosts).To(
						Σ(userHome, ".libstorage", "known_hosts"))
				})
		})

		Context("With token", func() {

			var pathConfig *types.PathConfig

			BeforeEach(func() {
				pathConfig = utils.NewPathConfig("", "rexray")
			})
			AfterEach(func() {
				pathConfig = nil
			})

			It("Home should be $HOME/.rexray", func() {
				Ω(pathConfig.Home).To(Σ(homeDir, ".rexray"))
			})
			It("Etc should be $HOME/.rexray/etc", func() {
				Ω(pathConfig.Etc).To(Σ(homeDir, ".rexray", "etc"))
			})
			It("TLS should be $HOME/.rexray/etc/tls", func() {
				Ω(pathConfig.TLS).To(Σ(homeDir, ".rexray", "etc", "tls"))
			})
			It("Lib should be $HOME/.rexray/var/lib", func() {
				Ω(pathConfig.Lib).To(Σ(homeDir, ".rexray", "var", "lib"))
			})
			It("Log should be $HOME/.rexray/var/log", func() {
				Ω(pathConfig.Log).To(Σ(homeDir, ".rexray", "var", "log"))
			})
			It("Run should be $HOME/.rexray/var/run", func() {
				Ω(pathConfig.Run).To(Σ(homeDir, ".rexray", "var", "run"))
			})
			It("tls crt should be $HOME/.rexray/etc/tls/rexray.crt",
				func() {
					Ω(pathConfig.DefaultTLSCertFile).To(Σ(
						homeDir, ".rexray", "etc", "tls", "rexray.crt"))
				})
			It("tls key should be $HOME/.rexray/etc/tls/rexray.key",
				func() {
					Ω(pathConfig.DefaultTLSKeyFile).To(Σ(
						homeDir, ".rexray", "etc", "tls", "rexray.key"))
				})
			It("tls cacerts should be $HOME/.rexray/etc/tls/cacerts",
				func() {
					Ω(pathConfig.DefaultTLSTrustedRootsFile).To(Σ(
						homeDir, ".rexray", "etc", "tls", "cacerts"))
				})
			It("tls known_hosts should be "+
				"$HOME/.rexray/etc/tls/known_hosts",
				func() {
					Ω(pathConfig.DefaultTLSKnownHosts).To(Σ(
						homeDir, ".rexray", "etc", "tls", "known_hosts"))
				})
			It("usr home should be $HOME", func() {
				Ω(pathConfig.UserHome).To(Σ(userHome, ".rexray"))
			})
			It("usr known_hosts should be $HOME/.rexray/known_hosts",
				func() {
					Ω(pathConfig.UserDefaultTLSKnownHosts).To(
						Σ(userHome, ".rexray", "known_hosts"))
				})
		})

		Context("With token & TOKEN_HOME_ETC_TLS", func() {

			var pathConfig *types.PathConfig

			BeforeEach(func() {
				Ω(tmpDir).ShouldNot(BeEmpty())
				os.Setenv("REXRAY_HOME_ETC_TLS", tmpDir)
				pathConfig = utils.NewPathConfig("", "rexray")
			})
			AfterEach(func() {
				os.Setenv("REXRAY_HOME_ETC_TLS", "")
				pathConfig = nil
			})

			It("Home should be $HOME/.rexray", func() {
				Ω(pathConfig.Home).To(Σ(homeDir, ".rexray"))
			})
			It("Etc should be $HOME/.rexray/etc", func() {
				Ω(pathConfig.Etc).To(Σ(homeDir, ".rexray", "etc"))
			})
			It("TLS should be tmpDir", func() {
				Ω(pathConfig.TLS).To(Σ(tmpDir))
			})
			It("Lib should be $HOME/.rexray/var/lib", func() {
				Ω(pathConfig.Lib).To(Σ(homeDir, ".rexray", "var", "lib"))
			})
			It("Log should be $HOME/.rexray/var/log", func() {
				Ω(pathConfig.Log).To(Σ(homeDir, ".rexray", "var", "log"))
			})
			It("Run should be $HOME/.rexray/var/run", func() {
				Ω(pathConfig.Run).To(Σ(homeDir, ".rexray", "var", "run"))
			})
			It("tls crt should be tmpDir/rexray.crt",
				func() {
					Ω(pathConfig.DefaultTLSCertFile).To(Σ(
						tmpDir, "rexray.crt"))
				})
			It("tls key should be tmpDir/rexray.key",
				func() {
					Ω(pathConfig.DefaultTLSKeyFile).To(Σ(
						tmpDir, "rexray.key"))
				})
			It("tls cacerts should be tmpDir/cacerts",
				func() {
					Ω(pathConfig.DefaultTLSTrustedRootsFile).To(Σ(
						tmpDir, "cacerts"))
				})
			It("tls known_hosts should be tmpDir/known_hosts",
				func() {
					Ω(pathConfig.DefaultTLSKnownHosts).To(Σ(
						tmpDir, "known_hosts"))
				})
			It("usr home should be $HOME", func() {
				Ω(pathConfig.UserHome).To(Σ(userHome, ".rexray"))
			})
			It("usr known_hosts should be $HOME/.rexray/known_hosts",
				func() {
					Ω(pathConfig.UserDefaultTLSKnownHosts).To(
						Σ(userHome, ".rexray", "known_hosts"))
				})
		})

		Context("With token & TOKEN_HOME", func() {

			var pathConfig *types.PathConfig

			BeforeEach(func() {
				Ω(tmpDir).ShouldNot(BeEmpty())
				os.Setenv("REXRAY_HOME", tmpDir)
				pathConfig = utils.NewPathConfig("", "rexray")
			})
			AfterEach(func() {
				os.Setenv("REXRAY_HOME", "")
				pathConfig = nil
			})

			It("Home should be tmpDir", func() {
				Ω(pathConfig.Home).To(Σ(tmpDir))
			})
			It("Etc should be tmpDir/etc", func() {
				Ω(pathConfig.Etc).To(Σ(tmpDir, "etc"))
			})
			It("TLS should be tmpDir/etc/tls", func() {
				Ω(pathConfig.TLS).To(Σ(tmpDir, "etc", "tls"))
			})
			It("Lib should be tmpDir/var/lib", func() {
				Ω(pathConfig.Lib).To(Σ(tmpDir, "var", "lib"))
			})
			It("Log should be tmpDir/var/log", func() {
				Ω(pathConfig.Log).To(Σ(tmpDir, "var", "log"))
			})
			It("Run should be tmpDir/var/run", func() {
				Ω(pathConfig.Run).To(Σ(tmpDir, "var", "run"))
			})
			It("tls crt should be tmpDir/etc/tls/rexray.crt",
				func() {
					Ω(pathConfig.DefaultTLSCertFile).To(Σ(
						tmpDir, "etc", "tls", "rexray.crt"))
				})
			It("tls key should be tmpDir/etc/tls/rexray.key",
				func() {
					Ω(pathConfig.DefaultTLSKeyFile).To(Σ(
						tmpDir, "etc", "tls", "rexray.key"))
				})
			It("tls cacerts should be tmpDir/etc/tls/cacerts",
				func() {
					Ω(pathConfig.DefaultTLSTrustedRootsFile).To(Σ(
						tmpDir, "etc", "tls", "cacerts"))
				})
			It("tls known_hosts should be tmpDir/etc/tls/known_hosts",
				func() {
					Ω(pathConfig.DefaultTLSKnownHosts).To(Σ(
						tmpDir, "etc", "tls", "known_hosts"))
				})
			It("usr home should be $HOME", func() {
				Ω(pathConfig.UserHome).To(Σ(userHome, ".rexray"))
			})
			It("usr known_hosts should be $HOME/.rexray/known_hosts",
				func() {
					Ω(pathConfig.UserDefaultTLSKnownHosts).To(
						Σ(userHome, ".rexray", "known_hosts"))
				})
		})

		Context("With LIBSTORAGE_HOME", func() {

			var pathConfig *types.PathConfig

			BeforeEach(func() {
				Ω(tmpDir).ShouldNot(BeEmpty())
				os.Setenv("LIBSTORAGE_HOME", tmpDir)
				pathConfig = utils.NewPathConfig("", "")
			})
			AfterEach(func() {
				os.Setenv("LIBSTORAGE_HOME", "")
				pathConfig = nil
			})

			It("Home should be tmpDir", func() {
				Ω(pathConfig.Home).To(Σ(tmpDir))
			})
			It("Etc should be tmpDir/etc", func() {
				Ω(pathConfig.Etc).To(Σ(tmpDir, "etc"))
			})
			It("TLS should be tmpDir/etc/tls", func() {
				Ω(pathConfig.TLS).To(Σ(tmpDir, "etc", "tls"))
			})
			It("Lib should be tmpDir/var/lib", func() {
				Ω(pathConfig.Lib).To(Σ(tmpDir, "var", "lib"))
			})
			It("Log should be tmpDir/var/log", func() {
				Ω(pathConfig.Log).To(Σ(tmpDir, "var", "log"))
			})
			It("Run should be tmpDir/var/run", func() {
				Ω(pathConfig.Run).To(Σ(tmpDir, "var", "run"))
			})
			It("tls crt should be tmpDir/etc/tls/libstorage.crt",
				func() {
					Ω(pathConfig.DefaultTLSCertFile).To(Σ(
						tmpDir, "etc", "tls", "libstorage.crt"))
				})
			It("tls key should be tmpDir/etc/tls/libstorage.key",
				func() {
					Ω(pathConfig.DefaultTLSKeyFile).To(Σ(
						tmpDir, "etc", "tls", "libstorage.key"))
				})
			It("tls cacerts should be tmpDir/etc/tls/cacerts",
				func() {
					Ω(pathConfig.DefaultTLSTrustedRootsFile).To(Σ(
						tmpDir, "etc", "tls", "cacerts"))
				})
			It("tls known_hosts should be tmpDir/etc/tls/known_hosts",
				func() {
					Ω(pathConfig.DefaultTLSKnownHosts).To(Σ(
						tmpDir, "etc", "tls", "known_hosts"))
				})
			It("usr home should be $HOME", func() {
				Ω(pathConfig.UserHome).To(Σ(userHome, ".libstorage"))
			})
			It("usr known_hosts should be $HOME/.libstorage/known_hosts",
				func() {
					Ω(pathConfig.UserDefaultTLSKnownHosts).To(
						Σ(userHome, ".libstorage", "known_hosts"))
				})
		})
	})
})

type pathMatcher struct {
	expected string
}

func (pm *pathMatcher) Match(actual interface{}) (bool, error) {
	szA, ok := actual.(string)
	if !ok {
		return false, errors.New("pathMatcher expects one or more strings")
	}
	return pm.expected == szA, nil
}
func (pm *pathMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n\t%#v\nto be equal to\n\t%#v", actual, pm.expected)
}
func (pm *pathMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected\n\t%#v\nnot to be equal to\n\t%#v", actual, pm.expected)
}
func Σ(paths ...string) gomegaTypes.GomegaMatcher {
	return &pathMatcher{path.Join(paths...)}
}
