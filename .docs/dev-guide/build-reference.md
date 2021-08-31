# Build Reference

How to build REX-Ray

---

## Build with Go
The following one-line command is the quickest, simplest, and most
deterministic approach to building REX-Ray:

```bash
$ go get github.com/AVENTER-UG/rexray
```

The above command will download the REX-Ray sources and build the
binary at `$GOPATH/bin/rexray`.

### SemVer
Using `go get` to download and install REX-Ray is simple, but it also
produces a binary without the correct semantic version. To create
the version information use `go generate`:

```bash
$ go generate github.com/AVENTER-UG/rexray
```

To download and build REX-Ray in one line with the correct version
information please use the following command:

```bash
$ go get -d github.com/AVENTER-UG/rexray && \
  go generate github.com/AVENTER-UG/rexray && \
  go install github.com/AVENTER-UG/rexray
```

### Build Tags
A REX-Ray build can be influenced through the use of Go build tags. For
example, the following command builds REX-Ray with only the EBS driver:

```
go build -tags ebs -o rexray
```

The table below includes the tags that can be used to determine what
type of REX-Ray binary is produced.

| Tag | Description |
|-----|-------------|
| `agent` | Builds an agent-only REX-Ray binary |
| `client` | Builds a client-only REX-Ray binary |
| `controller` | Builds a controller-only REX-Ray binary |

Additionally, each of the directories in `./libstorage/storage/drivers`
can be used as a build tag to produce a REX-Ray binary with specific
drivers.

If no build tags are provided then REX-Ray is built with all drivers
and the binary includes the client, agent, and controller modes.

### Go Build Requirements
Building REX-Ray with Go has the following requirements:

Requirement | Version
------------|--------
Operating System | Linux, OS X
[Go](https://golang.org/) | >=1.6
[Git](https://git-scm.com/) | >= 1.7

## Build with Docker
Docker can be used to build REX-Ray binaries that have checksums which
match the same binaries produced by REX-Ray's official builds on
Travis-CI. Simply clone the REX-Ray repository (or fork) and checkout
the desired reference. Then use the following command to build REX-Ray:

```bash
$ SRC=github.com/AVENTER-UG/rexray && \
  docker run -it \
  -e SRC -e GOOS -e GOARCH \
  -v "$(pwd)":/go/src/$SRC golang:1.8.3 \
  bash -c "cd src/$SRC && \
  XGOOS=$GOOS XGOARCH=$GOARCH GOOS= GOARCH= go generate && \
  go build -o rexray"
```

Additionally, if Docker is detected and running on the local host the
following command will also use Docker to build REX-Ray:

```bash
$ make
```

### Docker Build Requirements
This project has very few build requirements, but there are still one or two
items of which to be aware. Also, please note these are the requirements to
*build* REX-Ray, not run it.

Requirement | Version
------------|--------
Operating System | Linux, OS X
[Docker](https://www.docker.com/) | >=17.05

## Docker Plug-ins
This section describes how to build and create Docker plug-ins
and as such requires Docker.

### Creating Docker Plug-ins
The first step to creating a new Docker plug-in is creating its
directory and skeleton files. With `$DRIVER` being the storage platform
for which the plug-in is being created, execute the following command:

```bash
$ mkdir -p .docker/plugins/$DRIVER
```

The next step is to create the `README.md` and `config.json` files
that will be placed in the new directory. Please use this
[gist](https://gist.github.com/akutz/0212d43ddf502aa52ccfffc866320e7f)
as a starting point for creating those files.

### Building Docker Plug-ins
To build a Docker plug-in it is first necessary to create the REX-Ray
binary that the plug-in will use:

```bash
$ DRIVER=$DRIVER make
```

The above command will create a binary named `rexray` that embeds
only the driver specified as the environment variable `DRIVER`.
The next step is to create the Docker plug-in itself:

```bash
$ DRIVER=$DRIVER make build-docker-plugin
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
