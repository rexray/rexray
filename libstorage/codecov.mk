################################################################################
##                                  COVERAGE                                  ##
################################################################################
coverage.out: $(TEST_FRAMEWORK_COVR) $(TEST_DRIVERS_COVR)
	printf "mode: set\n" > $@
	$(foreach f,$?,grep -v "mode: set" $f >> $@ &&) true

clean-coverage:
	rm -f coverage.out

# add clean-coverage as a dependency of clean
clean: clean-coverage

COVERAGE_IMPORTS := github.com/onsi/gomega \
  github.com/onsi/ginkgo \
  golang.org/x/tools/cmd/cover

COVERAGE_IMPORTS_PATHS := $(addprefix $(GOPATH)/src/,$(COVERAGE_IMPORTS))

$(COVERAGE_IMPORTS_PATHS):
	go get $(subst $(GOPATH)/src/,,$@)

cover: coverage.out | $(COVERAGE_IMPORTS_PATHS)
	curl -sSL https://codecov.io/bash | bash -s -- -f $<

.PHONY: cover
