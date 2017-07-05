SHELL := $(shell env which bash)

all: build

# load the go bits
include go.mk

# configure and load the csi protobuf generator
CSI_PROTO_DIR := csi
include csi.mk

# the name of the program being built
PROG := $(PKG_DIR_GO)/$(IMPORT_PATH).a

# the target for building the gocsi library
$(PROG):$(CSI_GOSRC) \
		$(filter-out %_test.go,$(wildcard *.go)) \
		| $(PKG_DIR_GO)
	go install -pkgdir $(abspath $(PKG_DIR_GO))
	@echo $@

build: $(PROG)

clean:
	go clean -i

clobber: clean
	rm -fr $(CSI_PROTO_DIR)

.PHONY: build clean clobber