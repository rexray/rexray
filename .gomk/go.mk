ifneq (1,$(IS_GO_MK_LOADED))

# note that the file is loaded
IS_GO_MK_LOADED := 1

# enable go 1.5 vendoring
export GO15VENDOREXPERIMENT := 1

# include the main configuration file. this is where modules will put settings
# that should be readily apparent to users
include .gomk/config.mk

GO_BIN ?= $(GOPATH)/bin
GO_SRC := $(GOPATH)/src
GO_PKG := $(GOPATH)/pkg

# the path to the directory that contains the modular gomk makefiles
GOMK_I := .gomk/include

# text transformation funcs
LCASE = $(subst A,a,$(subst B,b,$(subst C,c,$(subst D,d,$(subst E,e,$(subst F,f,$(subst G,g,$(subst H,h,$(subst I,i,$(subst J,j,$(subst K,k,$(subst L,l,$(subst M,m,$(subst N,n,$(subst O,o,$(subst P,p,$(subst Q,q,$(subst R,r,$(subst S,s,$(subst T,t,$(subst U,u,$(subst V,v,$(subst W,w,$(subst X,x,$(subst Y,y,$(subst Z,z,$1))))))))))))))))))))))))))
UCASE = $(subst a,A,$(subst b,B,$(subst c,C,$(subst d,D,$(subst e,E,$(subst f,F,$(subst g,G,$(subst h,H,$(subst i,I,$(subst j,J,$(subst k,K,$(subst l,L,$(subst m,M,$(subst n,N,$(subst o,O,$(subst p,P,$(subst q,Q,$(subst r,R,$(subst s,S,$(subst t,T,$(subst u,U,$(subst v,V,$(subst w,W,$(subst x,X,$(subst y,Y,$(subst z,Z,$1))))))))))))))))))))))))))

################################################################################
##                             STD *NIX UTILS                                 ##
################################################################################
RM := rm
MKDIR := mkdir
TOUCH := touch
ENV := env
MV := mv
CP := cp
TAR := tar
CURL := curl
CAT := cat
GREP := grep
SED := sed
DATE := date

################################################################################
##                      PRE-BUILD RULES INCLUDES                              ##
################################################################################

# include the basic makefile pieces
include $(GOMK_I)/arch.mk \
		$(GOMK_I)/gnixutils.mk

################################################################################
##                               INDENT                                       ##
################################################################################

# the INDENT "function" enables indenting commands prior to them being
# echoed
ifeq (,$(INDENT_LEN))
INDENT_CHARS := ""
else
INDENT_CHARS := "$(foreach i,$(INDENT_LEN), )"
endif
ifeq (1,$(RECURSIVE))
INDENT_LEN := "$(strip . . $(INDENT_LEN))"
endif
INDENT := $(PRINTF) $(INDENT_CHARS)

################################################################################
##                               VENDOR                                       ##
################################################################################
GO_VENDOR_DIR := vendor
$(GO_VENDOR_DIR)-clean: | $(PRINTF) $(RM)
	@$(INDENT)
	$(RM) -f -d $(GO_VENDOR_DIR)
GO_CLOBBER += $(GO_VENDOR_DIR)-clean

# note that GO_CLEAN should be invoked as part of GO_CLOBBER
GO_CLOBBER += $(GO_CLEAN)

################################################################################
##                             GOMK TEMP DIRS                                 ##
################################################################################
GOMK_TMP_DIR := .gomk/tmp
GOMK_TAGGED_TMP_DIR := .gomk/tmp
ifneq (,$(GO_TAGS))
GOMK_TAGGED_TMP_DIR := $(GOMK_TMP_DIR)/$(subst $(SPACE),_,$(GO_TAGS))
endif

$(GOMK_TMP_DIR)-clobber: | $(PRINTF) $(RM)
	@$(INDENT)
	$(RM) -f -d $(GOMK_TMP_DIR)
GO_CLOBBER += $(GOMK_TAGGED_TMP_DIR)-clobber

# indicate where marker files are stored
GO_MARKERS_DIR := $(GOMK_TAGGED_TMP_DIR)/markers

# indicate where tests and test coverage output is stored
GO_TESTS_DIR := $(GOMK_TAGGED_TMP_DIR)/tests

# a temporary build area
GO_TMP_BUILD_DIR := $(GOMK_TAGGED_TMP_DIR)/build
GO_TMP_BUILD_BIN_DIR := $(GO_TMP_BUILD_DIR)/bin/$(GOOS)_$(GOARCH)
GO_TMP_BUILD_PKG_DIR := $(GO_TMP_BUILD_DIR)/pkg/$(GOOS)_$(GOARCH)

################################################################################
##                        POST-GOMK TEMP DIRS INCLUDES                        ##
################################################################################
include $(GOMK_I)/deps.mk \
		$(GOMK_I)/version.mk

