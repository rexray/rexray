ifneq (1,$(IS_GOMK_TGZ_LOADED))

# note that the file is loaded
IS_GOMK_TGZ_LOADED := 1

ifeq (1,$(PKG_TGZ_ENABLED))

define GO_PKG_TGZ_RULE
GO_PKG_TGZ_NAME_$1 := $2-tgz
GO_PKG_TGZ_FILE_$1 := $$(GO_TMP_PKG_DIR)/$$(notdir $1)-$$(V_OS)-$$(V_ARCH)-$$(V_SEMVER)$$(PKG_TGZ_EXTENSION)

$$(GO_PKG_TGZ_NAME_$1): $$(GO_PKG_TGZ_FILE_$1)
$$(GO_PKG_TGZ_FILE_$1): $1 | $$(TAR)
	@$$(MKDIR) -p $$(@D)
	@$$(INDENT)
	$$(TAR) -C $$(<D) -c -z -f $$@ $$(<F)

$$(GO_PKG_TGZ_NAME_$1)-clean:
	@$$(INDENT)
	$$(RM) -f $$(GO_PKG_TGZ_FILE_$1)

GO_PACKAGE_TGZ_FILES += $$(GO_PKG_TGZ_FILE_$1)
GO_PACKAGE_FILES += $$(GO_PKG_TGZ_FILE_$1)
GO_PACKAGE += $$(GO_PKG_TGZ_NAME_$1)
GO_PACKAGE_CLEAN += $$(GO_PKG_TGZ_NAME_$1)-clean
endef

endif

endif
