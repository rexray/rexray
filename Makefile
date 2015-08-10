.PHONY: all test clean deps build install

ROOT_DIR := $(shell pwd)
ifeq ($(origin GOPATH), undefined)
	export GOPATH := $(ROOT_DIR)/.build
endif
export GOBIN := $(GOPATH)/bin
GOFLAGS := $(GOFLAGS) -gcflags "-N -l"
RPMBUILD := $(ROOT_DIR)/.rpmbuild
EMCCODE := $(GOPATH)/src/github.com/emccode
GOAMZ_WRKDIR := $(GOPATH)/src/github.com/goamz/goamz
GOAMZ_GITDIR := $(GOAMZ_WRKDIR)/.git
GOAMZ_SNAP = https://github.com/clintonskitson/goamz.git
GOAMZ_SNAP_BRANCH = snapcopy
GOAMZ_GIT = git --git-dir=$(GOAMZ_GITDIR) --work-tree=$(GOAMZ_WRKDIR)
GOAMZ_BRANCH = $(shell $(GOAMZ_GIT) branch --list | grep '^*' | awk '{print $$2}')
GOAMZ_SNAPCOPY_REMOTE = $(GOAMZ_GIT) remote -v | grep snapcopy &> /dev/null
GOAMZ_SNAPCOPY_REMOTE_ADD = $(GOAMZ_GIT) remote add -f snapcopy $(GOAMZ_SNAP) &> /dev/null
GOAMZ_SNAPCOPY_CHECKOUT = $(GOAMZ_GIT) checkout -f snapcopy &> /dev/null

all: install

deps_rexray_sources:
	@mkdir -p $(EMCCODE)
	@if [ ! -e "$(EMCCODE)/rexray" ]; then \
		ln -s $(ROOT_DIR) $(EMCCODE)/rexray &> /dev/null; \
	fi
	
deps_go_get:
	@cd $(EMCCODE)/rexray; \
		go get -d $(GOFLAGS) ./...; \
		cd $(ROOT_DIR)

deps_goamz:
	@$(GOAMZ_SNAPCOPY_REMOTE) || $(GOAMZ_SNAPCOPY_REMOTE_ADD)
	@if [ "$(GOAMZ_BRANCH)" != "snapcopy" ]; then \
		$(GOAMZ_SNAPCOPY_CHECKOUT); \
	fi

deps: deps_rexray_sources deps_go_get deps_goamz

rpm: install
	@rm -fr $(RPMBUILD)
	
	@mkdir -p $(RPMBUILD)/{RPMS,SRPMS,SPECS,tmp}
	@ln -s $(ROOT_DIR) $(RPMBUILD)/BUILD
	@ln -s $(ROOT_DIR) $(RPMBUILD)/SOURCES
	@sed -e 's|$${RPMBUILD}|$(RPMBUILD)|g' \
		-e 's|$${GOPATH}|$(GOPATH)|g' \
		$(ROOT_DIR)/rexray.spec > $(RPMBUILD)/SPECS/rexray.spec

	@cd $(RPMBUILD); \
		rpmbuild -ba SPECS/rexray.spec; \
		cd $(ROOT_DIR)

build: deps
	@cd $(EMCCODE)/rexray; \
		go build $(GOFLAGS) ./...; \
		cd $(ROOT_DIR)

install: deps
	@cd $(EMCCODE)/rexray; \
		go install $(GOFLAGS) ./...; \
		cd $(ROOT_DIR)

test: install
	@cd $(EMCCODE)/rexray; \
		go test $(GOFLAGS) ./...; \
		cd $(ROOT_DIR)

bench: install
	@cd $(EMCCODE)/rexray; \
		go test -run=NONE -bench=. $(GOFLAGS) ./...; \
		cd $(ROOT_DIR)

clean:
	@cd $(EMCCODE)/rexray; \
		go clean $(GOFLAGS) -i ./...; \
		cd $(ROOT_DIR)