ifneq (1,$(IS_GO_MK_CONFIG_LOADED))

# note that the file is loaded
IS_GO_MK_CONFIG_LOADED := 1

EMPTY :=
SPACE := $(EMPTY) $(EMPTY)

# set the make flags
RECURSIVE := 0
ifneq (,$(filter %-all,$(MAKECMDGOALS)))
	RECURSIVE := 1
else
ifneq (,$(word 2,$(filter %_amd64,$(MAKECMDGOALS))))
	RECURSIVE := 1
else
ifneq (,$(word 2,$(filter %_386,$(MAKECMDGOALS))))
	RECURSIVE := 1
endif
endif
endif

MAKEFLAGS := --no-print-directory
ifeq (1,$(RECURSIVE))
MAKEFLAGS += --output-sync=recurse
endif

ifeq (,$(CURDIR))
CURDIR := $(shell pwd)
endif

# the build platforms for which to perform the cross-* goals
BUILD_PLATFORMS ?= Linux-x86_64 Darwin-x86_64

# an ordered, space-delimited list of the go package directories to build and
# install. the root package should be a '.', and all the remaining packages
# should be in the form of './subdir1', './subdir1/subdir1a', './subdir2'
#
# if this variable is empty then it will be populated automatically by using
# the find command to find all directories not beginning with a leading '.' or
# '_' and named 'vendor' or 'doc'.
GO_PKG_DIRS ?=

# an ordered, space-delimited list of directory patterns that should be ignored.
# for more information on the pattern matching scheme used, search for help on
# the makefile text functions, specifically the filter-out function.
GO_PKG_DIRS_IGNORE_PATTS ?= ./vendor% ./docs% \
							./daemon/module/admin/html/% \
							./daemon/module/docker/
#							./rexray%

# the suffix to append to the related make targets for a package that has the
# same name as its parent package
GO_DUPLICATE_PKG_SUFFIX ?= -cli

# flags indicating whether or not the following tools are executed against
# the project's source files. valid values are 1 (enabled) and 0 (disabled).
GO_FMT_ENABLED ?= 1
GO_LINT_ENABLED ?= 0
GO_CYCLO_ENABLED ?= 0
GO_VET_ENABLED ?= 0

# flags indicating whether or not the following tools are used to package
# artifacts after a successful build. valid values are 1 (enabled) and 0
# (diabled).
PKG_TGZ_ENABLED ?= 1
PKG_TGZ_EXTENSION ?= .tar.gz

# flag indicating whether or not test coverage results are submitted to
# coveralls. valid values are 1 (enabled) and 0 (disabled). please note that
# even if coveralls is enabled, it will be disabled if the build is not
# running on the travis-ci build system. coveralls is also disabled if no
# tests are detected or all tests are excluded
COVERALLS_ENABLED ?= 1

# a space-delimited list of coverage profile files to exclude when submitting
# results to coveralls
COVERALLS_EXCLUDE ?= .gomk/tmp/tests/rexray/cli/cli.test.out

# a flag indicating whether or not to use glide for dependency management.
# if a glide.yaml file is detected glide is automatically used unless this
# flag indicates it should be disabled. valid values are 1 (enabled) and 0
# (disabled)
GLIDE_ENABLED ?= 1

# the version of glide to install. glide releases can be found at
# https://github.com/Masterminds/glide/releases. this variable is only used
# if glide is used, which is determined by the presence of the file glide.yaml.
GLIDE_VERSION ?= 0.8.3

# the indent to print prior to echoing commands
INDENT_LEN ?=

# flags to use with go build
ifeq ($(origin GO_BUILD_FLAGS),undefined)
GO_BUILD_FLAGS ?=
endif

# flags to use with go install
ifeq ($(origin GO_INSTALL_FLAGS),undefined)
GO_INSTALL_FLAGS ?=
endif

# flags to use with go clean
ifeq ($(origin GO_CLEAN_FLAGS),undefined)
GO_CLEAN_FLAGS ?= -i
endif

# flags to use when executing tests
ifeq ($(origin GO_TEST_FLAGS),undefined)
GO_TEST_FLAGS ?= -test.v
endif

# a flag indicating whether or not to execute go get during a dependency goal.
# valid values are 1 (enabled) and 0 (disabled)
GO_GET_ENABLED ?= 1

endif
