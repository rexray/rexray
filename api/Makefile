SHELL := $(shell env which bash)

all: build

# define the Golang version as this will fail when importing ../golang.mk
# due to the path structure
GO_VERSION := $(shell grep -A 1 '^go:' ../.travis.yml | tail -n 1 | awk '{print $$2}')

-include ../golang.mk

################################################################################
##                               OS/ARCH INFO                                 ##
################################################################################

ifeq (,$(strip $(OS)))
ifeq ($(GOOS),windows)
OS := Windows_NT
endif
ifeq ($(GOOS),linux)
OS := Linux
endif
ifeq ($(GOOS),darwin)
OS := Darwin
endif
endif

ifeq (,$(strip $(ARCH)))

ifeq ($(GOARCH),386)
ARCH := i386
endif # ifeq ($(GOARCH),386)

ifeq ($(GOARCH),amd64)
ARCH := x86_64
endif # ifeq ($(GOARCH),amd64)

ifeq ($(GOARCH),arm)
ifeq (,$(strip $(GOARM)))
GOARM := 7
endif # ifeq (,$(strip $(GOARM)))
ARCH := ARMv$(GOARM)
endif # ifeq ($(GOARCH),arm)

ifeq ($(GOARCH),arm64)
ARCH := ARMv8
endif # ifeq ($(GOARCH),arm64)

endif # ifeq (,$(ARCH))


# if GOARCH=arm & GOARM="" then figure out what
# the correct GOARM version is and export it
ifeq (arm,$(GOARCH))
ifeq (,$(strip $(GOARM)))
ifeq (ARMv5,$(ARCH))
GOARM := 5
endif # ifeq (ARMv5,$(ARCH))
ifeq (ARMv6,$(ARCH))
GOARM := 6
endif # ifeq (ARMv6,$(ARCH))
ifeq (ARMv7,$(ARCH))
GOARM := 7
endif # ifeq (ARMv7,$(ARCH))
endif # ifeq (,$(strip $(GOARM)))
export GOARM
endif # ifeq (arm,$(GOARCH))


# if GOARCH is arm64 then undefine & unexport GOARM
ifeq (arm64,$(GOARCH))
ifneq (undefined,$(origin GOARM))
undefine GOARM
unexport GOARM
endif
endif # ifeq ($(GOARCH),arm64)


# ensure that GOARM is compatible with the GOOS &
# GOARCH per https://github.com/golang/go/wiki/GoArm
# when GOARCH=arm
ifeq (arm,$(GOARCH))
ifeq (darwin,$(GOOS))
GOARM_ALLOWED := 7
else
GOARM_ALLOWED := 5 6 7
endif # ifeq (darwin,$(GOOS))
ifeq (,$(strip $(filter $(GOARM),$(GOARM_ALLOWED))))
$(info incompatible GOARM version: $(GOARM))
$(info allowed GOARM versions are: $(GOARM_ALLOWED))
$(info plese see https://github.com/golang/go/wiki/GoArm)
exit 1
endif # ifeq (,$(strip $(filter $(GOARM),$(GOARM_ALLOWED))))
endif # ifeq (arm,$(GOARCH))

export OS
export ARCH

# get the project's root import path as well as whether or not
# this project is vendored inside of another
ROOT_IMPORT_PATH := $(patsubst %/api,%,$(shell go list))
ROOT_IMPORT_PATH_NV := $(ROOT_IMPORT_PATH)
VENDORED := 0
ifneq (,$(strip $(findstring vendor,$(ROOT_IMPORT_PATH))))
VENDORED := 1
ROOT_IMPORT_PATH_NV := $(shell echo $(ROOT_IMPORT_PATH) | sed 's/.*vendor\///g')
endif

# figure out the git dirs
GIT_WORK:=../
GIT_ROOT:=../.git
ifeq (1,$(VENDORED))
ifneq (,$(wildcard $(HOME)/.glide))
ROOT_IMPORT_PATH_DASH:=$(subst /,-,$(ROOT_IMPORT_PATH_NV))
VGIT_WORK:=$(shell find $(HOME)/.glide -name "*$(ROOT_IMPORT_PATH_DASH)" -type d)
ifneq (,$(wildcard $(VGIT_WORK)))
GIT_WORK:=$(VGIT_WORK)
ifneq (,$(wildcard $(VGIT_WORK)/.git))
GIT_ROOT:=$(VGIT_WORK)/.git
endif
endif
endif
endif

# parse a semver
SEMVER_PATT := ^[^\d]*(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z].+?))?(?:-(\d+)-g(.+?)(?:-(dirty))?)?$$
PARSE_SEMVER = $(shell echo $(1) | perl -pe 's/$(SEMVER_PATT)/$(2)/gim')

# describe the git information and create a parsing function for it
GIT_DESCRIBE := $(shell git --git-dir="$(GIT_ROOT)" --work-tree="$(GIT_WORK)" describe --long --dirty)
PARSE_GIT_DESCRIBE = $(call PARSE_SEMVER,$(GIT_DESCRIBE),$(1))

# parse the version components from the git information
V_MAJOR := $(call PARSE_GIT_DESCRIBE,$$1)
V_MINOR := $(call PARSE_GIT_DESCRIBE,$$2)
V_PATCH := $(call PARSE_GIT_DESCRIBE,$$3)
V_NOTES := $(call PARSE_GIT_DESCRIBE,$$4)
V_BUILD := $(call PARSE_GIT_DESCRIBE,$$5)
V_SHA_SHORT := $(call PARSE_GIT_DESCRIBE,$$6)
V_DIRTY := $(call PARSE_GIT_DESCRIBE,$$7)

