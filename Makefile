SHELL := $(shell which bash)

all: build

# include the golang makefile content
include golang.mk

# include the glide makefile content
include glide.mk

# a list of this package's Go sources and Go sources
# for packages on which this package depends
SRCS := $(shell \
  go list -f \
  '{{with $$p := .}}{{if not $$p.Standard}}{{range $$f := $$p.GoFiles}}{{printf "%s/%s\n" $$p.Dir $$f }}{{end}}{{end}}{{end}}' \
  . \
  $$(go list -f '{{if not .Standard}}{{join .Deps "\n"}}{{end}}') 2> /dev/null)

# define a list of the generated source files and a target
# for generating the files if missing
GENERATED_SRCS := ./api/api_version_generated.go \
  ./api/utils/schema/schema_generated.go
$(GENERATED_SRCS):
	$(MAKE) -C $(@D) $(@F)

# the target to build
PROG := $(shell go list -f '{{.Target}}')

$(PROG): $(SRCS) $(GENERATED_SRCS) $(GLIDE_LOCK) | $(VENDOR)
	go install -v .

build: $(PROG)

# include the test.mk file for testing targets
include test.mk

# include the codecov.mk file for coverage targets
include codecov.mk

################################################################################
##                                  CLEAN                                     ##
################################################################################

clean:
	go clean -i -x
	$(MAKE) -C api clean

.PHONY: clean
