# Build Reference

How to build libStorage

---

## Build Requirements
This project has very few build requirements, but there are still one or two
items of which to be aware. Also, please note that this are the requirements to
*build* `libStorage`, not run it.

Requirement | Version
------------|--------
Operating System | Linux, OS X
[Go](https://golang.org/) | >=1.6
[GNU Make](https://www.gnu.org/software/make/) | >=3.80
[Glide](https://glide.sh/) | >=0.10
[X-Code Command Line Tools (OS X only)](https://developer.apple.com/library/ios/technotes/tn2339/_index.html) | >= OS X 10.9
Linux Kernel Headers (Linux only) | >=Linux Kernel 3.13
[GNU C Compiler](https://gcc.gnu.org/) (Linux only) | >= 4.8

OS X ships with a very old version of GNU Make, and a package manager like
[Homebrew](http://brew.sh/) can be used to install the required version.

It's also possible to use GCC as the Cgo compiler for OS X or to use Clang on
Linux, but by default Clang is used on OS X and GCC on Linux.

## Cross-Compilation
This project's
[`Makefile`](https://github.com/emccode/libstorage/blob/master/Makefile)
is configured by default to cross-compile certain project components for both
Linux x86_64 and Darwin (OS X) x86_64. Therefore the build process will fail if
the local Go environment is not configured for cross-compilation. Please take a
minute to read this
[blog post](http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5)
regarding cross-compilation with Go >=1.5.

While *some* of this project's components are cross-compiled as part of a
standard build, the project as a whole is not. This is because the project
has key components that are dependent upon the Cgo compiler, and most build
systems do not possess the capability to cross-compile against the C stdlib
tool-chain.

## Basic Build
Building from source should be fairly straight-forward as the most basic build
can be achieved by executing `make` from the project's root directory.

```sh
$ make
make deps
make[1]: Entering directory './github.com/emccode/libstorage'
glide up && touch glide.lock.d
[INFO] Downloading dependencies. Please wait...
[INFO] Fetching updates for github.com/Sirupsen/logrus.
[INFO] Fetching updates for github.com/stretchr/testify.
[INFO] Fetching updates for github.com/akutz/golf.
[INFO] Fetching updates for github.com/akutz/gofig.
[INFO] Fetching updates for github.com/appropriate/go-virtualboxclient.
[INFO] Fetching updates for github.com/jteeuwen/go-bindata.
[INFO] Fetching updates for github.com/emccode/goisilon.
[INFO] Fetching updates for github.com/blang/semver.
[INFO] Fetching updates for github.com/cesanta/validate-json.
[INFO] Fetching updates for github.com/akutz/goof.
[INFO] Fetching updates for github.com/emccode/goscaleio.
[INFO] Fetching updates for github.com/akutz/gotil.
[INFO] Setting version for github.com/Sirupsen/logrus to feature/logrus-aware-types.
[INFO] Setting version for github.com/blang/semver to v3.0.1.
[INFO] Setting version for github.com/jteeuwen/go-bindata to feature/md5checksum.
[INFO] Setting version for github.com/emccode/goscaleio to 53ea76f52205380ab52b9c1f4ad89321c286bb95.
[INFO] Setting version for github.com/emccode/goisilon to f9b53f0aaadb12a26b134830142fc537f492cb13.
[INFO] Setting version for github.com/appropriate/go-virtualboxclient to e0978ab2ed407095400a69d5933958dd260058cd.
[INFO] Resolving imports
[INFO] Setting version for github.com/akutz/goof to master.
[INFO] Setting version for github.com/akutz/gotil to master.
[INFO] Fetching updates for github.com/spf13/pflag.
[INFO] Setting version for github.com/spf13/pflag to b084184666e02084b8ccb9b704bf0d79c466eb1d.
[INFO] Fetching updates for github.com/spf13/viper.
[INFO] Setting version for github.com/spf13/viper to support/rexray.
[INFO] Fetching updates for gopkg.in/yaml.v2.
[INFO] Setting version for gopkg.in/yaml.v2 to b4a9f8c4b84c6c4256d669c649837f1441e4b050.
[INFO] Fetching updates for golang.org/x/sys.
[INFO] Fetching updates for github.com/kardianos/osext.
[INFO] Setting version for github.com/kardianos/osext to master.
[INFO] Fetching updates for golang.org/x/net.
[INFO] Found Godeps.json file in vendor/github.com/stretchr/testify
[INFO] Fetching updates for github.com/davecgh/go-spew.
[INFO] Setting version for github.com/davecgh/go-spew to 5215b55f46b2b919f50a1df0eaa5886afe4e3b3d.
[INFO] Fetching updates for github.com/pmezard/go-difflib.
[INFO] Setting version for github.com/pmezard/go-difflib to d8ed2627bdf02c080bf22230dbb337003b7aba2d.
[INFO] Fetching updates for github.com/asaskevich/govalidator.
[INFO] Fetching updates for github.com/BurntSushi/toml.
[INFO] Fetching updates for github.com/kr/pretty.
[INFO] Fetching updates for github.com/magiconair/properties.
[INFO] Fetching updates for github.com/mitchellh/mapstructure.
[INFO] Fetching updates for github.com/spf13/cast.
[INFO] Fetching updates for github.com/spf13/jwalterweatherman.
[INFO] Fetching updates for gopkg.in/fsnotify.v1.
[INFO] Fetching updates for github.com/kr/text.
[INFO] Downloading dependencies. Please wait...
[INFO] Fetching updates for github.com/gorilla/mux.
[INFO] Fetching updates for github.com/cesanta/ucl.
[INFO] Fetching updates for github.com/gorilla/context.
[INFO] Setting references for remaining imports
[INFO] Project relies on 32 dependencies.
go install github.com/emccode/libstorage/vendor/github.com/jteeuwen/go-bindata/go-bindata
make[1]: Leaving directory './github.com/emccode/libstorage'
make build
make[1]: Entering directory './github.com/emccode/libstorage'
gcc -Wall -pedantic -std=c99 cli/semaphores/open.c -o cli/semaphores/open -lpthread
gcc -Wall -pedantic -std=c99 cli/semaphores/wait.c -o cli/semaphores/wait -lpthread
gcc -Wall -pedantic -std=c99 cli/semaphores/signal.c -o cli/semaphores/signal -lpthread
gcc -Wall -pedantic -std=c99 cli/semaphores/unlink.c -o cli/semaphores/unlink -lpthread
go install ./api/types
go install ./api/context
go install ./api/utils
go install ./api/registry
go install ./api/utils/schema
go install ./api/server/services
go install ./api/server/httputils
go install ./api/server/handlers
go install ./api/utils/paths
go install ./api/utils/config
go install ./api/utils/semaphore
go install ./drivers/storage/isilon
go install ./drivers/storage/isilon/storage
go install ./drivers/storage/scaleio
go install ./drivers/storage/scaleio/storage
go install ./drivers/storage/vbox
go install ./drivers/storage/vbox/storage
go install ./drivers/storage/vfs
go install ./drivers/storage/vfs/storage
go install ./imports/remote
env GOOS=linux GOARCH=amd64 make -j $GOPATH/bin/linux_amd64/lsx-linux
make[2]: Entering directory './github.com/emccode/libstorage'
go install ./api/types
go install ./drivers/storage/isilon
go install ./api/context
go install ./api/utils
go install ./api/registry
go install ./api/utils/config
go install ./drivers/storage/isilon/executor
go install ./drivers/storage/scaleio/executor
go install ./drivers/storage/vbox/executor
go install ./drivers/storage/vfs/executor
go install ./imports/executors
go install ./cli/lsx
go install ./cli/lsx/lsx-linux
make[2]: Leaving directory './github.com/emccode/libstorage'
go install ./imports/config
go install ./drivers/storage/isilon/executor
go install ./drivers/storage/scaleio/executor
go install ./drivers/storage/vbox/executor
go install ./drivers/storage/vfs/executor
go install ./imports/executors
go install ./cli/lsx
go install ./cli/lsx/lsx-darwin
go install github.com/emccode/libstorage/vendor/github.com/jteeuwen/go-bindata/go-bindata
$GOPATH/bin/go-bindata -md5checksum -pkg executors -prefix api/server/executors/bin -o api/server/executors/executors_generated.go api/server/executors/bin/...
go install ./api/server/executors
go install ./api/server/router/executor
go install ./api/server/router/root
go install ./api/server/router/service
go install ./api/utils/filters
go install ./api/server/router/volume
go install ./api/server/router/snapshot
go install ./api/server/router/tasks
go install ./imports/routers
go install ./api/server
go install ./drivers/integration/docker
go install ./drivers/os/darwin
go install ./drivers/os/linux
go install ./api/client
go install ./drivers/storage/libstorage
go install ./drivers/storage/vfs/client
go install ./imports/local
go install ./client
go install .
go install ./api
go install ./api/server/router
go install ./api/tests
go install ./cli/lss
go install ./cli/lss/lss-darwin
go install ./drivers/storage/vbox/client
make -j libstor-c libstor-s
make[2]: Entering directory './github.com/emccode/libstorage'
go build -buildmode=c-shared -o $GOPATH/pkg/darwin_amd64/github.com/emccode/libstorage/c/libstor-c.so ./c/libstor-c
go install ./drivers/storage/isilon/storage
go install github.com/emccode/libstorage/vendor/github.com/jteeuwen/go-bindata/go-bindata
go build -buildmode=c-shared -o $GOPATH/pkg/darwin_amd64/github.com/emccode/libstorage/c/libstor-s.so ./c/libstor-s
gcc -Wall -pedantic -std=c99 -I$GOPATH/src/github.com/emccode/libstorage/c/libstor-c \
          -I$GOPATH/pkg/darwin_amd64/github.com/emccode/libstorage/c/ \
          -L$GOPATH/pkg/darwin_amd64/github.com/emccode/libstorage/c/ \
          -o $GOPATH/bin/libstor-c \
          ./c/libstor-c.c \
          -lstor-c
gcc -Wall -pedantic -std=c99 -I$GOPATH/src/github.com/emccode/libstorage/c \
          -I$GOPATH/pkg/darwin_amd64/github.com/emccode/libstorage/c/ \
          -L$GOPATH/pkg/darwin_amd64/github.com/emccode/libstorage/c/ \
          -o $GOPATH/bin/libstor-s \
          ./c/libstor-s.c \
          -lstor-s
make[2]: Leaving directory './github.com/emccode/libstorage'
make build-lss
make[2]: Entering directory './github.com/emccode/libstorage'
go install ./drivers/storage/isilon/storage
go install github.com/emccode/libstorage/vendor/github.com/jteeuwen/go-bindata/go-bindata
make[2]: Leaving directory './github.com/emccode/libstorage'
make[1]: Leaving directory './github.com/emccode/libstorage'
```

## Version File
There is a file at the root of the project named `VERSION`. The file contains
a single line with the *target* version of the project in the file. The version
follows the format:

  `(?<major>\d+)\.(?<minor>\d+)\.(?<patch>\d+)(-rc\d+)?`

For example, during active development of version `0.1.0` the file would
contain the version `0.1.0`. When it's time to create `0.1.0`'s first
release candidate the version in the file will be changed to `0.1.0-rc1`. And
when it's time to release `0.1.0` the version is changed back to `0.1.0`.

So what's the point of the file if it's basically duplicating the utility of a
tag? Well, the `VERSION` file in fact has two purposes:

  1. First and foremost updating the `VERSION` file with the same value as that
     of the tag used to create a release provides a single, contextual reason to
     push a commit and tag. Otherwise some random commit off of `master` would
     be tagged as a release candidate or release. Always using the commit that
     is related to updating the `VERSION` file is much cleaner.

  2. The contents of the `VERSION` file are also used during the build process
     as a means of overriding the output of a `git describe`. This enables the
     semantic version injected into the produced binary to be created using
     the *targeted* version of the next release and not just the value of the
     last, tagged commit.
