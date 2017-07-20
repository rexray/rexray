################################################################################
##                                GOLANG                                      ##
################################################################################

# define the go version to use
ifeq (,$(strip $(GO_VERSION)))
GO_VERSION := $(TRAVIS_GO_VERSION)
ifeq (,$(strip $(GO_VERSION)))
GO_VERSION := $(shell grep -A 1 '^go:' .travis.yml | tail -n 1 | awk '{print $$2}')
endif
endif

# set the gopath if it's not set and then make sure
# it is set to a single path token, the first token
# if there are multiple ones
ifeq (,$(strip $(GOPATH)))
GOPATH := $(shell go env | grep GOPATH | sed 's/GOPATH="\(.*\)"/\1/')
endif
GOPATH := $(word 1,$(subst :, ,$(GOPATH)))

# ensure GOOS, GOARCH, GOHOSTOS, & GOHOSTARCH are set
ifeq (,$(strip $(GOOS)))
GOOS := $(shell go env | grep GOOS | sed 's/GOOS="\(.*\)"/\1/')
endif
ifeq (,$(strip $(GOARCH)))
GOARCH := $(shell go env | grep GOARCH | sed 's/GOARCH="\(.*\)"/\1/')
endif
ifeq (,$(strip $(GOHOSTOS)))
GOHOSTOS := $(shell go env | grep GOHOSTOS | sed 's/GOHOSTOS="\(.*\)"/\1/')
endif
ifeq (,$(strip $(GOHOSTARCH)))
GOHOSTARCH := $(shell go env | grep GOHOSTARCH | sed 's/GOHOSTARCH="\(.*\)"/\1/')
endif

# two helpful concatenations
GOOS_GOARCH := $(GOOS)_$(GOARCH)
GOHOSTOS_GOHOSTARCH := $(GOHOSTOS)_$(GOHOSTARCH)

# export GOPATH, GOOS, and GOARCH so they're visible to shell commands
export GOPATH
export GOOS
export GOARCH

# ensure that an exported PATH includes GOPATH/bin
PATH := $(PATH):$(GOPATH)/bin
export PATH

# the project's import path
IMPORT_PATH := $(shell go list)
