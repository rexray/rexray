# configure make
export MAKEFLAGS := $(MAKEFLAGS) --no-print-directory -k

# store the current working directory
CWD := $(shell pwd)

# enable go 1.5 vendoring
export GO15VENDOREXPERIMENT := 1

# set the go os and architecture types as well the sed command to use based on 
# the os and architecture types
ifeq ($(OS),Windows_NT)
	GOOS ?= windows
	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		export GOARCH ?= amd64
	endif
	ifeq ($(PROCESSOR_ARCHITECTURE),x86)
		export GOARCH ?= 386
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		export GOOS ?= linux
	endif
	ifeq ($(UNAME_S),Darwin)
		export GOOS ?= darwin
		export GOARCH ?= amd64
	endif
	ifeq ($(origin GOARCH), undefined)
		UNAME_P := $(shell uname -p)
		ifeq ($(UNAME_P),x86_64)
			export GOARCH = amd64
		endif
		ifneq ($(filter %86,$(UNAME_P)),)
			export GOARCH = 386
		endif
	endif
endif

# init the internal go os and architecture variable values used for naming files
_GOOS ?= $(GOOS)
_GOARCH ?= $(GOARCH)

# describe the git information and create a parsing function for it
GIT_DESCRIBE := $(shell git describe --long --dirty)
GIT_DESCRIBE_PATT := ^[^\d]*(\d+)\.(\d+)\.(\d+)(?:-(.+?))?(?:-(\d+)-g(.+?)(?:-(dirty))?)?$$
PARSE_GIT_DESCRIBE = $(shell echo $(GIT_DESCRIBE) | perl -pe 's/$(GIT_DESCRIBE_PATT)/$(1)/gim')

# parse the version components from the git information
V_MAJOR := $(call PARSE_GIT_DESCRIBE,$$1)
V_MINOR := $(call PARSE_GIT_DESCRIBE,$$2)
V_PATCH := $(call PARSE_GIT_DESCRIBE,$$3)
V_NOTES := $(call PARSE_GIT_DESCRIBE,$$4)
V_BUILD := $(call PARSE_GIT_DESCRIBE,$$5)
V_SHA_SHORT := $(call PARSE_GIT_DESCRIBE,$$6)
V_DIRTY := $(call PARSE_GIT_DESCRIBE,$$7)

# the version's binary os and architecture type
V_ARCH := $(_GOOS)_$(_GOARCH)

# the long commit hash
V_SHA_LONG := $(shell git show HEAD -s --format=%H)

# the branch name, possibly from travis-ci
ifeq ($(origin TRAVIS_BRANCH), undefined)
	TRAVIS_BRANCH := $(shell git branch | grep '*' | awk '{print $$2}')
else
	ifeq ($(strip $(TRAVIS_BRANCH)),)
		TRAVIS_BRANCH := $(shell git branch | grep '*' | awk '{print $$2}')
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
V_FILE := $(strip $(shell cat VERSION 2> /dev/null))

# if the version file's version is different than the version parsed from the
# git describe information then use the version file's version
ifneq ($(V_SEMVER),$(V_FILE))
	V_SEMVER := $(V_FILE)
endif

# append the build number and dirty values to the semver if appropriate
ifneq ($(V_BUILD),0)
	V_SEMVER := $(V_SEMVER)+$(V_BUILD)
endif
ifeq ($(V_DIRTY),dirty)
	V_SEMVER := $(V_SEMVER)+$(V_DIRTY)
endif

