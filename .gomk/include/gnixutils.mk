ifneq (1,$(IS_GOMK_GNIXUTILS_LOADED))

# note that the file is loaded
IS_GOMK_GNIXUTILS_LOADED := 1

GNIX_SUFFIX := gnix
GNIX_CLI_SRC_DIR := $(GOPATH)/src/github.com/akutz/gnixutils/cli

GNIX_TOOLS := printf rm mkdir touch env mv cp tar curl cat grep sed date

define GNIX_TOOLS_DEF
GNIX_TOOL_BIN_$1 := $$(GO_BIN)/$1-$$(GNIX_SUFFIX)
GNIX_TOOL_SRC_$1 := $$(GNIX_CLI_SRC_DIR)/$1/$1.go
$$(call UCASE,$1) := $$(GNIX_TOOL_BIN_$1)

ifeq ($1,printf)
$$(GNIX_TOOL_SRC_$1):
	go get -v -d -u github.com/akutz/gnixutils
else
$$(GNIX_TOOL_SRC_$1): $$(GNIX_CLI_SRC_DIR)/printf/printf.go
endif

gnix-$1: $$(GNIX_TOOL_BIN_$1)
$$(GNIX_TOOL_BIN_$1): $$(GNIX_TOOL_SRC_$1)
	go build -o $$@ github.com/akutz/gnixutils/cli/$1
GO_DEPS += $$(GNIX_TOOL_BIN_$1)
GNIX_UTILS += $$(GNIX_TOOL_BIN_$1)

endef
$(foreach t,$(GNIX_TOOLS),$(eval $(call GNIX_TOOLS_DEF,$(t))))

ECHO := $(PRINTF)

endif
