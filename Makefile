all: install

include .gomk/go.mk

deps: $(GO_DEPS)

build: $(GO_BUILD)

install: $(GO_INSTALL)

tests: $(GO_TEST)

tests-build: $(GO_TEST_BUILD)

tests-clean: $(GO_TEST_CLEAN)

cover: $(GO_COVER)

cover-clean: $(GO_COVER_CLEAN)

package: $(GO_PACKAGE)

package-clean: $(GO_PACKAGE_CLEAN)

clean: $(GO_CLEAN)

clobber: $(GO_CLOBBER)

.PHONY: all install build deps \
		test test-build test-clean \
		cover cover-clean \
		package package-clean \
		clean clobber \
		$(GO_PHONY)
