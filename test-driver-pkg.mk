# this Makefile can be reused by driver packages for testing

SHELL := $(shell env which bash)

# the name of the driver is grok'd from the basename
# of the parent directory
DRIVER := $(shell basename $$(dirname $$(pwd)))

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

# the test binary
$(DRIVER).test: $(SRCS)
	go test -tags $(patsubst %.test,%,$@) -cover -c -o $@ .

# the coverage file
$(DRIVER).test.out: $(DRIVER).test
	./$< -test.coverprofile $@

build: $(DRIVER).test

test: $(DRIVER).test.out

clean:
	rm -f $(DRIVER).test $(DRIVER).test.out

.PHONY: clean
