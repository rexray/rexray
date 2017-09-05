all: build

################################################################################
##                                   DEP                                      ##
################################################################################
DEP := ./dep
DEP_VER ?= 0.3.0
DEP_ZIP := dep-$$GOHOSTOS-$$GOHOSTARCH.zip
DEP_URL := https://github.com/golang/dep/releases/download/v$(DEP_VER)/$$DEP_ZIP

$(DEP):
	GOVERSION=$$(go version | awk '{print $$4}') && \
	GOHOSTOS=$$(echo $$GOVERSION | awk -F/ '{print $$1}') && \
	GOHOSTARCH=$$(echo $$GOVERSION | awk -F/ '{print $$2}') && \
	DEP_ZIP="$(DEP_ZIP)" && \
	DEP_URL="$(DEP_URL)" && \
	mkdir -p .dep && \
	cd .dep && \
	curl -sSLO $$DEP_URL && \
	unzip "$$DEP_ZIP" && \
	mv $(@F) ../ && \
	cd ../ && \
	rm -fr .dep
ifneq (./dep,$(DEP))
dep: $(DEP)
endif

dep-ensure: | $(DEP)
	$(DEP) ensure -v


########################################################################
##                             PROTOC                                 ##
########################################################################

# Only set PROTOC_VER if it has an empty value.
ifeq (,$(strip $(PROTOC_VER)))
PROTOC_VER := 3.3.0
endif

PROTOC_OS := $(shell uname -s)
ifeq (Darwin,$(PROTOC_OS))
PROTOC_OS := osx
endif

PROTOC_ARCH := $(shell uname -m)
ifeq (i386,$(PROTOC_ARCH))
PROTOC_ARCH := x86_32
endif

PROTOC := ./protoc
PROTOC_ZIP := protoc-$(PROTOC_VER)-$(PROTOC_OS)-$(PROTOC_ARCH).zip
PROTOC_URL := https://github.com/google/protobuf/releases/download/v$(PROTOC_VER)/$(PROTOC_ZIP)
PROTOC_TMP_DIR := .protoc
PROTOC_TMP_BIN := $(PROTOC_TMP_DIR)/bin/protoc

$(PROTOC):
	-mkdir -p "$(PROTOC_TMP_DIR)" && \
	  curl -L $(PROTOC_URL) -o "$(PROTOC_TMP_DIR)/$(PROTOC_ZIP)" && \
	  unzip "$(PROTOC_TMP_DIR)/$(PROTOC_ZIP)" -d "$(PROTOC_TMP_DIR)" && \
	  chmod 0755 "$(PROTOC_TMP_BIN)" && \
	  cp -f "$(PROTOC_TMP_BIN)" "$@"
	-rm -fr "$(PROTOC_TMP_DIR)"
	stat "$@" > /dev/null 2>&1


########################################################################
##                          PROTOC-GEN-GO                             ##
########################################################################

# This is the recipe for getting and installing the go plug-in
# for protoc
PROTOC_GEN_GO_PKG := ./vendor/github.com/golang/protobuf/protoc-gen-go
PROTOC_GEN_GO := protoc-gen-go
$(PROTOC_GEN_GO):
	go build -o "$@" $(PROTOC_GEN_GO_PKG)

# Update PATH with the current directory. This enables the protoc
# binary to discover the protoc-gen-go binary, built inside this
# directory.
export PATH := $(shell pwd):$(PATH)


########################################################################
##                               CSI SPEC                             ##
########################################################################

# The paths of the CSI spec, protobuf, and Go source file.
CSI_SPEC :=  vendor/github.com/container-storage-interface/spec/spec.md
CSI_PROTO := csi/csi.proto
CSI_GOSRC := csi/csi.pb.go

# The temporary area and files used to build and compare updated
# file content.
CSI_TMP_DIR := csi/.build
CSI_PROTO_TMP := $(CSI_TMP_DIR)/csi.proto
CSI_GOSRC_TMP := $(CSI_TMP_DIR)/csi.pb.go

# This is the target for building the temporary CSI protobuf file.
#
# The temporary file is not versioned, and thus will always be
# built on Travis-CI.
$(CSI_PROTO_TMP): $(CSI_SPEC)
	@mkdir -p "$(@D)"
	sed -n -e '/```protobuf$$/,/```$$/ p' "$<" | \
	  sed -e 's@^```.*$$@////////@g' > "$@"
.PHONY: $(CSI_PROTO_TMP)


