# Build Reference

How to build libStorage

---

## Basic Builds
The following one-line command is the quickest, simplest, and most
deterministic approach to building libStorage:

```bash
$ go get github.com/codedellemc/libstorage
```

### Basic Build Requirements
Building libStorage with Docker has the following requirements:

Requirement | Version
------------|--------
Operating System | Linux, OS X
[Go](https://golang.org/) | >=1.6
[Git](https://git-scm.com/) | >= 1.7

## Advanced Builds
While building libStorage can be simple, there are additional options
that facilitate advanced build configurations.

### Advanced Build Requirements
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
[Perl](https://www.perl.org/)  | >= 5.0
[Git](https://git-scm.com/) | >= 1.7

OS X ships with a very old version of GNU Make, and a package manager like
[Homebrew](http://brew.sh/) can be used to install the required version.

It's also possible to use GCC as the Cgo compiler for OS X or to use Clang on
Linux, but by default Clang is used on OS X and GCC on Linux.

### Advanced Build Targets
The following targets are available when building libStorage directly:

| Target | Description |
| --- | --- |
| `build` | Builds libStorage. |
| `test` | Executes all of the libStorage tests. |
| `clean` | This target removes all of the source file markers. |

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
