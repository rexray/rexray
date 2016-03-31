all: build

################################################################################
##                                  CONFIG                                    ##
################################################################################

# a space-delimited, ordered list of drivers for which to build the libstorage
# server, client(s), and executor(s)
DRIVERS ?= mock vfs

################################################################################
##                                 CONSTANTS                                  ##
################################################################################
COMMA := ,

# a list of the go 1.6 stdlib pacakges as grepped from https://golang.org/pkg/
GO_STDLIB := archive archive/tar archive/zip bufio builtin bytes compress \
			 compress/bzip2 compress/flate compress/gzip compress/lzw \
			 compress/zlib container container/heap container/list \
			 container/ring crypto crypto/aes crypto/cipher crypto/des \
			 crypto/dsa crypto/ecdsa crypto/elliptic crypto/hmac crypto/md5 \
			 crypto/rand crypto/rc4 crypto/rsa crypto/sha1 crypto/sha256 \
			 crypto/sha512 crypto/subtle crypto/tls crypto/x509 \
			 crypto/x509/pkix database database/sql database/sql/driver debug \
			 debug/dwarf debug/elf debug/gosym debug/macho debug/pe \
			 debug/plan9obj encoding encoding/ascii85 encoding/asn1 \
			 encoding/base32 encoding/base64 encoding/binary encoding/csv \
			 encoding/gob encoding/hex encoding/json encoding/pem encoding/xml \
			 errors expvar flag fmt go go/ast go/build go/constant go/doc \
			 go/format go/importer go/parser go/printer go/scanner go/token \
			 go/types hash hash/adler32 hash/crc32 hash/crc64 hash/fnv html \
			 html/template image image/color image/color/palette image/draw \
			 image/gif image/jpeg image/png index index/suffixarray io \
			 io/ioutil log log/syslog math math/big math/cmplx math/rand mime \
			 mime/multipart mime/quotedprintable net net/http net/http/cgi \
			 net/http/cookiejar net/http/fcgi net/http/httptest \
			 net/http/httputil net/http/pprof net/mail net/rpc net/rpc/jsonrpc \
			 net/smtp net/textproto net/url os os/exec os/signal os/user path \
			 path/filepath reflect regexp regexp/syntax runtime runtime/cgo \
			 runtime/debug runtime/msan runtime/pprof runtime/race \
			 runtime/trace sort strconv strings sync sync/atomic syscall \
			 testing testing/iotest testing/quick text text/scanner \
			 text/tabwriter text/template text/template/parse time unicode \
			 unicode/utf16 unicode/utf8 unsafe

################################################################################
##                               PROJECT INFO                                 ##
################################################################################

GO_LIST_BUILD_INFO_CMD := go list -tags '$(GO_TAGS)' -f '{{with $$ip:=.}}{{with $$ctx:=context}}{{printf "%s %s %s %s %s 0,%s" $$ip.ImportPath $$ip.Name $$ip.Dir $$ctx.GOOS $$ctx.GOARCH (join $$ctx.BuildTags ",")}}{{end}}{{end}}'
BUILD_INFO := $(shell $(GO_LIST_BUILD_INFO_CMD))
ROOT_IMPORT_PATH := $(word 1,$(BUILD_INFO))
ROOT_IMPORT_NAME := $(word 2,$(BUILD_INFO))
ROOT_DIR := $(word 3,$(BUILD_INFO))
GOOS := $(word 4,$(BUILD_INFO))
GOARCH := $(word 5,$(BUILD_INFO))
BUILD_TAGS := $(word 6,$(BUILD_INFO))
BUILD_TAGS := $(subst $(COMMA), ,$(BUILD_TAGS))
BUILD_TAGS := $(wordlist 2,$(words $(BUILD_TAGS)),$(BUILD_TAGS))

################################################################################
##                              PROJECT DETAIL                                ##
################################################################################

GO_LIST_IMPORT_PATHS_INFO_CMD := go list -tags '$(GO_TAGS)' -f '{{with $$ip:=.}}{{if $$ip.ImportPath | le "github.com/emccode/libstorage"}}{{if $$ip.ImportPath | gt "github.com/emccode/libstorage/vendor" }}{{printf "%s;%s;%s;%s;%v;0,%s;0,%s;0,%s;0,%s" $$ip.ImportPath $$ip.Name $$ip.Dir $$ip.Target $$ip.Stale (join $$ip.GoFiles ",") (join $$ip.TestGoFiles ",") (join $$ip.Imports ",") (join $$ip.TestImports ",")}};{{end}}{{end}}{{end}}' ./...
IMPORT_PATH_INFO := $(shell $(GO_LIST_IMPORT_PATHS_INFO_CMD))

