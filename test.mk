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

# add clean-tests as a dependency of clean
clean: clean-tests

# add TEST_DRIVERS_COVR and TEST_FRAMEWORK_COVR to COVERAGE_SRCS so
# the codoecov.mk file knows the source of the coverage reports
COVERAGE_SRCS := $(TEST_DRIVERS_COVR) $(TEST_FRAMEWORK_COVR)