# include all the tools used on the source files
include $(GOMK_I)/tools/source/*.mk

################################################################################
##                           PACKAGE INCLUDES                                 ##
################################################################################
include $(GOMK_I)/tools/pkg/*.mk

# pkg area
GO_TMP_PKG_DIR := $(GOMK_TMP_DIR)/pkg/$(V_OS)_$(V_ARCH)

################################################################################
##                                FUNCS                                       ##
################################################################################

# this function returns the path to a go source file's related marker file
GO_TOOL_MARKER = $(GO_MARKERS_DIR)/$(subst ./,,$(dir $(1))$(notdir $(1)).$(2))

# this function updates a marker file
GO_TOUCH_MARKER = $(MKDIR) -p $(dir $(1)) && $(TOUCH) $(1)

# recursively list the contents of a directory
RLSDIR = $(wildcard $1) $(foreach d,$(wildcard $1*),$(call RLSDIR,$d/))

################################################################################
##                            GO PROJ VARS                                    ##
################################################################################

# for examples of how the following varibles are initialized, assume the
# project is github.com/akutz/gomk

# gomk
GO_PROJ_NAME := $(notdir $(CURDIR))

# github.com/akutz/gomk
GO_PROJ_IMPORT_PATH := $(subst $(GO_SRC)/,,$(CURDIR))

# $GOPATH/pkg/$GOOS_$GOARCH/github.com/akutz/gomk
GO_PROJ_PKG_PATH := $(GO_PKG)/$(GOOS)_$(GOARCH)/$(GO_PROJ_IMPORT_PATH)

# $GOPATH/pkg/$GOOS_$GOARCH/github.com/akutz/
GO_PROJ_PKG_PARENT_PATH := $(dir $(GO_PROJ_PKG_PATH))

# if the GO_PKG_DIRS variable is empty then discover all sub-directories except
# the ones that should be specifically excluded
ifeq (,$(strip $(GO_PKG_DIRS)))
GO_PKG_DIRS := $(filter-out $(GO_PKG_DIRS_IGNORE_PATTS),$(sort $(dir $(call RLSDIR,./))))
GO_PKG_DIRS := $(addsuffix ...,$(GO_PKG_DIRS))
GO_PKG_DIRS := $(subst /...,,$(GO_PKG_DIRS))
endif

PATHS_MANAGED_BY_MAKEFILE :=

################################################################################
##                             GO BUILD / INSTALL                             ##
################################################################################

define GO_TOOL_DEF
ifeq (1,$$($2_ENABLED))
$$(foreach sf,$3,$$(eval $$(call $2,$$(sf),$1$4)))
GO_SRC_TOOL_MARKERS_$1$4 += $$($2_MARKER_PATHS_$1$4)
GO_SRC_TOOL_MARKERS += $$(GO_SRC_TOOL_MARKERS_$1$4)
endif
endef

define GO_PROJ_BUILD_ARCHIVE

ifeq ($1,.)
GO_PKG_NAME_$1 := $$(GO_PROJ_NAME)
GO_PKG_NAME_FULL_$1 := $$(GO_PROJ_NAME)
else
GO_PKG_NAME_$1 := $$(subst ./,,$1)
GO_PKG_NAME_FULL_$1 := $$(GO_PROJ_NAME)/$$(GO_PKG_NAME_$1)
endif

ifneq (,$$(wildcard $1/gomk.properties))
include $1/gomk.properties
endif

ifeq (,$$(GO_BUILD_OUTPUT_FILE_$1))
GO_PKG_ARCHIVE_PATH_$1 := $$(GO_PKG_NAME_$1).a
GO_PKG_ARCHIVE_PATH_FULL_$1 := $$(GO_PKG_NAME_FULL_$1).a
GO_BUILD_OUTPUT_FILE_$1 := $$(GO_TMP_BUILD_PKG_DIR)/$$(GO_PKG_ARCHIVE_PATH_FULL_$1)
endif

ifeq (,$$(GO_BUILD_FLAGS_$1))
GO_BUILD_FLAGS_$1 := $$(GO_BUILD_FLAGS)
endif
GO_BUILD_FLAGS_$1 := $$(call STRIP_GO_TAGS,$$(GO_BUILD_FLAGS_$1))

ifneq (,$$(GO_TAGS))
GO_BUILD_FLAGS_$1 += -tags '$$(GO_TAGS)'
endif

GO_PKG_ALL_SOURCE_FILES_$1 += $$(wildcard $1/*.go)
GO_PKG_ALL_SOURCE_FILES_$1 := $$(sort $$(GO_PKG_ALL_SOURCE_FILES_$1))
GO_PKG_ALL_SOURCE_FILES_$1 := $$(filter-out %_generated.go,$$(GO_PKG_ALL_SOURCE_FILES_$1))
GO_PKG_SOURCE_FILES_$1 := $$(filter-out %_test.go,$$(GO_PKG_ALL_SOURCE_FILES_$1))
GO_PKG_TEST_FILES_$1 := $$(filter %_test.go,$$(GO_PKG_ALL_SOURCE_FILES_$1))

ifneq (,$$(GO_PKG_ALL_SOURCE_FILES_$1))

# handle possible, duplicate target names
ifneq ($1,./$$(GO_PROJ_NAME))
GO_PKG_NAME_$1_TARGET_NAME := $$(GO_PKG_NAME_$1)
else
GO_PKG_NAME_$1_TARGET_NAME := $$(GO_PKG_NAME_$1)$$(GO_DUPLICATE_PKG_SUFFIX)
endif

# check to see if the build target is an archive, shared-object, or executable
# binary
ifeq (,$$(filter-out %.a,$$(GO_BUILD_OUTPUT_FILE_$1)))
	GO_PKG_IS_ARCHIVE_$1 := 1
else
	ifeq (,$$(filter-out %.so,$$(GO_BUILD_OUTPUT_FILE_$1)))
		GO_PKG_IS_SHARED_OBJ_$1 := 1
	else
		GO_PKG_IS_EXE_FILE_$1 := 1
	endif
endif

# indicate which tools should be executed against the source files
$$(foreach t,$$(GO_BUILD_DEP_RULES),$$(eval $$(call GO_TOOL_DEF,$1,$$(t),$$(GO_PKG_SOURCE_FILES_$1))))

$$(GO_PKG_SOURCE_FILES_$1):: $$(GO_GET_MARKERS) | $$(TOUCH)
	@$$(TOUCH) $$@

# go build
$$(GO_PKG_NAME_$1_TARGET_NAME): $$(GO_BUILD_OUTPUT_FILE_$1)
ifeq (0,$$(MAKELEVEL))
$$(GO_BUILD_OUTPUT_FILE_$1): $$(GO_SRC_TOOL_MARKERS_$1)
else
ifeq (1,$$(GOMK_TOOLS_ENABLE))
$$(GO_BUILD_OUTPUT_FILE_$1): $$(GO_SRC_TOOL_MARKERS_$1)
endif
endif
ifneq (,$$(GO_PKG_BUILD_DEPS_$1))
$$(GO_BUILD_OUTPUT_FILE_$1): $$(GO_PKG_BUILD_DEPS_$1)
endif
$$(GO_BUILD_OUTPUT_FILE_$1): $$(GO_PKG_SOURCE_FILES_$1) | $$(PRINTF) $$(MKDIR)
ifneq (1,$$(GO_PKG_IS_EXE_FILE_$1))
	@$$(INDENT)
	go build $$(GO_BUILD_FLAGS_$1) -o $$@ $1
else
ifeq (,$$(wildcard $$(dir $$(GO_BUILD_OUTPUT_FILE_$1))))
	@$$(MKDIR) -p $$(@D)
endif
	@$$(INDENT)
	go build $$(GO_BUILD_FLAGS_$1) -o $$@ $1
endif
GO_BUILD += $$(GO_PKG_NAME_$1_TARGET_NAME)

# go clean
$$(GO_PKG_NAME_$1_TARGET_NAME)-clean: | $$(PRINTF) $$(RM) $$(TOUCH)
	@$$(INDENT)
	go clean $(GO_CLEAN_FLAGS) $1
ifneq (,$$(strip $$(GO_BUILD_OUTPUT_FILE_$1)))
	@$$(INDENT)
	$$(RM) -f $$(GO_BUILD_OUTPUT_FILE_$1)
endif
GO_CLEAN += $$(GO_PKG_NAME_$1_TARGET_NAME)-clean

ifeq (1,$$(GO_PKG_TGZ_$1))
$$(eval $$(call GO_PKG_TGZ_RULE,$$(GO_BUILD_OUTPUT_FILE_$1),$$(GO_PKG_NAME_$1_TARGET_NAME)))
endif

# create the cross-install, cross-build, and cross-clean goals for this pkg
#$$(foreach bp,$$(BUILD_PLATFORMS),$$(eval $$(call CROSS_RULES,$$(bp),$$(GO_PKG_NAME_$1),$$(GO_PKG_NAME_$1_TARGET_NAME))))

################################################################################
##                               GO TEST                                      ##
################################################################################
ifneq (,$$(GO_PKG_TEST_FILES_$1))

# indicate which tools should be executed against the test source files
$$(foreach t,$$(GO_BUILD_DEP_RULES),$$(eval $$(call GO_TOOL_DEF,$1,$$(t),$$(GO_PKG_TEST_FILES_$1),_TEST)))

ifeq (.,$1)
GO_PKG_TEST_PATH_$1 := $$(GO_PKG_NAME_$1).test
GO_PKG_TEST_$1 := $$(GO_PKG_NAME_$1)-test
else
GO_PKG_TEST_PATH_$1 := $$(GO_PKG_NAME_$1)/$$(notdir $$(GO_PKG_NAME_$1)).test
GO_PKG_TEST_$1 := $$(GO_PKG_NAME_$1)/$$(notdir $$(GO_PKG_NAME_$1))-test
endif
GO_PKG_TEST_PATH_FULL_$1 := $$(GO_TESTS_DIR)/$$(GO_PKG_TEST_PATH_$1)

GO_PKG_COVER_PROFILE_$1 := $$(GO_PKG_TEST_PATH_FULL_$1).out

ifeq (,$$(findstring $$(GO_PKG_COVER_PROFILE_$1),$$(COVERAGE_EXCLUDE)))
GO_COVER_PROFILES += $$(GO_PKG_COVER_PROFILE_$1)
endif

ifneq (,$$(GO_TEST_COVER_PKGS_$1))
GO_TEST_COVER_PKGS_$1 := -coverpkg $$(GO_TEST_COVER_PKGS_$1)
endif

$$(GO_PKG_TEST_PATH_$1): $$(GO_PKG_TEST_PATH_FULL_$1)
$$(GO_PKG_TEST_PATH_FULL_$1):   $$(GO_SRC_TOOL_MARKERS_$1_TEST) \
								$$(GO_PKG_TEST_FILES_$1) \
								$$(GO_BUILD_OUTPUT_FILE_$1) \
								| $$(PRINTF)
	@$$(INDENT)
ifeq (,$$(GO_TAGS))
	go test $$(GO_TEST_BUILD_FLAGS) -cover -c $1 -o $$@ $$(GO_TEST_COVER_PKGS_$1)
else
	go test -tags '$$(GO_TAGS)' $$(GO_TEST_BUILD_FLAGS) -cover -c $1 -o $$@ $$(GO_TEST_COVER_PKGS_$1)
endif

$$(GO_PKG_TEST_$1): $$(GO_PKG_COVER_PROFILE_$1)
$$(GO_PKG_COVER_PROFILE_$1): $$(GO_PKG_TEST_PATH_FULL_$1) | $$(PRINTF)
	@$$(INDENT)
	$$? $$(GO_TEST_RUN_FLAGS) -test.coverprofile $$@

$$(GO_PKG_TEST_FILES_$1): $$(GO_PKG_SOURCE_FILES_$1) | $$(PRINTF) $$(TOUCH)
	@$$(TOUCH) $$@

$$(GO_PKG_TEST_PATH_$1)-clean: | $$(PRINTF) $$(RM)
	@$$(INDENT)
	$$(RM) -f $$(GO_PKG_TEST_PATH_FULL_$1) $$(GO_PKG_COVER_PROFILE_$1)
ifneq (,$$(strip $$(GO_SRC_TOOL_MARKERS_$1)))
	@$$(INDENT)
	$$(RM) -f $$(GO_SRC_TOOL_MARKERS_$1_TEST)
endif

GO_TEST_BUILD += $$(GO_PKG_TEST_PATH_$1)
GO_TEST += $$(GO_PKG_TEST_$1)
GO_TEST_CLEAN += $$(GO_PKG_TEST_PATH_$1)-clean
GO_CLEAN += $$(GO_PKG_TEST_PATH_$1)-clean

endif

endif
endef

define GO_PROJ_BUILD_RULES

# if the current path has a makefile in it, then we should note that
ifneq (.,$1)
ifneq (,$$(wildcard $1/Makefile))
PATHS_MANAGED_BY_MAKEFILE += $1%
endif
endif

# do not build the current path if it or any parent path has a Makefile present
ifeq ($1,$$(filter-out $$(PATHS_MANAGED_BY_MAKEFILE),$1))
$$(eval $$(call GO_PROJ_BUILD_ARCHIVE,$(1)))
endif

endef

# execute the build rules
$(foreach gpp,$(GO_PKG_DIRS),$(eval $(call GO_PROJ_BUILD_RULES,$(gpp))))

# only clean the targets in GO_CLEAN_ONCE at the initial make call
ifeq (0,$(MAKELEVEL))
GO_CLEAN += $(GO_CLEAN_ONCE)
endif

################################################################################
##                       POST-BUILD RULES INCLUDES                            ##
################################################################################

# include all the tools used during test processing
include $(GOMK_I)/tools/test/*.mk

################################################################################
##                         CROSS-BUILD INCLUDES                               ##
################################################################################
include $(GOMK_I)/cross.mk

endif
