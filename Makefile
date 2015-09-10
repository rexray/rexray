.PHONY: all install build deps fmt fix clean
.NOTPARALLEL: all test clean deps fmt fix pre-build build post-build install rpm

WD := $(shell pwd)
export MAKEFLAGS := $(MAKEFLAGS) -k
export GOBIN := $(GOPATH)/bin
export GO15VENDOREXPERIMENT := 1
GITDSC := $(shell git describe --long)
BRANCH := $(shell git branch | grep '*' | awk '{print $$2}')
TGTVER := $(shell cat VERSION | tr -d " \n\r\t")
BLDDTE := $(shell date +%s)
CMTDTE := $(shell git show HEAD -s --format=%ct)
CMTHSH := $(shell git show HEAD -s --format=%H)
GOFLAGS := $(GOFLAGS)
GLIDE := $(GOBIN)/glide
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
LDFLAGS := -ldflags "$(LDF_GITDSC) $(LDF_BRANCH) $(LDF_TGTVER) $(LDF_BLDDTE) $(LDF_CMTDTE) $(LDF_CMTHSH)" 
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

_build: _deps _fmt
	@echo "target: build"
	@printf "  ...building rexray..."; \
		cd $(BASEDIR); \
		go build $(GOFLAGS) $(LDFLAGS) $(NV); \
		$(PRINT_STATUS)

install: _pre-make _install _post-make

_install: _deps _fmt
	@echo "target: install"
	@printf "  ...building and installing rexray..."; \
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