GOFLAGS := $(GOFLAGS)
GLIDE := $(GOPATH)/bin/glide
NV := $$($(GLIDE) novendor)
BASEPKG := github.com/emccode/rexray
BASEDIR := $(GOPATH)/src/$(BASEPKG)
BASEDIR_NAME := $(shell basename $(BASEDIR))
BASEDIR_PARENTDIR := $(shell dirname $(BASEDIR))
BASEDIR_TEMPMVLOC := $(BASEDIR_PARENTDIR)/.$(BASEDIR_NAME)-$(shell date +%s)
VERSIONPKG := $(BASEPKG)/version_info
LDF_SEMVER := -X $(VERSIONPKG).SemVer=$(V_SEMVER)
LDF_BRANCH := -X $(VERSIONPKG).Branch=$(V_BRANCH)
LDF_EPOCH := -X $(VERSIONPKG).Epoch=$(V_EPOCH)
LDF_SHA_LONG := -X $(VERSIONPKG).ShaLong=$(V_SHA_LONG)
LDF_ARCH = -X $(VERSIONPKG).Arch=$(V_ARCH)
LDFLAGS = -ldflags "$(LDF_SEMVER) $(LDF_BRANCH) $(LDF_EPOCH) $(LDF_SHA_LONG) $(LDF_ARCH)"
RPMBUILD := $(CWD)/.rpmbuild
EMCCODE := $(GOPATH)/src/github.com/emccode
PRINT_STATUS = export EC=$$?; cd $(CWD); if [ "$$EC" -eq "0" ]; then printf "SUCCESS!\n"; else exit $$EC; fi
STAT_FILE_SIZE = stat --format '%s' $$FILE 2> /dev/null || stat -f '%z' $$FILE 2> /dev/null
	
all: install

_pre-make:
	@if [ "$(CWD)" != "$(BASEDIR)" ]; then \
		if [ -e "$(BASEDIR)" ]; then \
			mv $(BASEDIR) $(BASEDIR_TEMPMVLOC); \
		fi; \
		mkdir -p "$(BASEDIR_PARENTDIR)"; \
		ln -s "$(CWD)" "$(BASEDIR)"; \
	fi

_post-make:
	@if [ -e "$(BASEDIR_TEMPMVLOC)" -a -L $(BASEDIR) ]; then \
		rm -f $(BASEDIR); \
		mv $(BASEDIR_TEMPMVLOC) $(BASEDIR); \
	fi

deps: _pre-make _deps _post-make
_deps: 
	@echo "target: deps"
	@printf "  ...installing glide..."
	@go get github.com/Masterminds/glide; \
		$(PRINT_STATUS)
	@printf "  ...downloading go dependencies..."; \
		cd $(BASEDIR); \
		go get -d $(GOFLAGS) $(NV); \
		$(GLIDE) -q up 2> /dev/null; \
		$(PRINT_STATUS)

build: _pre-make _build _post-make
_build: _deps _fmt build_
build_: 
	@echo "target: build"
	@printf "  ...building rexray $(V_ARCH)..."; \
		cd $(BASEDIR); \
		FILE=.bin/$(V_ARCH)/rexray; \
		env GOOS=$(_GOOS) GOARCH=$(_GOARCH) go build -o $$FILE $(GOFLAGS) $(LDFLAGS) ./rexray; \
		$(PRINT_STATUS); \
		if [ "$$EC" -eq "0" ]; then \
			BYTES=$$($(STAT_FILE_SIZE)); \
			SIZE=$$(($$BYTES / 1024 / 1024)); \
			printf "\nThe REX-Ray binary is $${SIZE}MB and located at:\n\n"; \
			printf "  $$FILE\n\n"; \
		fi

build-all: _pre-make version-noarch _deps _fmt build-all_ _post-make
build-all_: build-linux-386_ build-linux-amd64_ build-darwin-amd64_
	@for BIN in $$(find .bin -type f -name "rexray"); do \
		BINDIR=$$(dirname $$BIN); \
		FARCH=$$(echo $$BINDIR | cut -c6-); \
		TARBALL=rexray-$$FARCH-$(V_SEMVER).tar.gz; \
		cd $$BINDIR; \
		tar -czf $$TARBALL rexray; \
		cd - > /dev/null; \
	done; \
	sed -e 's/$${SEMVER}/$(V_SEMVER)/g' \
		-e 's|$${DSCRIP}|$(V_SEMVER).Branch.$(V_BRANCH).Sha.$(V_SHA_LONG)|g' \
		-e 's/$${RELDTE}/$(V_RELEASE_DATE)/g' \
		.bintray.json > .bintray-filtered.json

build-linux-386: _pre-make _build-linux-386 _post-make
_build-linux-386: _deps _fmt build-linux-386_
build-linux-386_:
	@env _GOOS=linux _GOARCH=386 make build_

build-linux-amd64: _pre-make _build-linux-amd64 _post-make
_build-linux-amd64: _deps _fmt build-linux-amd64_
build-linux-amd64_:
	@env _GOOS=linux _GOARCH=amd64 make build_

