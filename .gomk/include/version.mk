ifneq (1,$(IS_GOMK_VERSION_LOADED))

# note that the file is loaded
IS_GOMK_VERSION_LOADED := 1

VERSION_DEPS := $(SED) $(GREP) $(CAT)

ifeq ($(words $(VERSION_DEPS)),$(words $(wildcard $(VERSION_DEPS))))
ifneq (,$(wildcard .git))

include $(GOMK_I)/arch.mk

# the version's binary os and architecture type
ifeq ($(GOOS),windows)
	V_OS := Windows_NT
endif
ifeq ($(GOOS),linux)
	V_OS := Linux
endif
ifeq ($(GOOS),darwin)
	V_OS := Darwin
endif
ifeq ($(GOARCH),386)
	V_ARCH := i386
endif
ifeq ($(GOARCH),amd64)
	V_ARCH := x86_64
endif
V_OS_ARCH := $(V_OS)-$(V_ARCH)

# parse a semver
SEMVER_PATT := ^[^\d]*(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z].+?))?(?:-(\d+)-g(.+?)(?:-(dirty))?)?$$
PARSE_SEMVER = $(shell echo $(1) | $(SED) -e 's/$(SEMVER_PATT)/$(2)/gim')

# describe the git information and create a parsing function for it
GIT_DESCRIBE := $(shell git describe --long --dirty 2> /dev/null)
ifeq (,$(GIT_DESCRIBE))
	GIT_DESCRIBE := v0.0.1-dev-0-g1234567-dirty
endif
PARSE_GIT_DESCRIBE = $(call PARSE_SEMVER,$(GIT_DESCRIBE),$(1))

# parse the version components from the git information
V_MAJOR := $(call PARSE_GIT_DESCRIBE,$$1)
V_MINOR := $(call PARSE_GIT_DESCRIBE,$$2)
V_PATCH := $(call PARSE_GIT_DESCRIBE,$$3)
V_NOTES := $(call PARSE_GIT_DESCRIBE,$$4)
V_BUILD := $(call PARSE_GIT_DESCRIBE,$$5)
V_SHA_SHORT := $(call PARSE_GIT_DESCRIBE,$$6)
V_DIRTY := $(call PARSE_GIT_DESCRIBE,$$7)

# the long commit hash
V_SHA_LONG := $(shell git show HEAD -s --format=%H)

# the branch name, possibly from travis-ci
ifeq ($(origin TRAVIS_BRANCH), undefined)
	TRAVIS_BRANCH := $(shell git branch | $(GREP) -F '*' | $(SED) 's/\* //')
else
	ifeq ($(strip $(TRAVIS_BRANCH)),)
		TRAVIS_BRANCH := $(shell git branch | $(GREP) -F '*' | $(SED) 's/\* //')
	endif
endif
ifeq ($(origin TRAVIS_TAG), undefined)
	TRAVIS_TAG := $(TRAVIS_BRANCH)
else
	ifeq ($(strip $(TRAVIS_TAG)),)
		TRAVIS_TAG := $(TRAVIS_BRANCH)
	endif
endif
V_BRANCH := $(TRAVIS_TAG)

# the build date as an epoch
V_EPOCH = $(shell $(DATE) +%s)

# the build date
V_BUILD_DATE = $(shell $(DATE) -d $(V_EPOCH) +"%a, %d %b %Y %H:%M:%S %Z")

# the release date as required by bintray
V_RELEASE_DATE = $(shell $(DATE) -d $(V_EPOCH) +"%Y-%m-%d")

# init the semver
V_SEMVER := $(V_MAJOR).$(V_MINOR).$(V_PATCH)
ifneq ($(V_NOTES),)
	V_SEMVER := $(V_SEMVER)-$(V_NOTES)
endif

# get the version file's version
ifneq (,$(wildcard VERSION))
V_FILE = $(strip $(shell $(CAT) VERSION 2> /dev/null))
endif

# append the build number and dirty values to the semver if appropriate
ifneq (,$(V_BUILD))
	ifneq (0,$(V_BUILD))
		# if the version file's version is different than the version parsed
		# from the git describe information then use the version file's
		# version
		ifneq (,$(wildcard VERSION))
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
		endif
		V_SEMVER := $(V_SEMVER)+$(V_BUILD)
	endif
endif
ifeq ($(V_DIRTY),dirty)
	V_SEMVER := $(V_SEMVER)+$(V_DIRTY)
endif

# the rpm version cannot have any dashes
V_RPM_SEMVER := $(subst -,+,$(V_SEMVER))

version:
	@echo SemVer: $(V_SEMVER)
	@echo RpmVer: $(V_RPM_SEMVER)
	@echo Binary: $(V_OS_ARCH)
	@echo Branch: $(V_BRANCH)
	@echo Commit: $(V_SHA_LONG)
	@echo Formed: $(V_BUILD_DATE)
GO_PHONY += version

endif
endif

endif
