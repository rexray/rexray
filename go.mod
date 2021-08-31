module github.com/AVENTER-UG/rexray

go 1.15

require (
	cloud.google.com/go v0.12.0 // indirect
	github.com/akutz/gofig v0.1.9
	github.com/Azure/azure-sdk-for-go v7.0.1-beta+incompatible
	github.com/Azure/go-autorest v7.2.5+incompatible
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/go-winio v0.4.5 // indirect
	github.com/SermoDigital/jose v0.0.0-20161205224733-f6df55f235c2
	github.com/akutz/go-bindata v0.0.0-20160412065030-1dd44b25b79c // indirect
	github.com/akutz/golf v0.1.3
	github.com/akutz/goof v0.1.2
	github.com/akutz/gotil v0.1.0
	github.com/akutz/gournal v0.5.0
	github.com/akutz/logrus v0.0.0-20170830210741-d842de504ca8 // indirect
	github.com/akutz/yaml v0.0.0-20160725221316-bc35f417f8a7 // indirect
	github.com/appropriate/go-virtualboxclient v0.0.0-20150929060730-83f77bb344f9 // indirect
	github.com/asaskevich/govalidator v0.0.0-20170425121227-4918b99a7cb9 // indirect
	github.com/aws/aws-sdk-go v1.12.46
	github.com/cesanta/ucl v0.0.0-20150604132806-97c016fce90e
	github.com/cesanta/validate-json v0.0.0-20150603122804-2f16017c76fc
	github.com/clintonskitson/go-virtualboxclient v0.0.0-20151220033032-e0978ab2ed40 // indirect
	github.com/codenrhoden/go-vhd v0.0.0-20170208185941-96a0db67ea82 // indirect
	github.com/coreos/go-systemd v0.0.0-20170731111925-d21964639418 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgrijalva/jwt-go v3.0.0+incompatible // indirect
	github.com/digitalocean/godo v1.2.0
	github.com/docker/go-connections v0.3.0 // indirect
	github.com/docker/go-plugins-helpers v0.0.0-20170817192157-a9ef19c479cb
	github.com/fatih/color v1.12.0 // indirect
	github.com/fsnotify/fsnotify v1.4.2 // indirect
	github.com/go-ini/ini v1.28.2 // indirect
	github.com/google/go-querystring v0.0.0-20170111101155-53e6ce116135 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20170916161221-b4c2377fa779
	github.com/gorilla/context v0.0.0-20160226214623-1ea25387ff6f
	github.com/gorilla/mux v1.4.0
	github.com/hashicorp/hcl v0.0.0-20170825171336-8f6b1344a92f // indirect
	github.com/inconshreveable/mousetrap v0.0.0-20141017200713-76626ae9c91c // indirect
	github.com/jmespath/go-jmespath v0.0.0-20160202185014-0b12d6b521d8 // indirect
	github.com/jteeuwen/go-bindata v3.0.7+incompatible
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1 // indirect
	github.com/magiconair/properties v1.7.3 // indirect
	github.com/mitchellh/mapstructure v0.0.0-20170523030023-d0303fe80992 // indirect
	github.com/onsi/ginkgo v1.4.0
	github.com/onsi/gomega v1.2.0
	github.com/pelletier/go-buffruneio v0.2.0 // indirect
	github.com/pelletier/go-toml v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rubiojr/go-vhd v0.0.0-20200706122120-ccecf6c0760f
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/soheilhy/cmux v0.1.3
	github.com/spf13/afero v0.0.0-20170901052352-ee1bd8ee15a1 // indirect
	github.com/spf13/cast v1.1.0 // indirect
	github.com/spf13/cobra v0.0.0-20170905172051-b78744579491
	github.com/spf13/jwalterweatherman v0.0.0-20170901151539-12bd96e66386 // indirect
	github.com/spf13/pflag v0.0.0-20170901120850-7aff26db30c1
	github.com/spf13/viper v1.0.0 // indirect
	github.com/stretchr/testify v1.2.2
	github.com/tent/http-link-go v0.0.0-20130702225549-ac974c61c2f9 // indirect
	github.com/thecodeteam/csi-blockdevices v0.0.0-20171027052650-70757e2170eb
	github.com/thecodeteam/csi-nfs v0.0.0-20171027052018-d2e5dceda9b1
	github.com/thecodeteam/csi-vfs v0.2.0
	github.com/thecodeteam/gocsi v0.0.0-20171026153342-83f15105815b
	github.com/thecodeteam/goioc v0.2.0
	github.com/thecodeteam/goisilon v1.7.0
	github.com/thecodeteam/goscaleio v0.0.0-20171027002729-35ca2e98889a
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190603091049-60506f45cf65
	golang.org/x/oauth2 v0.0.0-20170901193052-d89af98d7c6b
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	google.golang.org/api v0.0.0-20170906000354-38eaa396bab4
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20170904050139-595979c8a7bf // indirect
	google.golang.org/grpc v1.6.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/akutz/gofig v0.1.9 => github.com/AVENTER-UG/gofig v0.1.10
)
