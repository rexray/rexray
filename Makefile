all: install

include .gomk/go.mk

deps: $(GO_DEPS)

build: $(GO_BUILD)

install: $(GO_INSTALL)

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
	$(ENV) GOMK_TOOLS_ENABLE=1 GO_TAGS='run mock driver' $(MAKE) install
	$(ENV) GOMK_TOOLS_ENABLE=1 GO_TAGS='run mock driver' $(MAKE) test

run-debug: | $(ENV)
	$(ENV) LIBSTORAGE_DEBUG=true $(MAKE) run

run-tls: | $(ENV)
	$(ENV) LIBSTORAGE_TESTRUN_TLS=true $(MAKE) run

run-tls-debug: | $(ENV)
	$(ENV) LIBSTORAGE_TESTRUN_TLS=true $(MAKE) run-debug

.PHONY: all install build deps \
		test test-build test-clean \
		test-mock test-build-mock \
		cover cover-clean \
		dist dist-clean \
		clean clobber \
		run run-debug run-tls run-tls-debug \
		$(GO_PHONY)
