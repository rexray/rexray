ifneq (1,$(IS_GOMK_CODECOV_LOADED))

IS_GOMK_CODECOV_LOADED := 1

ifeq (true,$(TRAVIS))
ifeq (1,$(CODECOV_ENABLED))
ifneq (,$(GO_COVER_PROFILES))

GOMEGA_SRC := $(GOPATH)/src/github.com/onsi/gomega/README.md
$(GOMEGA_SRC):
	go get -v github.com/onsi/gomega

GINKGO_SRC := $(GOPATH)/src/github.com/onsi/ginkgo/README.md
$(GINKGO_SRC):
	go get -v github.com/onsi/ginkgo

COVER_BIN := $(GO_BIN)/cover
$(COVER_BIN):
	go get -v golang.org/x/tools/cmd/cover

GO_COVER_DEPS := $(GOMEGA_SRC) $(GINKGO_SRC) $(COVER_BIN)
GO_TEST_DEPS += $(GO_COVER_DEPS)
GO_DEPS += $(GO_COVER_DEPS)

CODECOV_PROFILE := $(GO_TESTS_DIR)/codecov.out
CODECOV_MARKER := $(GO_MARKERS_DIR)/go.codecov

$(CODECOV_PROFILE): $(GO_COVER_PROFILES)
	echo "mode: set" > $@
	$(foreach f,$?,$(GREP) -v "mode: set" $(f) >> $@ &&) true

$(CODECOV_PROFILE)-clean:
	$(RM) -f $(CODECOV_PROFILE)

GO_COVER := $(CODECOV_MARKER)
$(CODECOV_MARKER): $(CODECOV_PROFILE)
	curl -sSL https://codecov.io/bash | bash -s -- -f $?
	@$(call GO_TOUCH_MARKER,$@)

$(CODECOV_MARKER)-clean:
	$(RM) -f $(CODECOV_MARKER)

GO_COVER_CLEAN := $(CODECOV_PROFILE)-clean $(CODECOV_MARKER)-clean
GO_CLEAN += $(GO_COVER_CLEAN)

endif
endif
endif

endif
