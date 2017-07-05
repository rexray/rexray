################################################################################
## csi.mk - include to generate csi protobuf & go sources                     ##
##                                                                            ##
## envars:                                                                    ##
##                                                                            ##
##   CSI_SPEC_FILE    the path to a local spec file used to generate          ##
##                    the protobuf                                            ##
##                                                                            ##
##                    if specified then the CSI_GIT_* and other               ##
##                    CSI_SPEC_* variables are ignored                        ##
##                                                                            ##
##   CSI_GIT_OWNER    the github user or organization that owns the           ##
##                    git repository that contains the csi spec file          ##
##                                                                            ##
##                    default: container-storage-interface                    ##
##                                                                            ##
##   CSI_GIT_REPO     the github repository that contains the csi             ##
##                    spec file                                               ##
##                                                                            ##
##                    default: spec                                           ##
##                                                                            ##
##   CSI_GIT_REF      the git ref to use when getting the csi spec file       ##
##                    can be a branch name, a tag, or a git commit id         ##
##                                                                            ##
##                    default: master                                         ##
##                                                                            ##
##   CSI_SPEC_NAME    the name of the csi spec markdown file                  ##
##                                                                            ##
##                    default: spec.md                                        ##
##                                                                            ##
##   CSI_SPEC_PATH    the remote path of the csi spec markdown file           ##
##                                                                            ##
##   CSI_PROTO_NAME   the name of the proto file to generate.                 ##
##                    the value should not include the file                   ##
##                    extension                                               ##
##                                                                            ##
##                    default: csi                                            ##
##                                                                            ##
##   CSI_PROTO_DIR    the path of the directory in which the protobuf         ##
##                    and go source files will be generated. if this          ##
##                    directory does not exist it will be created             ##
##                                                                            ##
##                    default: .                                              ##
##                                                                            ##
##   CSI_PROTO_ADD    a list of addition protobuf files used when             ##
##                    building the go source file                             ##
##                                                                            ##
##   CSI_IMPORT_PATH  the package of the generated go source                  ##
##                                                                            ##
##                    default: csi                                            ##
##                                                                            ##
##   CSI_BUILD_TAGS   the build tags to add to the top of the generated       ##
##                    source file                                             ##
##                                                                            ##
## targets:                                                                   ##
##                                                                            ##
##   $(CSI_PROTO_NAME).proto   the csi protobuf file generated from           ##
##                             the spec file                                  ##
##                                                                            ##
##   $(CSI_PROTO_NAME).pb.go   the go source file generated from the          ##
##                             protobuf file                                  ##
##                                                                            ##
################################################################################

# if the config vars are not already set then initialize them
# with default values. the make ?= notation is not used as it
# will eval the var every time it is accessed. there is no ?:=
# assignment, hence the strip check
ifndef CSI_PROTO_NAME
CSI_PROTO_NAME := csi
endif
ifndef CSI_PROTO_DIR
CSI_PROTO_DIR := .
endif
ifndef CSI_IMPORT_PATH
CSI_IMPORT_PATH := csi
endif

# only assign CSI_GIT_* and CSI_SPEC_NAME default values
# if CSI_SPEC_FILE is not set
ifndef CSI_SPEC_FILE

ifndef CSI_GIT_OWNER
CSI_GIT_OWNER := container-storage-interface
endif
ifndef CSI_GIT_REPO
CSI_GIT_REPO := spec
endif
ifndef CSI_GIT_REF
CSI_GIT_REF := master
endif
ifndef CSI_SPEC_NAME
CSI_SPEC_NAME := spec.md
endif

# the uri of the git repository that contains the spec file
CSI_GIT_URI := https://github.com/$(CSI_GIT_OWNER)/$(CSI_GIT_REPO)

# the uri of the spec file via github's raw content scheme
CSI_SPEC_URI := $(CSI_GIT_OWNER)/$(CSI_GIT_REPO)/$(CSI_GIT_REF)
CSI_SPEC_URI := $(CSI_SPEC_URI)/$(CSI_SPEC_PATH)/$(CSI_SPEC_NAME)
CSI_SPEC_URI := https://raw.githubusercontent.com/$(subst //,/,$(CSI_SPEC_URI))

endif # ifndef CSI_SPEC_FILE

# the names of the csi protobuf and go source files
CSI_PROTO := $(CSI_PROTO_DIR)/$(CSI_PROTO_NAME).proto
CSI_GOSRC := $(CSI_PROTO_DIR)/$(CSI_PROTO_NAME).pb.go

# the printf format string used to create the 
# protobuf file's header
CSI_PRINTF_FMT := // spec: %s\nsyntax = \"proto3\";\n
CSI_PRINTF_FMT := $(CSI_PRINTF_FMT)package %s;\n
CSI_PRINTF_FMT := $(CSI_PRINTF_FMT)option go_package = \"%s\";\n

ifdef CSI_SPEC_FILE # local spec file

# CSI_SPEC_REF indicates from which spec file the
# protobuf file was generated. defer to travis-ci
# for the spec ref info when available
CSI_SPEC_REF := $(abspath $(CSI_SPEC_FILE))
ifeq (true,$(TRAVIS))
ifdef TRAVIS_PULL_REQUEST_SLUG
CSI_SPEC_REF := $(TRAVIS_PULL_REQUEST_SLUG)
else
CSI_SPEC_REF := $(TRAVIS_REPO_SLUG)
endif
CSI_SPEC_REF := $(CSI_SPEC_REF):$(TRAVIS_COMMIT)
endif # ifeq (true,$(TRAVIS))

$(CSI_PROTO):
	@mkdir -p $(@D)
	printf "$(CSI_PRINTF_FMT)" \
		"$(CSI_SPEC_REF)" \
		"$(CSI_IMPORT_PATH)" \
		"$(CSI_IMPORT_PATH)" > $@ && \
	cat $(CSI_SPEC_FILE) | \
	sed -n -e '/```protobuf$$/,/```$$/ p' | \
	sed -e 's@^```.*$$@////////@g' \
	    -e 's@\(ProbeNodeError probe_node_error = 9;$$\)@\1\
    GetNodeIDError get_node_id_error = 10;@g' >> $@

else # remote spec file

$(CSI_PROTO):
	@mkdir -p $(@D)
	ref=$$(git ls-remote $(CSI_GIT_URI) $(CSI_GIT_REF) | awk '{print $$1}') && \
	ref="$${ref:-$(CSI_GIT_REF)}" && \
	printf "$(CSI_PRINTF_FMT)" \
	  "$(CSI_GIT_OWNER)/$(CSI_GIT_REPO):$${ref}" \
	  "$(CSI_IMPORT_PATH)" \
	  "$(CSI_IMPORT_PATH)" > $@ && \
	curl -sSL $(CSI_SPEC_URI) | \
	sed -n -e '/```protobuf$$/,/```$$/ p' | \
	sed -e 's@^```.*$$@////////@g' \
	    -e 's@\(ProbeNodeError probe_node_error = 9;$$\)@\1\
    GetNodeIDError get_node_id_error = 10;@g' >> $@

endif # ifdef CSI_SPEC_FILE

$(CSI_GOSRC): $(CSI_PROTO) $(CSI_PROTO_ADD)
	protoc $(foreach i,$(sort $(^D)),-I $i) --go_out=plugins=grpc:$(CSI_PROTO_DIR) $^
ifneq (,$(strip $(CSI_BUILD_TAGS)))
	sed -i '1s/^/$(CSI_BUILD_TAGS)\n/' $@
endif
