ifneq (1,$(IS_GOMK_CROSS_LOADED))

# note that the file is loaded
IS_GOMK_CROSS_LOADED := 1

define CROSS_RULES
X_BP_$1 := $$(subst -, ,$1)
X_BP_OS_$1 := $$(firstword $$(X_BP_$1))
X_BP_ARCH_$1 := $$(lastword $$(X_BP_$1))

ifeq (Linux,$$(X_BP_OS_$1))
	X_GOOS_$1 := linux
else
	ifeq (Darwin,$$(X_BP_OS_$1))
		X_GOOS_$1 := darwin
	else
		ifeq (Windows,$$(X_BP_OS_$1))
			X_GOOS_$1 := windows
		endif
	endif
endif

ifeq (i386,$$(X_BP_ARCH_$1))
	X_GOARCH_$1 := 386
else
	ifeq (x86_64,$$(X_BP_ARCH_$1))
		X_GOARCH_$1 := amd64
	endif
endif

ifeq (1,$(RECURSIVE))
	RECURSIVE_INDENT := INDENT_LEN=$$(INDENT_LEN)
endif

ifeq (,$2)

build-$$(X_GOOS_$1)_$$(X_GOARCH_$1): $$(GO_DEPS) $$(GO_SRC_TOOL_MARKERS) | $$(ENV)
	$$(ENV) $$(RECURSIVE_INDENT) GOMK_TOOLS_ENABLE=0 GOOS=$$(X_GOOS_$1) GOARCH=$$(X_GOARCH_$1) $$(MAKE) build
GO_CROSS_BUILD += build-$$(X_GOOS_$1)_$$(X_GOARCH_$1)

#build-executors-$$(X_GOOS_$1)_$$(X_GOARCH_$1): $$(GO_DEPS) $$(GO_SRC_TOOL_MARKERS) | $$(ENV)
#	$$(ENV) $$(RECURSIVE_INDENT) GOMK_TOOLS_ENABLE=0 GOOS=$$(X_GOOS_$1) GOARCH=$$(X_GOARCH_$1) $$(MAKE) build-executors
#GO_CROSS_BUILD_EXECUTORS += build-executors-$$(X_GOOS_$1)_$$(X_GOARCH_$1)

dist-$$(X_GOOS_$1)_$$(X_GOARCH_$1): $$(GO_DEPS) $$(GO_SRC_TOOL_MARKERS) | $$(ENV)
	$$(ENV) $$(RECURSIVE_INDENT) GOMK_TOOLS_ENABLE=0 GOOS=$$(X_GOOS_$1) GOARCH=$$(X_GOARCH_$1) $$(MAKE) dist
GO_CROSS_DIST += dist-$$(X_GOOS_$1)_$$(X_GOARCH_$1)

clean-$$(X_GOOS_$1)_$$(X_GOARCH_$1): $$(GO_CLEAN_ONCE) | $$(ENV)
	$$(ENV) $$(RECURSIVE_INDENT) GOMK_TOOLS_ENABLE=0 GOOS=$$(X_GOOS_$1) GOARCH=$$(X_GOARCH_$1) $$(MAKE) clean
GO_CROSS_CLEAN += clean-$$(X_GOOS_$1)_$$(X_GOARCH_$1)

else

$3-build-$$(X_GOOS_$1)_$$(X_GOARCH_$1): $$(GO_DEPS) $$(GO_SRC_TOOL_MARKERS) | $$(ENV)
	$$(ENV) $$(RECURSIVE_INDENT) GOMK_TOOLS_ENABLE=0 GOOS=$$(X_GOOS_$1) GOARCH=$$(X_GOARCH_$1) $$(MAKE) $2 build
GO_CROSS_BUILD += $3-build-$$(X_GOOS_$1)_$$(X_GOARCH_$1)

$3-clean-$$(X_GOOS_$1)_$$(X_GOARCH_$1): $$(GO_CLEAN_ONCE) | $$(ENV)
	$$(ENV) $$(RECURSIVE_INDENT) GOMK_TOOLS_ENABLE=0 GOOS=$$(X_GOOS_$1) GOARCH=$$(X_GOARCH_$1) $$(MAKE) $2 clean
GO_CROSS_CLEAN += $3-clean-$$(X_GOOS_$1)_$$(X_GOARCH_$1)

endif

endef

$(foreach bp,$(BUILD_PLATFORMS),$(eval $(call CROSS_RULES,$(bp))))
#GO_PHONY += $(GO_CROSS_BUILD) $(GO_CROSS_BUILD_EXECUTORS) \

GO_PHONY += $(GO_CROSS_BUILD) \
			$(GO_CROSS_DIST) $(GO_CROSS_CLEAN)

build-all: $(GO_CROSS_BUILD)
#build-executors-all: $(GO_CROSS_BUILD_EXECUTORS)
dist-all: $(GO_CROSS_DIST)
clean-all: $(GO_CROSS_CLEAN)
#GO_PHONY += build-all build-executors-all dist-all clean-all
GO_PHONY += build-all dist-all clean-all

endif
