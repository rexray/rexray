ifneq (1,$(IS_GOMK_COVERALLS_LOADED))

IS_GOMK_COVERALLS_LOADED := 1

ifeq (true,$(TRAVIS))
ifeq (1,$(COVERALLS_ENABLED))
ifneq (,$(GO_COVER_PROFILES))

GOCOV_BIN := $(GO_BIN)/gocov
$(GOCOV_BIN):
	go get -v github.com/axw/gocov/gocov

GOVERALLS_BIN := $(GO_BIN)/goveralls
$(GOVERALLS_BIN):
	go get -v github.com/mattn/goveralls

COVER_BIN := $(GO_BIN)/cover
$(COVER_BIN):
	go get -v golang.org/x/tools/cmd/cover

GO_COVER_DEPS := $(GOCOV_BIN) $(GOVERALLS_BIN) $(COVER_BIN)
GO_TEST_DEPS += $(GO_COVER_DEPS)
GO_DEPS += $(GO_COVER_DEPS)

COVERALLS_PROFILE := $(GO_TESTS_DIR)/coveralls.out
COVERALLS_MARKER := $(GO_MARKERS_DIR)/go.coveralls

$(COVERALLS_PROFILE): $(GO_COVER_PROFILES)
	echo "mode: set" > $@
	$(foreach f,$?,$(GREP) -v "mode: set" $(f) >> $@ &&) true

$(COVERALLS_PROFILE)-clean:
	$(RM) -f $(COVERALLS_PROFILE)

GO_COVER := $(COVERALLS_MARKER)
$(COVERALLS_MARKER): $(COVERALLS_PROFILE)
	goveralls -coverprofile=$?
	@$(call GO_TOUCH_MARKER,$@)

$(COVERALLS_MARKER)-clean:
	$(RM) -f $(COVERALLS_MARKER)

GO_COVER_CLEAN := $(COVERALLS_PROFILE)-clean $(COVERALLS_MARKER)-clean
GO_CLEAN += $(GO_COVER_CLEAN)

endif
endif
endif

endif
