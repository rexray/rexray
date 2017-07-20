SHELL := $(shell which bash)

all: build

# include the golang makefile content
include golang.mk

# include the glide makefile content
include glide.mk


################################################################################
##                                   BUILD                                    ##
################################################################################

# a list of this package's Go sources and Go sources
# for packages on which this package depends
#
# do not generate the sources unless the vendor directory is present,
# otherwise a myriad of errors will occur due to missing dependencies
# during the execution of the "go list" tool
ifneq (,$(strip $(wildcard vendor)))
SRCS := $(shell \
  go list -f \
  '{{with $$p := .}}{{if not $$p.Standard}}{{range $$f := $$p.GoFiles}}{{printf "%s/%s\n" $$p.Dir $$f }}{{end}}{{end}}{{end}}' \
  . \
  $$(go list -f '{{if not .Standard}}{{join .Deps "\n"}}{{end}}'))
endif

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

################################################################################
##                                   TESTS                                    ##
################################################################################

# test all of the drivers that have a Makefile that match the pattern
# ./drivers/storage/%/tests/Makefile. The % is extracted as the name
# of the driver
TEST_DRIVERS := $(strip $(patsubst ./drivers/storage/%/tests/Makefile,\
  %,\
  $(wildcard ./drivers/storage/*/tests/Makefile)))

# a list of the driver packages to test
TEST_DRIVERS_PKGS := $(foreach d,\
  $(TEST_DRIVERS),\
  ./drivers/storage/$d/tests)

# a list of the driver packages' test binaries
TEST_DRIVERS_BINS := $(foreach d,\
  $(TEST_DRIVERS),\
  ./drivers/storage/$d/tests/$d.test)

# only the VFS driver is used for coverage at this level. it's possible
# to visit each driver's tests package and produce coverage files
# directly from there
TEST_DRIVERS_COVR := ./drivers/storage/vfs/tests/vfs.test.out

# a list of the framework packages to test
TEST_FRAMEWORK_PKGS :=  ./api/context \
  ./api/server/auth \
  ./api/types \
  ./api/utils/filters \
  ./api/utils/schema \
  ./api/utils

# a list of the framework packages' test binaries
TEST_FRAMEWORK_BINS := $(foreach p,\
  $(TEST_FRAMEWORK_PKGS),\
  $p/$(notdir $p).test)

# a list of the framework packages' test coverage output
TEST_FRAMEWORK_COVR := $(addsuffix .out,$(TEST_FRAMEWORK_BINS))

# the recipe for building the test binaries
$(TEST_DRIVERS_BINS) $(TEST_FRAMEWORK_BINS):
	$(MAKE) -C $(@D) build-tests

# the recipe for executing the test binaries
$(TEST_DRIVERS_COVR) $(TEST_FRAMEWORK_COVR): %.out: %
	$(MAKE) -C $(@D) test

# builds all the tests
build-tests: $(TEST_FRAMEWORK_BINS) $(TEST_DRIVERS_BINS)

# executes the framework test binaries and the vfs test binary
test: $(TEST_FRAMEWORK_COVR) $(TEST_DRIVERS_COVR)

# a target for cleaning all the test binaries and coverage files
CLEAN_TESTS := $(addprefix clean-,$(TEST_FRAMEWORK_PKGS) $(TEST_DRIVERS_PKGS))
$(CLEAN_TESTS):
	$(MAKE) -C $(patsubst clean-%,%,$@) clean
clean-tests: $(CLEAN_TESTS)

################################################################################
##                                  COVERAGE                                  ##
################################################################################
coverage.out: $(TEST_FRAMEWORK_COVR) $(TEST_DRIVERS_COVR)
	printf "mode: set\n" > $@
	$(foreach f,$?,grep -v "mode: set" $f >> $@ &&) true

clean-coverage:
	rm -f coverage.out

COVERAGE_IMPORTS := github.com/onsi/gomega \
  github.com/onsi/ginkgo \
  golang.org/x/tools/cmd/cover

COVERAGE_IMPORTS_PATHS := $(addprefix $(GOPATH)/src/,$(COVERAGE_IMPORTS))

$(COVERAGE_IMPORTS_PATHS):
	go get $(subst $(GOPATH)/src/,,$@)

cover: coverage.out | $(COVERAGE_IMPORTS_PATHS)
	curl -sSL https://codecov.io/bash | bash -s -- -f $<

################################################################################
##                                  CLEAN                                     ##
################################################################################

clean: clean-tests clean-coverage
	go clean -i -x
	$(MAKE) -C api clean

.PHONY: clean cover
