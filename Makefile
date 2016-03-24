# the space-delimited list of drivers for which to build the libstorage
# server, client(s), and executor(s)
DRIVERS := mock

all: build

include .gomk/go.mk

deps: $(GO_DEPS)

build: $(GO_BUILD)

install: $(GO_BUILD)

build-executors: $(LIBSTORAGE_EXECUTORS)

test-build: $(GO_TEST_BUILD)

test: $(GO_TEST)

test-clean: $(GO_TEST_CLEAN)

cover: $(GO_COVER)

cover-clean: $(GO_COVER_CLEAN)

dist: $(GO_PACKAGE)

dist-clean: $(GO_PACKAGE_CLEAN)

clean: $(GO_CLEAN)

clobber: $(GO_CLOBBER)

run: | $(ENV)
	$(ENV) GOMK_TOOLS_ENABLE=1 GO_TAGS='$(DRIVERS) driver' $(MAKE)
	$(ENV) GOMK_TOOLS_ENABLE=1 GO_TAGS='run $(DRIVERS) driver' $(MAKE) test

run-debug: | $(ENV)
	$(ENV) LIBSTORAGE_DEBUG=true $(MAKE) run

run-tls: | $(ENV)
	$(ENV) LIBSTORAGE_TESTRUN_TLS=true $(MAKE) run

run-tls-debug: | $(ENV)
	$(ENV) LIBSTORAGE_TESTRUN_TLS=true $(MAKE) run-debug

.PHONY: all \
		cover cover-clean \
		dist dist-clean \
		clean clobber \
		run run-debug run-tls run-tls-debug \
		$(GO_PHONY)
