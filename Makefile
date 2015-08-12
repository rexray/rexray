.PHONY: all test clean deps build install

ROOT_DIR := $(shell pwd)
export GOBIN := $(GOPATH)/bin
export GO15VENDOREXPERIMENT := 1
SEMVER := $(shell git describe --tags | sed -E 's/^v.?//g')
GOFLAGS := $(GOFLAGS)
NV := $(shell glide novendor)
VERSIONPKG := github.com/emccode/rexray/version
LDFLAGS := -ldflags "-X $(VERSIONPKG).Version=$(SEMVER)" 
RPMBUILD := $(ROOT_DIR)/.rpmbuild
EMCCODE := $(GOPATH)/src/github.com/emccode
PRINT_STATUS = export EC=$$?; if [ "$$EC" -eq "0" ]; then printf "SUCCESS!\n"; else exit $$EC; fi
STAT_FILE_SIZE = stat --format '%s' $$FILE 2> /dev/null || stat -f '%z' $$FILE 2> /dev/null

all: install
	
deps: 
	@echo "target: deps"
	@printf "  ...downloading go dependencies..."
	@go get -d $(GOFLAGS) $(NV); \
		glide -q up 2> /dev/null; \
		$(PRINT_STATUS)

build-nodeps: fmt
	@echo "target: build-nodeps"
	@printf "  ...building rexray..."
	@go build $(GOFLAGS) $(LDFLAGS) $(NV); \
		$(PRINT_STATUS)

build: deps fmt
	@echo "target: build"
	@printf "  ...building rexray..."
	@go build $(GOFLAGS) $(LDFLAGS) $(NV); \
		$(PRINT_STATUS)

install: deps fmt
	@echo "target: install"
	@printf "  ...building and installing rexray..."; \
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

fmt:
	@echo "target: fmt"
	@printf "  ...formatting rexray..."; \
		go fmt $(NV); \
		$(PRINT_STATUS)
		
fix:
	@echo "target: fix"
	@printf "  ...fixing rexray..."; \
		go fmt $(NV); \
		$(PRINT_STATUS)

test: install
	@echo "target: test"
	@printf "  ...testing rexray..."; \
		go test $(GOFLAGS) $(NV); \
		$(PRINT_STATUS)

bench: install
	@echo "target: bench"
	@printf "  ...benchmarking rexray..."; \
		go test -run=NONE -bench=. $(GOFLAGS) $(NV); \
		$(PRINT_STATUS)

clean:
	@echo "target: clean"
	@go clean $(GOFLAGS) -i $(NV); \
		$(PRINT_STATUS)

rpm: install
	@echo "target: rpm"
	@rm -fr $(RPMBUILD)
	
	@mkdir -p $(RPMBUILD)/{RPMS,SRPMS,SPECS,tmp}
	@ln -s $(ROOT_DIR) $(RPMBUILD)/BUILD
	@ln -s $(ROOT_DIR) $(RPMBUILD)/SOURCES
	@sed -e 's|$${RPMBUILD}|$(RPMBUILD)|g' \
		-e 's|$${GOPATH}|$(GOPATH)|g' \
		$(ROOT_DIR)/rexray.spec > $(RPMBUILD)/SPECS/rexray.spec

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