# If SKIP_CODEGEN is true then do not define a recipe for
# generating the protobuf.
ifneq (true,$(SKIP_CODEGEN))

# This is the target for building the CSI protobuf file.
#
# This target depends on its temp file, which is not versioned.
# Therefore when built on Travis-CI the temp file will always
# be built and trigger this target. On Travis-CI the temp file
# is compared with the real file, and if they differ the build
# will fail.
#
# Locally the temp file is simply copied over the real file.
$(CSI_PROTO): $(CSI_PROTO_TMP)
ifeq (true,$(TRAVIS))
	diff "$@" "$<"
else
	diff "$@" "$<" > /dev/null 2>&1 || cp -f "$<" "$@"
endif

endif

# This recipe generates the Go bindings to a temp area.
$(CSI_GOSRC_TMP): $(CSI_PROTO) | $(PROTOC) $(PROTOC_GEN_GO)
	@mkdir -p "$(@D)"
	$(PROTOC) -I "$(<D)" --go_out=plugins=grpc:"$(@D)" "$<"

# If SKIP_CODEGEN is true then do not define a recipe for
# generating the Go language binding.
ifneq (true,$(SKIP_CODEGEN))

# The temp language bindings are compared to the ones that are
# versioned. If they are different then it means the language
# bindings were not updated prior to being committed.
$(CSI_GOSRC): $(CSI_GOSRC_TMP)
ifeq (true,$(TRAVIS))
	diff "$@" "$<"
else
	diff "$@" "$<" > /dev/null 2>&1 || cp -f "$<" "$@"
endif

endif


########################################################################
##                     COMP W CSI SPEC MASTER                         ##
########################################################################

CSI_SPEC_MASTER := $(CSI_TMP_DIR)/csi.proto.master

CSI_SPEC_MASTER_URI := https://raw.githubusercontent.com/container-storage-interface/spec/master/spec.md
$(CSI_SPEC_MASTER):
	@mkdir -p "$(@D)"
	curl -sSL $(CSI_SPEC_MASTER_URI) | \
	  sed -n -e '/```protobuf$$/,/```$$/ p' | \
	  sed -e 's@^```.*$$@////////@g' > "$@"

# This recipe fetches the "spec.md" file from the "master" branch of the
# CSI spec repo and compares the generated protobuf with one generated
# from the aforementioned spec. If there is a difference between the
# two files this recipe will fail.
#
# Upon failure, aside from the output from the diff command, the last
# line will print the Git commit ID of the spec.md that was used
# in the comparison.
CSI_SPEC_MASTER_GIT := https://github.com/container-storage-interface/spec.git
csi-spec-comp-master: $(CSI_SPEC_MASTER) $(CSI_PROTO)
	-diff $^ || \
	git ls-remote -h $(CSI_SPEC_MASTER_GIT) | grep master | awk '{print $$1}'

.PHONY: csi-spec-comp-master $(CSI_SPEC_MASTER)


########################################################################
##                               GOCSI                                ##
########################################################################

GOCSI_A := gocsi.a
$(GOCSI_A): $(CSI_GOSRC) *.go
	@go install .
	go build -o "$@" .


########################################################################
##                               TEST                                 ##
########################################################################
GINKGO := ./ginkgo
GINKGO_PKG := ./vendor/github.com/onsi/ginkgo/ginkgo
$(GINKGO):
	go build -o "$@" $(GINKGO_PKG)

# The test recipe executes the Go tests with the Ginkgo test
# runner. This is the reason for the boolean OR condition
# that is part of the test script. The condition allows for
# the test run to exit with a status set to the value Ginkgo
# uses if it detects programmatic involvement. Please see
# https://goo.gl/CKz4La for more information.
ifneq (true,$(TRAVIS))
test: build
endif
test: | $(GINKGO)
	$(GINKGO) -v . || test "$$?" -eq "197"

.PHONY: test


########################################################################
##                               BUILD                                ##
########################################################################

build: $(GOCSI_A)
	$(MAKE) -C csc $@
	$(MAKE) -C mock $@

clean:
	go clean -i -v .
	rm -f "$(GOCSI_A)" "$(CSI_PROTO)" "$(CSI_GOSRC)"
	$(MAKE) -C csc $@
	$(MAKE) -C mock $@

clobber: clean
	rm -fr "$(CSI_TMP_DIR)"
	$(MAKE) -C csc $@
	$(MAKE) -C mock $@

.PHONY: build test clean clobber
