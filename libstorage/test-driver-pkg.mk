# this Makefile can be reused by driver packages for testing

SHELL := $(shell env which bash)

# the path of the parent directory
PARENT := $(shell dirname $$(pwd))

# the name of the driver is grok'd from the basename
# of the parent directory
DRIVER := $(shell basename "$(PARENT)")

# add the driver name to the list of build tags
BUILD_TAGS += $(DRIVER)

# define aliases for the go commands
GOBUILD := go build
GOCLEAN := go clean
GOINSTALL := go install
GOLIST := go list
GOTEST := go test

# update the go commands if there are build tags to consider
ifneq (,$(strip $(BUILD_TAGS)))
GOBUILD += -tags '$(BUILD_TAGS)'
GOCLEAN += -tags '$(BUILD_TAGS)'
GOINSTALL += -tags '$(BUILD_TAGS)'
GOLIST += -tags '$(BUILD_TAGS)'
GOTEST += -tags '$(BUILD_TAGS)'
endif

# update the go commands if there is a pkgdir to consider
ifneq (,$(strip $(PKGDIR)))
GOBUILD += -pkgdir '$(PKGDIR)'
GOCLEAN += -pkgdir '$(PKGDIR)'
GOINSTALL += -pkgdir '$(PKGDIR)'
GOLIST += -pkgdir '$(PKGDIR)'
GOTEST += -pkgdir '$(PKGDIR)'
endif

# if SKIP_SRCS=1 then do not attempt to determine the source files
# on which the test binary depends. otherwise SRCS is a list of
# this package's Go sources, Go test sources, Go Xtest sources
# (test sources in this directory but belonging to a different package)
#, and Go sources for packages on which this package depends
ifneq (1,$(SKIP_SRCS))
SRCS := $(shell \
  $(GOLIST) -f \
  '{{join .GoFiles "\n"}}{{"\n"}}{{join .TestGoFiles "\n"}}{{"\n"}}{{join .XTestGoFiles "\n"}}' \
  && \
  $(GOLIST) -f \
  '{{with $$p := .}}{{if not $$p.Standard}}{{range $$f := $$p.GoFiles}}{{printf "%s/%s\n" $$p.Dir $$f }}{{end}}{{end}}{{end}}' \
  $$($(GOLIST) -f '{{if not .Standard}}{{join .Deps "\n"}}{{end}}' \
  . \
  $$($(GOLIST) -f '{{join .TestImports "\n"}}{{"\n"}}{{join .XTestImports "\n"}}') | sort -u) 2> /dev/null)
endif

# the packages the test covers
COVERPKG := $(PARENT),$(PARENT)/storage,$(PARENT)/executor
ifneq (,$(strip $(wildcard $(PARENT)/utils)))
COVERPKG := $(COVERPKG),$(PARENT)/utils
endif
COVERPKG := $(subst $(GOPATH)/src/,,$(COVERPKG))

# the test binary
$(DRIVER).test: $(SRCS)
	$(GOTEST) -i -coverpkg '$(COVERPKG)' -c -o $@

# the coverage file
$(DRIVER).test.out: $(DRIVER).test
	./$< -test.coverprofile $@

build-tests: $(DRIVER).test

test: $(DRIVER).test.out

clean-tests:
	rm -f $(DRIVER).test $(DRIVER).test.out

clean: clean-tests

.PHONY: clean clean-tests
