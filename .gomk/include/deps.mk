ifneq (1,$(IS_GO_MK_DEPS_LOADED))

# note that the file is loaded
IS_GO_MK_DEPS_LOADED := 1

################################################################################
##                                 GLIDE                                      ##
################################################################################
ifeq (1,$(GLIDE_ENABLED))
ifneq (,$(wildcard glide.yaml))

ifeq (,$(strip $(GLIDE_HOME)))
ifeq (windows,$(SYS_GOOS))
export GLIDE_HOME=$(USERPROFILE)
else
export GLIDE_HOME=$(HOME)
endif
endif

GLIDE_BIN := $(GO_BIN)/glide
GLIDE_URL_BASE := https://github.com/Masterminds/glide/releases/download
GLIDE_URL_FILE := glide-$(GLIDE_VERSION)-$(SYS_GOOS)-$(SYS_GOARCH).tar.gz
GLIDE_URL := $(GLIDE_URL_BASE)/$(GLIDE_VERSION)/$(GLIDE_URL_FILE)

GLIDE_LOCK := glide.lock
GLIDE_YAML := glide.yaml

$(GLIDE_LOCK): $(GLIDE_YAML)
	$(GLIDE_BIN) up

$(GLIDE_LOCK)-clean:
	$(RM) -f $(GLIDE_LOCK)
GO_CLOBBER += $(GLIDE_LOCK)-clean

$(GLIDE_YAML): $(GLIDE_BIN)

$(GLIDE_BIN): | $(GNIX_UTILS)
	$(MKDIR) -p .glide && \
		cd .glide && \
		$(CURL) -S -L -o $(GLIDE_URL_FILE) $(GLIDE_URL) && \
		$(TAR) -x -v -z -f $(GLIDE_URL_FILE) && \
		$(MKDIR) -p $(GO_BIN) && \
		$(MV) $(SYS_GOOS)-$(SYS_GOARCH)/glide $(GLIDE_BIN) && \
		cd .. && \
		$(RM) -f -d .glide

GO_GET_MARKERS += $(GLIDE_LOCK)
GO_BUILD_DEPS += $(GLIDE_LOCK)
GO_DEPS += $(GLIDE_LOCK)

endif
endif

################################################################################
##                                 GO GET                                     ##
################################################################################
ifeq (1,$(GO_GET_ENABLED))
GO_GET := $(GO_MARKERS_DIR)/go.get

ifneq (,$(GLIDE_LOCK))
$(GO_GET): $(GLIDE_LOCK)
endif
$(GO_GET): | $(GNIX_UTILS)
	go get -v -d -t $(GO_PKG_DIRS)
	@$(call GO_TOUCH_MARKER,$@)

$(GO_GET)-clean:
	$(RM) -f $(GO_GET)
GO_CLOBBER += $(GO_GET)-clean

GO_GET_MARKERS += $(GO_GET)
GO_BUILD_DEPS += $(GO_GET)
GO_DEPS += $(GO_GET)

endif
endif
