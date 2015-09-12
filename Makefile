WD := $(shell pwd)
export MAKEFLAGS := $(MAKEFLAGS) --no-print-directory -k
export GO15VENDOREXPERIMENT ?= 1
export GOOS ?= $(shell go version | awk '{print $$4}' | tr '/' ' ' | awk '{print $$1}')
export GOARCH ?= $(shell go version | awk '{print $$4}' | tr '/' ' ' | awk '{print $$2}')
_GOOS ?= $(GOOS)
_GOARCH ?= $(GOARCH)
ARCH = $(_GOOS)_$(_GOARCH)
GITDSC := $(shell git describe --long)
TRAVIS_BRANCH ?= $(shell git branch | grep '*' | awk '{print $$2}')
BRANCH = $(TRAVIS_BRANCH)
TGTVER := $(shell cat VERSION | tr -d " \n\r\t")
BLDDTE := $(shell date +%s)
CMTDTE := $(shell git show HEAD -s --format=%ct)
CMTHSH := $(shell git show HEAD -s --format=%H)
RELDTE := $(shell date -u "+%Y-%m-%d")
SHA := $(shell git show HEAD -s --format=%h)
GOFLAGS := $(GOFLAGS)
GLIDE := $(GOPATH)/bin/glide
NV := $$($(GLIDE) novendor)
BASEPKG := github.com/emccode/rexray
BASEDIR := $(GOPATH)/src/$(BASEPKG)
BASEDIR_NAME := $(shell basename $(BASEDIR))
BASEDIR_PARENTDIR := $(shell dirname $(BASEDIR))
BASEDIR_TEMPMVLOC := $(BASEDIR_PARENTDIR)/.$(BASEDIR_NAME)-$(shell date +%s)
VERSIONPKG := $(BASEPKG)/version_info
LDF_GITDSC := -X $(VERSIONPKG).GitDescribe=$(GITDSC)
LDF_BRANCH := -X $(VERSIONPKG).BranchName=$(BRANCH)
LDF_TGTVER := -X $(VERSIONPKG).TargetVersion=$(TGTVER)
LDF_BLDDTE := -X $(VERSIONPKG).BuildDateEpochStr=$(BLDDTE)
LDF_CMTDTE := -X $(VERSIONPKG).CommitDateEpochStr=$(CMTDTE)
LDF_CMTHSH := -X $(VERSIONPKG).CommitHash=$(CMTHSH)
LDF_GOARCH = -X $(VERSIONPKG).BinArchStr=$(ARCH)
LDFLAGS = -ldflags "$(LDF_GITDSC) $(LDF_BRANCH) $(LDF_TGTVER) $(LDF_BLDDTE) $(LDF_CMTDTE) $(LDF_CMTHSH) $(LDF_GOARCH)" 
RPMBUILD := $(WD)/.rpmbuild
EMCCODE := $(GOPATH)/src/github.com/emccode
PRINT_STATUS = export EC=$$?; cd $(WD); if [ "$$EC" -eq "0" ]; then printf "SUCCESS!\n"; else exit $$EC; fi
STAT_FILE_SIZE = stat --format '%s' $$FILE 2> /dev/null || stat -f '%z' $$FILE 2> /dev/null

all: install

_pre-make:
	@if [ "$(WD)" != "$(BASEDIR)" ]; then \
		if [ -e "$(BASEDIR)" ]; then \
			mv $(BASEDIR) $(BASEDIR_TEMPMVLOC); \
		fi; \
		mkdir -p "$(BASEDIR_PARENTDIR)"; \
		ln -s "$(WD)" "$(BASEDIR)"; \
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
	@printf "  ...building rexray $(ARCH)..."; \
		cd $(BASEDIR); \
		FILE=.bin/$(ARCH)/rexray; \
		env GOOS=$(_GOOS) GOARCH=$(_GOARCH) go build -o $$FILE $(GOFLAGS) $(LDFLAGS) ./rexray; \
		$(PRINT_STATUS); \
		if [ "$$EC" -eq "0" ]; then \
			BYTES=$$($(STAT_FILE_SIZE)); \
			SIZE=$$(($$BYTES / 1024 / 1024)); \
			printf "\nThe REX-Ray binary is $${SIZE}MB and located at:\n\n"; \
			printf "  $$FILE\n\n"; \
		fi

build-all: _pre-make _deps _fmt build-all_ _post-make
build-all_: build-linux-386_ build-linux-amd64_ build-darwin-amd64_
	@SEMVER=$$(.bin/$(GOOS)_$(GOARCH)/rexray version | grep SemVer | awk '{print $$2}'); \
		for BIN in $$(find .bin -type f -name "rexray"); do \
			BINDIR=$$(dirname $$BIN); \
			FARCH=$$(echo $$BINDIR | cut -c6-); \
			TARBALL=rexray-$$FARCH-$$SEMVER.tar.gz; \
			cd $$BINDIR; \
			tar -czf $$TARBALL rexray; \
			cd - > /dev/null; \
		done; \
		sed -e 's/$${SEMVER}/'"$$SEMVER"'/g' \
			-e 's|$${DSCRIP}|'"$$SEMVER"'.Branch.$(BRANCH).Sha.$(CMTHSH)|g' \
			-e 's/$${RELDTE}/$(RELDTE)/g' \
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

install: _pre-make _install _post-make
_install: _deps _fmt
	@echo "target: install"
	@printf "  ...installing rexray $$GOOS-$$GOARCH..."; \
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

rpm: install
	@echo "target: rpm"
	@rm -fr $(RPMBUILD)
	
	@mkdir -p $(RPMBUILD)/{RPMS,SRPMS,SPECS,tmp}
	@ln -s $(WD) $(RPMBUILD)/BUILD
	@ln -s $(WD) $(RPMBUILD)/SOURCES
	@sed -e 's|$${RPMBUILD}|$(RPMBUILD)|g' \
		-e 's|$${GOPATH}|$(GOPATH)|g' \
		$(WD)/rexray.spec > $(RPMBUILD)/SPECS/rexray.spec

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

.PHONY: all install build build_ build-all deps fmt fix clean
.NOTPARALLEL: all test clean deps _deps fmt _fmt fix pre-make _pre-make post-make _post-make build build-all_ install rpm