# this runtime ruleset acts as a pre-processor, processing the import path
# information completely before creating the build targets for the project
define IMPORT_PATH_PREPROCS_DEF

IMPORT_PATH_INFO_$1 := $$(subst ;, ,$2)

DIR_$1 := $1
IMPORT_PATH_$1 := $$(word 1,$$(IMPORT_PATH_INFO_$1))
NAME_$1 := $$(word 2,$$(IMPORT_PATH_INFO_$1))
TARGET_$1 := $$(word 4,$$(IMPORT_PATH_INFO_$1))
STALE_$1 := $$(word 5,$$(IMPORT_PATH_INFO_$1))

ifeq (1,$$(DEBUG))
$$(info name=$$(NAME_$1), target=$$(TARGET_$1), stale=$$(STALE_$1), dir=$$(DIR_$1))
endif

SRCS_$1 := $$(subst $$(COMMA), ,$$(word 6,$$(IMPORT_PATH_INFO_$1)))
SRCS_$1 := $$(wordlist 2,$$(words $$(SRCS_$1)),$$(SRCS_$1))
SRCS_$1 := $$(addprefix $$(DIR_$1)/,$$(SRCS_$1))
#$$(info SRCS_$1=$$(SRCS_$1))

DEPS_$1 := $$(subst $$(COMMA), ,$$(word 8,$$(IMPORT_PATH_INFO_$1)))
DEPS_$1 := $$(wordlist 2,$$(words $$(DEPS_$1)),$$(DEPS_$1))
DEPS_$1 := $$(filter-out $$(GO_STDLIB),$$(DEPS_$1))

INT_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)/vendor/%,$$(DEPS_$1))
INT_DEPS_$1 := $$(filter $$(ROOT_IMPORT_PATH)%,$$(INT_DEPS_$1))

