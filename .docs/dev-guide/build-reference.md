# Build Reference

How to build REX-Ray

---

## Build Requirements
This project has very few build requirements, but there are still one or two
items of which to be aware. Also, please note that this are the requirements to
*build* `REX-Ray`, not run it.

Requirement | Version
------------|--------
Operating System | Linux, OS X
[Go](https://golang.org/) | >=1.5
[GNU Make](https://www.gnu.org/software/make/) | >=3.80

OS X ships with a very old version of GNU Make, and a package manager like
[Homebrew](http://brew.sh/) can be used to install the required version.

## Cross-Compilation
This project's [`Makefile`](https://github.com/emccode/rexray/blob/master/Makefile)
is configured by default to cross-compile for Linux x86 & x86_64 as well as
Darwin (OS X) x86_64. Therefore the build process will fail if the local Go
environment is not set up for cross-compilation. Please take a minute to read
this [blog post](http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5)
regarding cross-compilation with Go >=1.5.

## Build Binary
Building from source is pretty simple as all steps, including fetching
dependencies (as well as fetching the tool that fetches dependencies), are
configured as part of the included `Makefile`.

### Basic Build
The most basic build can be achieved by simply typing `make` from the project's
root directory. For what it's worth, executing `make` sans arguments is the
same as executing `make install` for this project's `Makefile`.

```sh
$ make
SemVer: 0.3.1-rc1+8+dirty
RpmVer: 0.3.1+rc1+8+dirty
Branch: feature/dev-guide
Commit: ea59957557d01725ec147d5f71ce6a30aa8698e9
Formed: Wed, 16 Dec 2015 14:58:33 CST

target: deps
  ...installing glide...SUCCESS!
  ...downloading go dependencies...SUCCESS!
target: fmt
  ...formatting rexray...SUCCESS!
target: install
  ...installing rexray Darwin-x86_64...SUCCESS!

The REX-Ray binary is 62MB and located at:

  /Users/akutz/Projects/go/bin/rexray
```

### Binary Size
Please note that the extraordinary size of the binary is due to the Isilon
storage adapter's dependency on the
[VMware VMOMI library for Go](https://github.com/vmware/govmomi). The types
definition source in that package compiles to a 45MB archive. We're currently
working to figure out an alternative to this, even if it means creating VMware
SOAP messages from scratch.

### Build All
In order to build all versions of the binary type the following:

```sh
$ make build-all
SemVer: 0.3.1-rc1+8+dirty
RpmVer: 0.3.1+rc1+8+dirty
Branch: feature/dev-guide
Commit: ea59957557d01725ec147d5f71ce6a30aa8698e9
Formed: Wed, 16 Dec 2015 15:03:27 CST

target: deps
  ...installing glide...SUCCESS!
  ...downloading go dependencies...SUCCESS!
target: fmt
  ...formatting rexray...SUCCESS!
target: build
  ...building rexray Linux-i386...SUCCESS!

The REX-Ray binary is 53MB and located at:

  .build/bin/Linux-i386/rexray

target: build
  ...building rexray Linux-x86_64...SUCCESS!

The REX-Ray binary is 62MB and located at:

  .build/bin/Linux-x86_64/rexray

target: build
  ...building rexray Darwin-x86_64...SUCCESS!

The REX-Ray binary is 62MB and located at:

  .build/bin/Darwin-x86_64/rexray
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
Darwin-x86_64/
Linux-i386/
Linux-x86_64/
latest/
```

Looking inside the directory `.build/deploy/Linux-x86_64` reveals:

```sh
$ ls .build/deploy/Linux-x86_64/
rexray-Linux-x86_64-0.3.1-rc1+8+dirty.tar.gz
```

### Build RPMs
The `Makefile`'s `rpm-all` target will package the binary as
architecture-specific RPMs when executed on a system that supports the RPM
development environment:

```sh
$ make rpm-all
target: rpm
  ...building rpm i386...SUCCESS!

The REX-Ray RPM is 6MB and located at:

  .build/deploy/Linux-i386/rexray-0.3.1+rc1+8+dirty-1.i386.rpm

target: rpm
  ...building rpm x86_64...SUCCESS!

The REX-Ray RPM is 7MB and located at:

  .build/deploy/Linux-x86_64/rexray-0.3.1+rc1+8+dirty-1.x86_64.rpm
```

### Build DEBs
The `Makefile`'s `deb-all` target will package the binary as
architecture-specific DEBs when executed on a system that supports the Alien
RPM-to-DEB conversion tools:

```sh
$ make deb-all
target: deb
  ...building deb x86_64...SUCCESS!

The REX-Ray DEB is 4MB and located at:

  .build/deploy/Linux-x86_64/rexray_0.3.1+rc1+8+dirty-1_amd64.deb
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