V_OS := $(OS)
V_ARCH := $(ARCH)
V_OS_ARCH := $(V_OS)-$(V_ARCH)

# the long commit hash
V_SHA_LONG := $(shell git --git-dir="$(GIT_ROOT)" --work-tree="$(GIT_WORK)" show HEAD -s --format=%H)

# the branch name, possibly from travis-ci
ifeq ($(origin TRAVIS_BRANCH), undefined)
	TRAVIS_BRANCH := $(shell git --git-dir="$(GIT_ROOT)" --work-tree="$(GIT_WORK)" branch | grep '*')
else
ifeq (,$(strip $(TRAVIS_BRANCH)))
	TRAVIS_BRANCH := $(shell git --git-dir="$(GIT_ROOT)" --work-tree="$(GIT_WORK)" branch | grep '*')
endif
endif

ASTERIK := *
LPAREN := (
RPAREN := )

TRAVIS_BRANCH := $(subst $(ASTERIK) ,,$(TRAVIS_BRANCH))
TRAVIS_BRANCH := $(subst $(LPAREN)HEAD detached at ,,$(TRAVIS_BRANCH))
TRAVIS_BRANCH := $(subst $(LPAREN)detached at ,,$(TRAVIS_BRANCH))
TRAVIS_BRANCH := $(subst $(LPAREN)HEAD detached from ,,$(TRAVIS_BRANCH))
TRAVIS_BRANCH := $(subst $(LPAREN)detached from ,,$(TRAVIS_BRANCH))
TRAVIS_BRANCH := $(subst $(RPAREN),,$(TRAVIS_BRANCH))

ifeq ($(origin TRAVIS_TAG), undefined)
	TRAVIS_TAG := $(TRAVIS_BRANCH)
else
	ifeq ($(strip $(TRAVIS_TAG)),)
		TRAVIS_TAG := $(TRAVIS_BRANCH)
	endif
endif
V_BRANCH := $(TRAVIS_TAG)

# the build date as an epoch
V_EPOCH := $(shell date +%s)

# the build date
V_BUILD_DATE := $(shell perl -e 'use POSIX strftime; print strftime("%a, %d %b %Y %H:%M:%S %Z", localtime($(V_EPOCH)))')

# the release date as required by bintray
V_RELEASE_DATE := $(shell perl -e 'use POSIX strftime; print strftime("%Y-%m-%d", localtime($(V_EPOCH)))')

# init the semver
V_SEMVER := $(V_MAJOR).$(V_MINOR).$(V_PATCH)
ifneq ($(V_NOTES),)
	V_SEMVER := $(V_SEMVER)-$(V_NOTES)
endif

# get the version file's version
V_FILE := $(strip $(shell cat ../VERSION 2> /dev/null))

# append the build number and dirty values to the semver if appropriate
ifneq ($(V_BUILD),)
	ifneq ($(V_BUILD),0)
		# if the version file's version is different than the version parsed from the
		# git describe information then use the version file's version
		ifneq ($(V_SEMVER),$(V_FILE))
			V_MAJOR := $(call PARSE_SEMVER,$(V_FILE),$$1)
			V_MINOR := $(call PARSE_SEMVER,$(V_FILE),$$2)
			V_PATCH := $(call PARSE_SEMVER,$(V_FILE),$$3)
			V_NOTES := $(call PARSE_SEMVER,$(V_FILE),$$4)
			V_SEMVER := $(V_MAJOR).$(V_MINOR).$(V_PATCH)
			ifneq ($(V_NOTES),)
				V_SEMVER := $(V_SEMVER)-$(V_NOTES)
			endif
		endif
		V_SEMVER := $(V_SEMVER)+$(V_BUILD)
	endif
endif
ifeq ($(V_DIRTY),dirty)
	V_SEMVER := $(V_SEMVER)+$(V_DIRTY)
endif

define API_GENERATED_CONTENT
package api

import (
	"time"

	"github.com/codedellemc/libstorage/api/types"
)

func init() {
	Version = &types.VersionInfo{}
	Version.Arch = "$(V_OS_ARCH)"
	Version.Branch = "$(V_BRANCH)"
	Version.BuildTimestamp = time.Unix($(V_EPOCH), 0)
	Version.SemVer = "$(V_SEMVER)"
	Version.ShaLong = "$(V_SHA_LONG)"
}
endef
export API_GENERATED_CONTENT

PRINTF_VERSION_CMD += @printf "SemVer: %s\nBinary: %s\nBranch: %s\nCommit:
PRINTF_VERSION_CMD += %s\nFormed: %s\n" "$(V_SEMVER)" "$(V_OS_ARCH)"
PRINTF_VERSION_CMD += "$(V_BRANCH)" "$(V_SHA_LONG)" "$(V_BUILD_DATE)"
API_GENERATED_SRC := api_version_generated.go
$(API_GENERATED_SRC):
	echo generating $@
	@echo "$$API_GENERATED_CONTENT" > $@

build: $(API_GENERATED_SRC)

version:
	$(PRINTF_VERSION_CMD)

clean:
	rm -f $(API_GENERATED_SRC)

.PHONY: version clean
