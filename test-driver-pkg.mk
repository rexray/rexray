# this Makefile can be reused by driver packages for testing

SHELL := $(shell env which bash)

# the path of the parent directory
PARENT := $(shell dirname $$(pwd))

# the name of the driver is grok'd from the basename
# of the parent directory
DRIVER := $(shell basename "$(PARENT)")

# a list of this package's Go sources, Go test sources, Go Xtest sources
# (test sources in this directory but belonging to a different package)
#, and Go sources for packages on which this package depends
SRCS := $(shell \
  go list -tags $(DRIVER) -f \
  '{{join .GoFiles "\n"}}{{"\n"}}{{join .TestGoFiles "\n"}}{{"\n"}}{{join .XTestGoFiles "\n"}}' \
  && \
  go list -tags $(DRIVER) -f \
  '{{with $$p := .}}{{if not $$p.Standard}}{{range $$f := $$p.GoFiles}}{{printf "%s/%s\n" $$p.Dir $$f }}{{end}}{{end}}{{end}}' \
  $$(go list -tags $(DRIVER) -f '{{if not .Standard}}{{join .Deps "\n"}}{{end}}' \
  . \
  $$(go list -tags $(DRIVER) -f '{{join .TestImports "\n"}}{{"\n"}}{{join .XTestImports "\n"}}') | sort -u))

# the packages the test covers
COVERPKG := $(PARENT),$(PARENT)/storage,$(PARENT)/executor
ifneq (,$(strip $(wildcard $(PARENT)/utils)))
COVERPKG := $(COVERPKG),$(PARENT)/utils
endif
COVERPKG := $(subst $(GOPATH)/src/,,$(COVERPKG))

# the test binary
$(DRIVER).test: $(SRCS)
	go test -tags $(patsubst %.test,%,$@) -coverpkg '$(COVERPKG)' -c -o $@ .

# the coverage file
$(DRIVER).test.out: $(DRIVER).test
	./$< -test.coverprofile $@

build-tests: $(DRIVER).test

test: $(DRIVER).test.out

clean: $(CLEAN)
	rm -f $(DRIVER).test $(DRIVER).test.out

.PHONY: clean
