ifneq (1,$(IS_GOMK_GNIXUTILS_LOADED))

# note that the file is loaded
IS_GOMK_GNIXUTILS_LOADED := 1

GNIX_SUFFIX := gnix

RM := $(GO_BIN)/rm-$(GNIX_SUFFIX)
MKDIR := $(GO_BIN)/mkdir-$(GNIX_SUFFIX)
TOUCH := $(GO_BIN)/touch-$(GNIX_SUFFIX)
ENV := $(GO_BIN)/env-$(GNIX_SUFFIX)
MV := $(GO_BIN)/mv-$(GNIX_SUFFIX)
CP := $(GO_BIN)/cp-$(GNIX_SUFFIX)
TAR := $(GO_BIN)/tar-$(GNIX_SUFFIX)
CURL := $(GO_BIN)/curl-$(GNIX_SUFFIX)
CAT := $(GO_BIN)/cat-$(GNIX_SUFFIX)
GREP := $(GO_BIN)/grep-$(GNIX_SUFFIX)
SED := $(GO_BIN)/sed-$(GNIX_SUFFIX)
DATE := $(GO_BIN)/date-$(GNIX_SUFFIX)
PRINTF := $(GO_BIN)/printf-$(GNIX_SUFFIX)
ECHO := $(PRINTF)

GNIX_UTILS := $(GO_SRC)/github.com/akutz/gnixutils/LICENSE

$(GNIX_UTILS):
	go get -v -d -u github.com/akutz/gnixutils

$(RM): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/rm
GO_DEPS += $(RM)

$(MKDIR): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/mkdir
GO_DEPS += $(MKDIR)

$(TOUCH): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/touch
GO_DEPS += $(TOUCH)

$(ENV): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/env
GO_DEPS += $(ENV)

$(MV): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/mv
GO_DEPS += $(MV)

$(CP): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/cp
GO_DEPS += $(CP)

$(TAR): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/tar
GO_DEPS += $(TAR)

$(CURL): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/curl
GO_DEPS += $(CURL)

$(CAT): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/cat
GO_DEPS += $(CAT)

$(GREP): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/grep
GO_DEPS += $(GREP)

$(SED): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/sed
GO_DEPS += $(SED)

$(DATE): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/date
GO_DEPS += $(DATE)

$(PRINTF): | $(GNIX_UTILS)
	go build -o $@ github.com/akutz/gnixutils/cli/printf
GO_DEPS += $(PRINTF)

endif
