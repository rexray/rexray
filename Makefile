.PHONY: all test clean deps build install

ROOT_DIR := $(shell pwd)
ifeq ($(origin GOPATH), undefined)
	export GOPATH := $(ROOT_DIR)/.build
endif
export GOBIN := $(GOPATH)/bin
GOFLAGS := $(GOFLAGS) -gcflags "-N -l"
EMCCODE := $(GOPATH)/src/github.com/emccode
GOAMZ_WRKDIR := $(GOPATH)/src/github.com/goamz/goamz
GOAMZ_GITDIR := $(GOAMZ_WRKDIR)/.git
GOAMZ_SNAP = https://github.com/clintonskitson/goamz.git
GOAMZ_SNAP_BRANCH = snapcopy
GOAMZ_GIT = git --git-dir=$(GOAMZ_GITDIR) --work-tree=$(GOAMZ_WRKDIR)
GOAMZ_BRANCH = $(shell $(GOAMZ_GIT) branch --list | grep '^*' | tr -d '* ')
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
	@go get -d $(GOFLAGS) ./...

deps_goamz:
	@$(GOAMZ_SNAPCOPY_REMOTE) || $(GOAMZ_SNAPCOPY_REMOTE_ADD)
	@if [ "$(GOAMZ_BRANCH)" != "snapcopy" ]; then \
		$(GOAMZ_SNAPCOPY_CHECKOUT); \
	fi

deps: deps_rexray_sources deps_go_get deps_goamz

build: deps
	@go build $(GOFLAGS) ./...

install: deps
	@cd $(EMCCODE)/rexray; \
		go install $(GOFLAGS) ./...; \
		cd $(ROOT_DIR)

test: install
	@go test $(GOFLAGS) ./...

bench: install
	@go test -run=NONE -bench=. $(GOFLAGS) ./...

clean:
	@go clean $(GOFLAGS) -i ./...