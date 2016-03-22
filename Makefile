all: install

include .gomk/go.mk

deps: $(GO_DEPS)

build: deps $(GO_SRC_TOOL_MARKERS) $(GO_BUILD)

install: $(GO_INSTALL)

test: $(GO_TEST)

test-build: $(GO_TEST_BUILD)

test-clean: $(GO_TEST_CLEAN)

cover: $(GO_COVER)

cover-clean: $(GO_COVER_CLEAN)

package: $(GO_PACKAGE)

package-clean: $(GO_PACKAGE_CLEAN)

clean: $(GO_CLEAN)

clobber: $(GO_CLOBBER)

run:
	go test -tags 'run mock driver' -v

run-debug:
	env LIBSTORAGE_DEBUG=true $(MAKE) run

run-tls:
	env LIBSTORAGE_TESTRUN_TLS=true $(MAKE) run

run-tls-debug:
	env LIBSTORAGE_DEBUG=true LIBSTORAGE_TESTRUN_TLS=true $(MAKE) run

.PHONY: all install build deps \
		test test-build test-clean \
		test-mock test-build-mock \
		cover cover-clean \
		package package-clean \
		clean clobber \
		run run-debug run-tls run-tls-debug \
		$(GO_PHONY)
