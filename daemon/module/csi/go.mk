# set the gopath if it's not set and then make sure
# it is set to a single path token, the first token
# if there are multiple ones
ifndef GOPATH
GOPATH := $(shell go env | grep GOPATH | sed 's/GOPATH="\(.*\)"/\1/')
endif
GOPATH := $(word 1,$(subst :, ,$(GOPATH)))

# ensure GOOS, GOARCH, GOHOSTOS, & GOHOSTARCH are set
ifndef GOOS
GOOS := $(shell go env | grep GOOS | sed 's/GOOS="\(.*\)"/\1/')
endif
ifndef GOARCH
GOARCH := $(shell go env | grep GOARCH | sed 's/GOARCH="\(.*\)"/\1/')
endif
ifndef GOHOSTOS
GOHOSTOS := $(shell go env | grep GOHOSTOS | sed 's/GOHOSTOS="\(.*\)"/\1/')
endif
ifndef GOHOSTARCH
GOHOSTARCH := $(shell go env | grep GOHOSTARCH | sed 's/GOHOSTARCH="\(.*\)"/\1/')
endif

# the project's import path
IMPORT_PATH := $(shell go list)

# define a build dir as well as its bin and pkg directories
# for the targeted GOOS_GOARCH as well as the system's
# GOHOST_GOHOSTARCH combinations
ifndef BUILD_DIR
BUILD_DIR := $(GOPATH)
endif
BIN_DIR ?= $(BUILD_DIR)/bin
PKG_DIR ?= $(BUILD_DIR)/pkg
SRC_DIR ?= $(BUILD_DIR)/src
BIN_DIR_GO ?= $(BIN_DIR)/$(GOOS)_$(GOARCH)
BIN_DIR_GOHOST ?= $(BIN_DIR)/$(GOHOSTOS)_$(GOHOSTARCH)
PKG_DIR_GO ?= $(PKG_DIR)/$(GOOS)_$(GOARCH)
PKG_DIR_GOHOST ?= $(PKG_DIR)/$(GOHOSTOS)_$(GOHOSTARCH)
$(sort $(BIN_DIR_GO) $(BIN_DIR_GOHOST) $(PKG_DIR_GO) $(PKG_DIR_GOHOST)):
	mkdir -p $@
