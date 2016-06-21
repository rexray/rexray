# Build Reference

How to build REX-Ray

---

## Build Requirements
This project has very few build requirements, but there are still one or two
items of which to be aware. Also, please note that this are the requirements to
*build* REX-Ray, not run it.

Requirement | Version
------------|--------
Operating System | Linux, OS X
[Go](https://golang.org/) | >=1.6
[GNU Make](https://www.gnu.org/software/make/) | >=3.80

OS X ships with a very old version of GNU Make, and a package manager like
[Homebrew](http://brew.sh/) can be used to install the required version.

## Cross-Compilation
This project's [`Makefile`](https://github.com/emccode/rexray/blob/master/Makefile)
is configured by default to build for a Linux x86_64 target, but
cross-compilation *is* supported. Therefore the build environment should be
configured to support cross-compilation. Please take a minute to read
this [blog post](http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5)
regarding cross-compilation with Go >=1.5.

## Build Binary
Building from source is pretty simple as all steps, including fetching
dependencies (as well as fetching the tool that fetches dependencies), are
configured as part of the included `Makefile`.

### Build Dependencies
The `make deps` command should be executed prior to any other targets in order
to ensure the necessary dependencies are available.

```sh
$ make deps
target: deps
  ...installing glide...SUCCESS!
  ...glide up...[WARN] Only resolving dependencies for the current OS/Arch
[INFO] Downloading dependencies. Please wait...
[INFO] Fetching updates for github.com/Sirupsen/logrus.
[INFO] Fetching updates for golang.org/x/net.
[INFO] Fetching updates for github.com/spf13/cobra.
[INFO] Fetching updates for github.com/akutz/gofig.
[INFO] Fetching updates for github.com/emccode/libstorage.
[INFO] Fetching updates for gopkg.in/yaml.v1.
[INFO] Fetching updates for gopkg.in/yaml.v2.
[INFO] Fetching updates for google.golang.org/api.
[INFO] Fetching updates for github.com/go-yaml/yaml.
[INFO] Fetching updates for github.com/akutz/gotil.
[INFO] Fetching updates for github.com/spf13/pflag.
[INFO] Fetching updates for github.com/akutz/golf.
[INFO] Fetching updates for github.com/akutz/goof.
[INFO] Setting version for github.com/akutz/golf to v0.1.1.
[INFO] Setting version for gopkg.in/yaml.v1 to b4a9f8c4b84c6c4256d669c649837f1441e4b050.
[INFO] Setting version for github.com/go-yaml/yaml to b4a9f8c4b84c6c4256d669c649837f1441e4b050.
[INFO] Setting version for github.com/spf13/cobra to 363816bb13ce1710460c2345017fd35593cbf5ed.
[INFO] Setting version for google.golang.org/api to fd081149e482b10c55262756934088ffe3197ea3.
[INFO] Setting version for github.com/akutz/goof to v0.1.0.
[INFO] Setting version for github.com/akutz/gofig to v0.1.4.
[INFO] Setting version for github.com/spf13/pflag to b084184666e02084b8ccb9b704bf0d79c466eb1d.
[INFO] Setting version for github.com/Sirupsen/logrus to feature/logrus-aware-types.
[INFO] Setting version for gopkg.in/yaml.v2 to b4a9f8c4b84c6c4256d669c649837f1441e4b050.
[INFO] Setting version for github.com/akutz/gotil to v0.1.0.
[INFO] Setting version for github.com/emccode/libstorage to v0.1.3.
[INFO] Resolving imports
[INFO] Fetching github.com/spf13/viper into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Setting version for github.com/spf13/viper to support/rexray.
[INFO] Fetching github.com/kardianos/osext into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Setting version for github.com/kardianos/osext to master.
[INFO] Fetching github.com/gorilla/context into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/gorilla/mux into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/gorilla/handlers into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/cpuguy83/go-md2man/md2man into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/inconshreveable/mousetrap into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/BurntSushi/toml into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/kr/pretty into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/magiconair/properties into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/mitchellh/mapstructure into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/spf13/cast into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/spf13/jwalterweatherman into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching gopkg.in/fsnotify.v1 into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/emccode/goisilon into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Setting version for github.com/emccode/goisilon to f9b53f0aaadb12a26b134830142fc537f492cb13.
[INFO] Fetching github.com/emccode/goscaleio into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Setting version for github.com/emccode/goscaleio to 53ea76f52205380ab52b9c1f4ad89321c286bb95.
[INFO] Fetching github.com/appropriate/go-virtualboxclient/vboxwebsrv into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Setting version for github.com/appropriate/go-virtualboxclient to e0978ab2ed407095400a69d5933958dd260058cd.
[INFO] Fetching github.com/russross/blackfriday into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/kr/text into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching golang.org/x/sys/unix into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/cesanta/ucl into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/cesanta/validate-json/schema into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/shurcooL/sanitized_anchor_name into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/asaskevich/govalidator into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Fetching github.com/jteeuwen/go-bindata into /Users/akutz/Projects/go/src/github.com/emccode/rexray/vendor
[INFO] Setting version for github.com/jteeuwen/go-bindata to feature/md5checksum.
[INFO] Downloading dependencies. Please wait...
[INFO] Setting references for remaining imports
[INFO] Versions did not change. Skipping glide.lock update.
[INFO] Project relies on 38 dependencies.
SUCCESS!
  ...go get...SUCCESS!
cd vendor/github.com/emccode/libstorage && make api/api_generated.go && cd -
echo generating api/api_generated.go
generating api/api_generated.go
/Users/akutz/Projects/go/src/github.com/emccode/rexray
cd vendor/github.com/emccode/libstorage && make api/server/executors/executors_generated.go && cd -
GOOS=darwin GOARCH=amd64 go install ./api/types
GOOS=darwin GOARCH=amd64 go install ./api/context
GOOS=darwin GOARCH=amd64 go install ./api/utils
GOOS=darwin GOARCH=amd64 go install ./api/registry
GOOS=darwin GOARCH=amd64 go install ./api/utils/config
GOOS=darwin GOARCH=amd64 go install ./imports/config
GOOS=darwin GOARCH=amd64 go install ./drivers/storage/isilon
GOOS=darwin GOARCH=amd64 go install ./drivers/storage/isilon/executor
GOOS=darwin GOARCH=amd64 go install ./drivers/storage/scaleio
GOOS=darwin GOARCH=amd64 go install ./drivers/storage/scaleio/executor
GOOS=darwin GOARCH=amd64 go install ./drivers/storage/vbox
GOOS=darwin GOARCH=amd64 go install ./drivers/storage/vbox/executor
GOOS=darwin GOARCH=amd64 go install ./drivers/storage/vfs
GOOS=darwin GOARCH=amd64 go install ./drivers/storage/vfs/executor
GOOS=darwin GOARCH=amd64 go install ./imports/executors
GOOS=darwin GOARCH=amd64 go install ./cli/lsx
GOOS=darwin GOARCH=amd64 go install ./cli/lsx/lsx-darwin
/Users/akutz/Projects/go/bin/go-bindata -md5checksum -pkg executors -prefix api/server/executors/bin -o api/server/executors/executors_generated.go api/server/executors/bin/...
/Users/akutz/Projects/go/src/github.com/emccode/rexray
```

### Basic Build
The most basic build can be achieved by simply typing `make` from the project's
root directory. For what it's worth, executing `make` sans arguments is the
same as executing `make install` for this project's `Makefile`.

```sh
$ make
SemVer: 0.4.0-rc4+10+dirty
RpmVer: 0.4.0+rc4+10+dirty
Branch: release/0.4.0-rc4
Commit: d2f0283c7fed29ad0a142a6a51624828893f5db5
Formed: Wed, 15 Jun 2016 16:53:44 CDT

target: fmt
  ...formatting rexray...SUCCESS!
target: install
  ...installing rexray Darwin-x86_64...SUCCESS!

The REX-Ray binary is 40MB and located at:

  /Users/akutz/Projects/go/bin/rexray
```

### Binary Size
Please note that the large size of the binary is due to the Isilon
storage adapter's dependency on the
[VMware VMOMI library for Go](https://github.com/vmware/govmomi). The types
definition source in that package compiles to a 45MB, uncompressed archive.
Efforts are being made to figure out an alternative to this dependency in order
to reduce the binary size, even if it means creating VMware SOAP messages from
scratch.

### Build All
In order to build all versions of the binary type the following:

```sh
$ make build-all
SemVer: 0.4.0-rc4+10
RpmVer: 0.4.0+rc4+10
Branch: release/0.4.0-rc4
Commit: d2f0283c7fed29ad0a142a6a51624828893f5db5
Formed: Wed, 15 Jun 2016 16:42:33 CDT

target: fmt
  ...formatting rexray...SUCCESS!
target: build
  ...building rexray Linux-x86_64...SUCCESS!

The REX-Ray binary is 40MB and located at:

  .build/bin/Linux-x86_64/rexray
```

## Build Packages
The `Makefile` also includes targets that assist in the creation of
distributable packages using the produced artifact.

### Build Tarballs
The `Makefile`'s `build` and `build-all` targets will not only build the binary
in place, but it will also compress the binary as a tarball so it's ready for
deployment. For example, after the `make build-all` above, this is the contents
of the directory `.build/deploy`:

```sh
$ ls .build/deploy
latest/
Linux-x86_64/
```

Looking inside the directory `.build/deploy/Linux-x86_64` reveals:

```sh
$ ls .build/deploy/Linux-x86_64/
rexray-Linux-x86_64-0.4.0-rc4+10.tar.gz
```

### Build RPMs
The `Makefile`'s `rpm-all` target will package the binary as
architecture-specific RPMs when executed on a system that supports the RPM
development environment:

```sh
$ make rpm-all
target: rpm
  ...building rpm x86_64...SUCCESS!

The REX-Ray RPM is 11MB and located at:

  .build/deploy/Linux-x86_64/rexray-0.4.0+rc4+10-1.x86_64.rpm
```

### Build DEBs
The `Makefile`'s `deb-all` target will package the binary as
architecture-specific DEBs when executed on a system that supports the Alien
RPM-to-DEB conversion tools:

```sh
$ make deb-all
target: deb
  ...building deb x86_64...SUCCESS!

The REX-Ray DEB is 9MB and located at:

  .build/deploy/Linux-x86_64/rexray_0.4.0+rc4+10-1_amd64.deb
```

## Version File
There is a file at the root of the project named `VERSION`. The file contains
a single line with the *target* version of the project in the file. The version
follows the format:

  `(?<major>\d+)\.(?<minor>\d+)\.(?<patch>\d+)(-rc\d+)?`

For example, during active development of version `0.4.0` the file would
contain the version `0.4.0`. When it's time to create `0.4.0`'s first
release candidate the version in the file will be changed to `0.4.0-rc1`. And
when it's time to release `0.4.0` the version is changed back to `0.4.0`.

Please note that we've discussed making the actively developed version the
targeted version with a `-dev` suffix, but trying this resulted in confusion
for the RPM and DEB package managers when using `unstable` releases.

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