build-darwin-amd64: _pre-make _build-darwin-amd64 _post-make
_build-darwin-amd64: _deps _fmt build-darwin-amd64_
build-darwin-amd64_:
	@env _GOOS=darwin _GOARCH=amd64 make build_

install: _pre-make version-noarch _install _post-make
_install: _deps _fmt
	@echo "target: install"
	@printf "  ...installing rexray $(V_ARCH)..."; \
		cd $(BASEDIR); \
		go clean -i $(VERSIONPKG); \
		go install $(GOFLAGS) $(LDFLAGS) $(NV); \
		$(PRINT_STATUS); \
		if [ "$$EC" -eq "0" ]; then \
			FILE=$(GOPATH)/bin/rexray; \
			BYTES=$$($(STAT_FILE_SIZE)); \
			SIZE=$$(($$BYTES / 1024 / 1024)); \
			printf "\nThe REX-Ray binary is $${SIZE}MB and located at:\n\n"; \
			printf "  $$FILE\n\n"; \
		fi

fmt: _pre-make _fmt _post-make 

_fmt:
	@echo "target: fmt"
	@printf "  ...formatting rexray..."; \
		cd $(BASEDIR); \
		go fmt $(NV); \
		$(PRINT_STATUS)

fix: _pre-make _fix _post-make

_fix:
	@echo "target: fix"
	@printf "  ...fixing rexray..."; \
		cd $(BASEDIR); \
		go fmt $(NV); \
		$(PRINT_STATUS)

test: _pre-make _test _post-make

_test: _install
	@echo "target: test"
	@printf "  ...testing rexray..."; \
		cd $(BASEDIR); \
		go test $(GOFLAGS) $(NV); \
		$(PRINT_STATUS)

bench: _pre-make _bench _post-make

_bench: _install
	@echo "target: bench"
	@printf "  ...benchmarking rexray..."; \
		cd $(BASEDIR); \
		go test -run=NONE -bench=. $(GOFLAGS) $(NV); \
		$(PRINT_STATUS)

clean: _pre-make _clean _post-make

_clean:
	@echo "target: clean"
	@printf "  ...cleaning rexray..."; \
		cd $(BASEDIR); \
		go clean $(GOFLAGS) -i $(NV); \
		$(PRINT_STATUS)
		
version:
	@echo SemVer: $(V_SEMVER)
	@echo Binary: $(V_ARCH)
	@echo Branch: $(V_BRANCH)
	@echo Commit: $(V_SHA_LONG)
	@echo Formed: $(V_BUILD_DATE)
	
version-noarch:
	@echo SemVer: $(V_SEMVER)
	@echo Branch: $(V_BRANCH)
	@echo Commit: $(V_SHA_LONG)
	@echo Formed: $(V_BUILD_DATE)
	@echo

rpm: install
	@echo "target: rpm"
	@rm -fr $(RPMBUILD)
	
	@mkdir -p $(RPMBUILD)/{RPMS,SRPMS,SPECS,tmp}
	@ln -s $(CWD) $(RPMBUILD)/BUILD
	@ln -s $(CWD) $(RPMBUILD)/SOURCES
	@sed -e 's|$${RPMBUILD}|$(RPMBUILD)|g' \
		-e 's|$${GOPATH}|$(GOPATH)|g' \
		$(CWD)/rexray.spec > $(RPMBUILD)/SPECS/rexray.spec

	@printf "  ...building rpm..."; \
		rpmbuild -ba --quiet SPECS/rexray.spec; \
		$(PRINT_STATUS); \
		if [ "$$EC" -eq "0" ]; then \
			FILE=$$(readlink -f $$(find $(RPMBUILD)/RPMS -name *.rpm)); \
			BYTES=$$($(STAT_FILE_SIZE)); \
			SIZE=$$(($$BYTES / 1024 / 1024)); \
			printf "\nThe REX-Ray RPM is $${SIZE}MB and located at:\n\n"; \
			printf "  $$FILE\n\n"; \
		fi

.PHONY: all install build build_ build-all deps fmt fix clean version
.NOTPARALLEL: all test clean deps _deps fmt _fmt fix pre-make _pre-make post-make _post-make build build-all_ install rpm
