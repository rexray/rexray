# this Makefile can be reused by framework packages for testing

SHELL := $(shell env which bash)

# the name of the directory that contains this Makefile
PKG := $(shell basename $$(pwd))

# a list of this package's Go sources, Go test sources, Go Xtest sources
# (test sources in this directory but belonging to a different package)
#, and Go sources for packages on which this package depends
SRCS := $(shell \
  go list -f \
  '{{join .GoFiles "\n"}}{{"\n"}}{{join .TestGoFiles "\n"}}{{"\n"}}{{join .XTestGoFiles "\n"}}' \
  && \
  go list -f \
  '{{with $$p := .}}{{if not $$p.Standard}}{{range $$f := $$p.GoFiles}}{{printf "%s/%s\n" $$p.Dir $$f }}{{end}}{{end}}{{end}}' \
  $$(go list -f '{{if not .Standard}}{{join .Deps "\n"}}{{end}}' \
  . \
  $$(go list -f '{{join .TestImports "\n"}}{{"\n"}}{{join .XTestImports "\n"}}') | sort -u))

# the test binary
$(PKG).test: $(SRCS)
	go test -cover -c -o $@ .

# the coverage file
$(PKG).test.out: $(PKG).test
	./$< -test.coverprofile $@

build: $(PKG).test

test: $(PKG).test.out

clean:
	rm -f $(PKG).test $(PKG).test.out

.PHONY: clean
