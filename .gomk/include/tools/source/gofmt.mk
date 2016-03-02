ifneq (1,$(IS_GOMK_GO_FMT_LOADED))

# note that the file is loaded
IS_GOMK_GO_FMT_LOADED := 1

ifeq (1,$(GO_FMT_ENABLED))

define GO_FMT
GO_FMT_MARKER_FILE_$1 := $$(call GO_TOOL_MARKER,$1,fmt)
GO_FMT_MARKER_PATHS_$2 += $$(GO_FMT_MARKER_FILE_$1)

$1-fmt: $$(GO_FMT_MARKER_FILE_$1)
$$(GO_FMT_MARKER_FILE_$1): $1
	@$$(INDENT)
	gofmt -w $$?
	@$$(call GO_TOUCH_MARKER,$$@)

$$(GO_FMT_MARKER_FILE_$1)-clean:
	@$$(INDENT)
	$$(RM) -f $$(subst -clean,,$$@)
GO_PHONY += $$(GO_FMT_MARKER_FILE_$1)-clean
GO_CLEAN_ONCE += $$(GO_FMT_MARKER_FILE_$1)-clean

endef

GO_BUILD_DEP_RULES += GO_FMT
endif

endif
