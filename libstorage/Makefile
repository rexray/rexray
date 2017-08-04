SHELL := $(shell which bash)

all: build

# include the golang makefile content
include golang.mk

# the root project's vendor directory
VENDOR := ../vendor
vendor: $(VENDOR)

# the target for making the dependencies
$(VENDOR):
	$(MAKE) -C $(@D) $(@F)

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

# a list of this package's Go sources and Go sources
# for packages on which this package depends
SRCS := $(shell \
  $(GOLIST) -f \
  '{{with $$p := .}}{{if not $$p.Standard}}{{range $$f := $$p.GoFiles}}{{printf "%s/%s\n" $$p.Dir $$f }}{{end}}{{end}}{{end}}' \
  . \
  $$($(GOLIST) -f '{{if not .Standard}}{{join .Deps "\n"}}{{end}}') 2> /dev/null)

# define a list of the generated source files and a target
# for generating the files if missing
GENERATED_SRCS := ./api/utils/schema/schema_generated.go
$(GENERATED_SRCS):
	$(MAKE) -C $(@D) $(@F)

# the target to build
PROG := $(shell go list -f '{{.Target}}')

$(PROG): $(SRCS) $(GENERATED_SRCS) $(GLIDE_LOCK) | $(VENDOR)
	$(GOINSTALL) -v

build: $(PROG)

# include the test.mk file for testing targets
include test.mk

# include the codecov.mk file for coverage targets
include codecov.mk

################################################################################
##                                  CLEAN                                     ##
################################################################################

clean:
	$(GOCLEAN) -i -x

.PHONY: clean vendor
