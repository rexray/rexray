# Build Reference

How to build REX-Ray

---

## Basic Builds
The following one-line command is the quickest, simplest, and most
deterministic approach to building REX-Ray:

```bash
$ git clone https://github.com/codedellemc/rexray && make -C rexray
```

!!! note "note"

    The above `make` command defaults to the `docker-build` target only if
    Docker is detected and running on a host, otherwise the `build` target is
    used. For more information about the `build` target, please see the
    [Advanced Builds](#advanced-builds) section.

### Basic Build Requirements
Building REX-Ray with Docker has the following requirements:

Requirement | Version
------------|--------
Operating System | Linux, OS X
[Docker](https://www.docker.com/) | >=1.11
[GNU Make](https://www.gnu.org/software/make/) | >=3.80
[Git](https://git-scm.com/) | >= 1.7

OS X ships with a very old version of GNU Make, and a package manager like
[Homebrew](http://brew.sh/) can be used to install the required version.

### Basic Build Targets
The following targets are available when building REX-Ray with Docker:

| Target | Description |
| --- | --- |
| `docker-build` | Builds REX-Ray inside a Docker container. |
| `docker-test` | Executes all of the REX-Ray tests inside the container. |
| `docker-clean` | This target stops and removes the default container used for REX-Ray builds. The name of the default container is `build-rexray`. |
| `docker-clobber` | This target stops and removes all Docker containers that have a name that matches the name of the configured container prefix (default prefix is `build-rexray`). |
| `docker-list` | Lists all Docker containers that have a name that matches the name of the configured prefix (default prefix is `build-rexray`). |

### Basic Build Options
The following options (via environment variables) can be used to influence
how REX-Ray is built with Docker:

| Environment Variable | Description |
| --- | --- |
| `DRIVERS` | This variable can be set to a space-delimited list of driver names in order to indicate which storage platforms to support. For example, the command `$ DRIVERS="ebs scaleio" make docker-build` would build REX-Ray for only the EBS and ScaleIO storage platforms.
| `DBUILD_ONCE` | When set to `1`, this environment variable instructs the Makefile to create a temporary, one-time use container for the subsequent build. The container is removed upon a successful build. If the build fails the container is not removed. This is because Makefile error logic is lacking. However, `make docker-clobber` can be used to easily clean up these containers. The containers will follow a given pattern using the container prefix (`build-rexray` is the default prefix value). The one-time containers use `PREFIX-EPOCH`. For example, `build-rexray-1474691232`. |
| `DGOOS` | This sets the OS target for which to build the REX-Ray binaries. Valid values are `linux` and `darwin`. If omitted the host OS value returned from `uname -s` is used instead. |
| `DLOCAL_IMPORTS` | Specify a list of space-delimited import paths that will be copied from the host OS's `GOPATH` into the container build's vendor area, overriding the dependency code that would normally be fetched by Glide.<br/><br/>For example, the project's `glide.yaml` file might specify to build REX-Ray with libStorage v0.2.1. However, the following command will build REX-Ray using the libStorage sources on the host OS at `$GOPATH/src/github.com/codedellemc/libstorage`:<br/><br/><pre lang="bash">$ DLOCAL_IMPORTS=github.com/codedellemc/libstorage make docker-build</pre>Using local sources can sometimes present a problem due to missing dependencies. Please see the next environment variable for instructions on how to overcome this issue. |
| `DGLIDE_YAML` | Specify a file that will be used for the container build in place of the standard `glide.yaml` file.<br/><br/>This is necessary for occasions when sources injected into the build via the `DLOCAL_IMPORTS` variable import packages that are not imported by the package specified in the project's standard `glide.yaml` file.<br/><br/>For example, if `glide.yaml` specifies that REX-Ray depends upon AWS SDK v1.2.2, but `DLOCAL_IMPORTS` specifies the value `github.com/aws/aws-sdk-go` and the AWS SDK source code on the host includes a new dependency not present in the v1.2.2 version, Glide will not fetch the new dependency when doing the container build.<br/><br/>So it may be necessary to use `DGLIDE_YAML` to provide a superset of the project's standard `glide.yaml` file which also includes the dependencies necessary to build the packages specified in `DLOCAL_IMPORTS`. |

## Advanced Builds
While building REX-Ray with Docker is simple, it ultimately relies on the
same `Makefile` included in the REX-Ray repository and so it's entirely
possible (and often desirable) to build REX-Ray directly.

### Advanced Build Requirements
This project has very few build requirements, but there are still one or two
items of which to be aware. Also, please note that this are the requirements to
*build* REX-Ray, not run it.

Requirement | Version
------------|--------
Operating System | Linux, OS X
[Go](https://golang.org/) | >=1.6
[GNU Make](https://www.gnu.org/software/make/) | >=3.80
[Glide](https://glide.sh/) | >=0.10
[X-Code Command Line Tools (OS X only)](https://developer.apple.com/library/ios/technotes/tn2339/_index.html) | >= OS X 10.9
Linux Kernel Headers (Linux only) | >=Linux Kernel 3.13
[GNU C Compiler](https://gcc.gnu.org/) (Linux only) | >= 4.8
[Perl](https://www.perl.org/)  | >= 5.0
[Git](https://git-scm.com/) | >= 1.7

OS X ships with a very old version of GNU Make, and a package manager like
[Homebrew](http://brew.sh/) can be used to install the required version.

It's also possible to use GCC as the Cgo compiler for OS X or to use Clang on
Linux, but by default Clang is used on OS X and GCC on Linux.

### Advanced Build Targets
The following targets are available when building REX-Ray directly:

| Target | Description |
| --- | --- |
| `build` | Builds REX-Ray. |
| `test` | Executes all of the REX-Ray tests. |
| `clean` | This target removes all of the source file markers. |
| `clobber` | This is the same as `clean` but also removes any produced artifacts. |

### Advanced Build Options
The following options (via environment variables) can be used to influence
how REX-Ray is built:

| Environment Variable | Description |
| --- | --- |
| `DRIVERS` | This variable can be set to a space-delimited list of driver names in order to indicate which storage platforms to support. For example, the command `$ DRIVERS="ebs scaleio" make build` would build REX-Ray for only the EBS and ScaleIO storage platforms.

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