EXT_VENDORED_DEPS_$1 := $$(filter $$(ROOT_IMPORT_PATH)/vendor/%,$$(DEPS_$1))
EXT_VENDORED_DEPS_$1 := $$(subst $$(ROOT_IMPORT_PATH)/vendor/,,$$(EXT_VENDORED_DEPS_$1))
EXT_VENDORED_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)%,$$(EXT_VENDORED_DEPS_$1))
EXT_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)%,$$(DEPS_$1))
EXT_DEPS_$1 += $$(EXT_VENDORED_DEPS_$1)
EXT_DEPS += $$(EXT_DEPS_$1)
EXT_DEPS_SRCS_$1 := $$(addprefix $$(GOPATH)/src/,$$(addsuffix /*.go,$$(EXT_DEPS_$1)))
EXT_DEPS_SRCS += $$(EXT_DEPS_SRCS_$1)

DEPS_ARKS_$1 := $$(addprefix $$(GOPATH)/pkg/$$(GOOS)_$$(GOARCH)/,$$(addsuffix .a,$$(INT_DEPS_$1)))

TEST_SRCS_$1 := $$(subst $$(COMMA), ,$$(word 7,$$(IMPORT_PATH_INFO_$1)))
TEST_SRCS_$1 := $$(wordlist 2,$$(words $$(TEST_SRCS_$1)),$$(TEST_SRCS_$1))
TEST_SRCS_$1 := $$(addprefix $$(DIR_$1)/,$$(TEST_SRCS_$1))

ifneq (,$$(strip $$(TEST_SRCS_$1)))
TEST_DEPS_$1 := $$(subst $$(COMMA), ,$$(word 9,$$(IMPORT_PATH_INFO_$1)))
TEST_DEPS_$1 := $$(wordlist 2,$$(words $$(TEST_DEPS_$1)),$$(TEST_DEPS_$1))
TEST_DEPS_$1 := $$(filter-out $$(GO_STDLIB),$$(TEST_DEPS_$1))

TEST_INT_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)/vendor/%,$$(TEST_DEPS_$1))
TEST_INT_DEPS_$1 := $$(filter $$(ROOT_IMPORT_PATH)%,$$(TEST_INT_DEPS_$1))

TEST_EXT_VENDORED_DEPS_$1 := $$(filter $$(ROOT_IMPORT_PATH)/vendor/%,$$(TEST_DEPS_$1))
TEST_EXT_VENDORED_DEPS_$1 := $$(subst $$(ROOT_IMPORT_PATH)/vendor/,,$$(TEST_EXT_VENDORED_DEPS_$1))
TEST_EXT_VENDORED_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)%,$$(TEST_EXT_VENDORED_DEPS_$1))
TEST_EXT_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)%,$$(TEST_DEPS_$1))
TEST_EXT_DEPS_$1 += $$(TEST_EXT_VENDORED_DEPS_$1)
TEST_EXT_DEPS += $$(TEST_EXT_DEPS_$1)
TEST_EXT_DEPS_SRCS_$1 := $$(addprefix $$(GOPATH)/src/,$$(addsuffix /*.go,$$(TEST_EXT_DEPS_$1)))
TEST_EXT_DEPS_SRCS += $$(TEST_EXT_DEPS_SRCS_$1)

TEST_DEPS_ARKS_$1 := $$(addprefix $$(GOPATH)/pkg/$$(GOOS)_$$(GOARCH)/,$$(addsuffix .a,$$(TEST_INT_DEPS_$1)))
endif

endef
$(foreach i,\
	$(IMPORT_PATH_INFO),\
	$(eval $(call IMPORT_PATH_PREPROCS_DEF,$(subst $(ROOT_DIR),.,$(word 3,$(subst ;, ,$(i)))),$(i))))

################################################################################
##                               DEPENDENCIES                                 ##
################################################################################
GOGET_LOCK := goget.lock
GLIDE_LOCK := glide.lock
GLIDE_YAML := glide.yaml

EXT_DEPS := $(sort $(EXT_DEPS))
EXT_DEPS_SRCS := $(sort $(EXT_DEPS_SRCS))
TEST_EXT_DEPS := $(sort $(TEST_EXT_DEPS))
TEST_EXT_DEPS_SRCS := $(sort $(TEST_EXT_DEPS_SRCS))
ALL_EXT_DEPS := $(sort $(EXT_DEPS) $(TEST_EXT_DEPS))
ALL_EXT_DEPS_SRCS := $(sort $(EXT_DEPS_SRCS) $(TEST_EXT_DEPS_SRCS))

ifneq (,$(GLIDE_YAML))
$(ALL_EXT_DEPS_SRCS): $(GLIDE_LOCK)

$(GLIDE_LOCK): $(GLIDE_YAML)
	glide up

$(GLIDE_LOCK)-clean:
	rm -f $(GLIDE_LOCK)
GO_PHONY += $(GLIDE_LOCK)-clean
GO_CLOBBER += $(GLIDE_LOCK)-clean
else
$(ALL_EXT_DEPS_SRCS): $(GOGET_LOCK)

$(GOGET_LOCK):
	go get -d $(ALL_EXT_DEPS) && touch $@

$(GOGET_LOCK)-clean:
	rm -f $(GOGET_LOCK)
GO_PHONY += $(GOGET_LOCK)-clean
GO_CLOBBER += $(GOGET_LOCK)-clean
endif

################################################################################
##                               PROJECT BUILD                                ##
################################################################################

define IMPORT_PATH_BUILD_DEF

ifneq (,$$(strip $$(SRCS_$1)))

PKG_A_$1 := $$(TARGET_$1)
PKG_D_$1 := $$(DIR_$1)/$$(NAME_$1).d

DEPS_SRCS_$1 := $$(foreach d,$$(INT_DEPS_$1),$$(SRCS_.$$(subst $$(ROOT_IMPORT_PATH),,$$(d))))

$$(PKG_D_$1): $$(SRCS_$1)
	$$(file >$$@,$$(PKG_A_$1) $$(PKG_D_$1): $$(DEPS_SRCS_$1))

-include $$(PKG_D_$1)

$$(PKG_D_$1)-clean:
	rm -f $$(PKG_D_$1)
GO_CLOBBER += $$(PKG_D_$1)-clean

$$(PKG_A_$1): $$(EXT_DEPS_SRCS_$1) $$(SRCS_$1) | $$(DEPS_ARKS_$1)
	go install $1

ifeq (true,$$(STALE_$1))
GO_PHONY += $$(PKG_A_$1)
endif

$$(PKG_A_$1)-clean:
	go clean -i -x $1 && rm -f $$(PKG_A_$1)

GO_BUILD += $$(PKG_A_$1)
GO_CLEAN += $$(PKG_A_$1)-clean

endif

################################################################################
##                               PROJECT TESTS                                ##
################################################################################
ifneq (,$$(strip $$(TEST_SRCS_$1)))

PKG_TA_$1 := $$(DIR_$1)/$$(NAME_$1).test
PKG_TC_$1 := $$(DIR_$1)/$$(NAME_$1).test.out
PKG_TD_$1 := $$(DIR_$1)/$$(NAME_$1).test.d

TEST_DEPS_SRCS_$1 := $$(foreach d,$$(TEST_INT_DEPS_$1),$$(SRCS_.$$(subst $$(ROOT_IMPORT_PATH),,$$(d))))

$$(PKG_TD_$1): $$(TEST_SRCS_$1)
	$$(file >$$@,$$(PKG_TA_$1) $$(PKG_TD_$1): $$(TEST_DEPS_SRCS_$1))

$$(PKG_TD_$1)-clean:
	rm -f $$(PKG_TD_$1)
GO_CLOBBER += $$(PKG_TD_$1)-clean

-include $$(PKG_TD_$1)

-include $1/coverage.mk
COVERPKG_$1 ?= $$(IMPORT_PATH_$1)

ifneq (,$$(strip $$(PKG_A_$1)))
$$(PKG_TA_$1): $$(PKG_A_$1)
ifeq (true,$$(STALE_$1))
GO_PHONY += $$(PKG_TA_$1)
endif
endif
ifneq (,$$(strip $$(SRCS_$1)))
$$(PKG_TA_$1): $$(SRCS_$1)
endif

$$(PKG_TA_$1): $$(TEST_SRCS_$1) $$(TEST_EXT_DEPS_SRCS_$1) | $$(TEST_DEPS_ARKS_$1)
	go test -cover -coverpkg '$$(COVERPKG_$1)' -c -o $$@ $1

$$(PKG_TC_$1): $$(PKG_TA_$1)
	$$(PKG_TA_$1) -test.v -test.coverprofile $$@
TEST_PROFILES += $$(PKG_TC_$1)

$$(PKG_TC_$1)-clean:
	rm -f $$(PKG_TC_$1)
GO_PHONY += $$(PKG_TC_$1)-clean

GO_TEST += $$(PKG_TC_$1)
GO_BUILD_TESTS += $$(PKG_TA_$1)
GO_CLEAN += $$(PKG_TC_$1)-clean

endif

endef
$(foreach i,\
	$(IMPORT_PATH_INFO),\
	$(eval $(call IMPORT_PATH_BUILD_DEF,$(subst $(ROOT_DIR),.,$(word 3,$(subst ;, ,$(i)))),$(i))))

################################################################################
##                                  SCHEMA                                    ##
################################################################################
LIBSTORAGE_JSON := libstorage.json
LIBSTORAGE_SCHEMA_GENERATED := api/utils/schema/schema_generated.go
LIBSTORAGE_SCHEMA_A := $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ROOT_IMPORT_PATH)/api/utils/schema.a
$(LIBSTORAGE_SCHEMA_A): $(LIBSTORAGE_SCHEMA_GENERATED)

$(LIBSTORAGE_SCHEMA_GENERATED): $(LIBSTORAGE_JSON)
	@echo generating $@
	@printf "package schema\n\nconst (\n" >$@; \
		printf "\t// JSONSchema is the libStorage API JSON schema\n" >>$@; \
		printf "\tJSONSchema = \`" >>$@; \
		sed -e 's/^//' $< >>$@; \
		printf "\`\n)\n" >>$@;

$(LIBSTORAGE_SCHEMA_GENERATED)-clean:
	rm -f $(LIBSTORAGE_SCHEMA_GENERATED)
GO_CLEAN += $(LIBSTORAGE_SCHEMA_GENERATED)-clean
GO_PHONY += $(LIBSTORAGE_SCHEMA_GENERATED)-clean

################################################################################
##                                 EXECUTORS                                  ##
################################################################################

EXECUTORS := $(foreach d,$(DRIVERS),$(GOPATH)/bin/libstorage-$(d)-executor)
$(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ROOT_IMPORT_PATH)/api/server.a: $(EXECUTORS)

################################################################################
##                                  COVERAGE                                  ##
################################################################################
COVERAGE := coverage.out
GO_COVERAGE := $(COVERAGE)
$(COVERAGE): $(TEST_PROFILES)
	printf "mode: set\n" > $@
	$(foreach f,$?,grep -v "mode: set" $(f) >> $@ &&) true

$(COVERAGE)-clean:
	rm -f $(COVERAGE)
GO_CLEAN += $(COVERAGE)-clean
GO_PHONY += $(COVERAGE)-clean

codecov: $(COVERAGE)
	curl -sSL https://codecov.io/bash | bash -s -- -f $?
GO_PHONY += codecov

################################################################################
##                                  TARGETS                                   ##
################################################################################

info:
	$(info Project Import Path  - $(ROOT_IMPORT_PATH))
	$(info Project Name         - $(ROOT_IMPORT_NAME))
	$(info GOOS_GOARCH          - $(GOOS)_$(GOARCH))

build-tests: $(GO_BUILD_TESTS)

build: $(GO_BUILD)

test: $(GO_TEST)

test-debug:
	env LIBSTORAGE_DEBUG=true $(MAKE) test

cover: $(GO_COVERAGE)

cover-debug:
	env LIBSTORAGE_DEBUG=true $(MAKE) cover

clean: $(GO_CLEAN)

clobber: clean $(GO_CLOBBER)

.PHONY: info build \
		clean clobber \
		build-tests test test-debug \
		cover cover-debug cover-clean \
		$(GO_PHONY)
