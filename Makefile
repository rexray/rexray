all: build

################################################################################
##                                 CONSTANTS                                  ##
################################################################################
EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
COMMA := ,
5S := $(SPACE)$(SPACE)$(SPACE)$(SPACE)$(SPACE)

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

GO_LIST_BUILD_INFO_CMD := go list -f '{{with $$ip:=.}}{{with $$ctx:=context}}{{printf "%s %s %s %s %s 0,%s" $$ip.ImportPath $$ip.Name $$ip.Dir $$ctx.GOOS $$ctx.GOARCH (join $$ctx.BuildTags ",")}}{{end}}{{end}}'
BUILD_INFO := $(shell $(GO_LIST_BUILD_INFO_CMD))
ROOT_IMPORT_PATH := $(word 1,$(BUILD_INFO))
ROOT_IMPORT_NAME := $(word 2,$(BUILD_INFO))
ROOT_DIR := $(word 3,$(BUILD_INFO))
GOOS ?= $(word 4,$(BUILD_INFO))
GOARCH ?= $(word 5,$(BUILD_INFO))
BUILD_TAGS := $(word 6,$(BUILD_INFO))
BUILD_TAGS := $(subst $(COMMA), ,$(BUILD_TAGS))
BUILD_TAGS := $(wordlist 2,$(words $(BUILD_TAGS)),$(BUILD_TAGS))
VENDORED := 0
ifneq (,$(strip $(findstring vendor,$(ROOT_IMPORT_PATH))))
VENDORED := 1
endif

################################################################################
##                              PROJECT DETAIL                                ##
################################################################################

GO_LIST_IMPORT_PATHS_INFO_CMD := go list -f '{{with $$ip:=.}}{{if $$ip.ImportPath | le "$(ROOT_IMPORT_PATH)"}}{{if $$ip.ImportPath | gt "$(ROOT_IMPORT_PATH)/vendor" }}{{printf "%s;%s;%s;%s;%v;0,%s,%s,%s,%s;0,%s;0,%s;0,%s" $$ip.ImportPath $$ip.Name $$ip.Dir $$ip.Target $$ip.Stale (join $$ip.GoFiles ",") (join $$ip.CgoFiles ",") (join $$ip.CFiles ",") (join $$ip.HFiles ",") (join $$ip.TestGoFiles ",") (join $$ip.Imports ",") (join $$ip.TestImports ",")}};{{end}}{{end}}{{end}}' ./...
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
SRCS += $$(SRCS_$1)

ifneq (,$$(strip $$(SRCS_$1)))
PKG_A_$1 := $$(TARGET_$1)
PKG_D_$1 := $$(DIR_$1)/$$(NAME_$1).d

ALL_PKGS += $$(PKG_A_$1)

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
ifneq (,$$(filter $$(GOPATH)/src/C/%,$$(EXT_DEPS_SRCS_$1)))
EXT_DEPS_SRCS_$1 := $$(filter-out $$(GOPATH)/src/C/%,$$(EXT_DEPS_SRCS_$1))
ifeq (main,$$(NAME_$1))
C_$1 := 1
endif
endif
EXT_DEPS_SRCS += $$(EXT_DEPS_SRCS_$1)

DEPS_ARKS_$1 := $$(addprefix $$(GOPATH)/pkg/$$(GOOS)_$$(GOARCH)/,$$(addsuffix .a,$$(INT_DEPS_$1)))
endif

TEST_SRCS_$1 := $$(subst $$(COMMA), ,$$(word 7,$$(IMPORT_PATH_INFO_$1)))
TEST_SRCS_$1 := $$(wordlist 2,$$(words $$(TEST_SRCS_$1)),$$(TEST_SRCS_$1))
TEST_SRCS_$1 := $$(addprefix $$(DIR_$1)/,$$(TEST_SRCS_$1))
TEST_SRCS += $$(TEST_SRCS_$1)

ifneq (,$$(strip $$(TEST_SRCS_$1)))
PKG_TA_$1 := $$(DIR_$1)/$$(NAME_$1).test
PKG_TD_$1 := $$(DIR_$1)/$$(NAME_$1).test.d
PKG_TC_$1 := $$(DIR_$1)/$$(NAME_$1).test.out

ALL_TESTS += $$(PKG_TA_$1)

-include $1/coverage.mk
TEST_COVERPKG_$1 ?= $$(IMPORT_PATH_$1)

TEST_DEPS_$1 := $$(subst $$(COMMA), ,$$(word 9,$$(IMPORT_PATH_INFO_$1)))
TEST_DEPS_$1 := $$(wordlist 2,$$(words $$(TEST_DEPS_$1)),$$(TEST_DEPS_$1))
TEST_DEPS_$1 := $$(filter-out $$(GO_STDLIB),$$(TEST_DEPS_$1))

TEST_INT_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)/vendor/%,$$(TEST_DEPS_$1))
TEST_INT_DEPS_$1 := $$(filter $$(ROOT_IMPORT_PATH)%,$$(TEST_INT_DEPS_$1))

TEST_EXT_VENDORED_DEPS_$1 := $$(filter $$(ROOT_IMPORT_PATH)/vendor/%,$$(TEST_DEPS_$1))
TEST_EXT_VENDORED_DEPS_$1 := $$(subst $$(ROOT_IMPORT_PATH)/vendor/,,$$(TEST_EXT_VENDORED_DEPS_$1))
TEST_EXT_VENDORED_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)%,$$(TEST_EXT_VENDORED_DEPS_$1))
TEST_EXT_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)%,$$(TEST_DEPS_$1))
TEST_EXT_DEPS_$1 := $$(filter-out $$(GOPATH)/src/C/%,$$(TEST_EXT_DEPS_$1))
TEST_EXT_DEPS_$1 += $$(TEST_EXT_VENDORED_DEPS_$1)
TEST_EXT_DEPS += $$(TEST_EXT_DEPS_$1)
TEST_EXT_DEPS_SRCS_$1 := $$(addprefix $$(GOPATH)/src/,$$(addsuffix /*.go,$$(TEST_EXT_DEPS_$1)))
ifneq (,$$(filter $$(GOPATH)/src/C/%,$$(TEST_EXT_DEPS_SRCS_$1)))
TEST_EXT_DEPS_SRCS_$1 := $$(filter-out $$(GOPATH)/src/C/%,$$(TEST_EXT_DEPS_SRCS_$1))
ifeq (main,$$(NAME_$1))
TEST_C_$1 := 1
endif
endif

TEST_EXT_DEPS_SRCS += $$(TEST_EXT_DEPS_SRCS_$1)

TEST_DEPS_ARKS_$1 := $$(addprefix $$(GOPATH)/pkg/$$(GOOS)_$$(GOARCH)/,$$(addsuffix .a,$$(TEST_INT_DEPS_$1)))
endif

ALL_SRCS_$1 += $$(SRCS_$1) $$(TEST_SRCS_$1)
ALL_SRCS += $$(ALL_SRCS_$1)

endef
$(foreach i,\
	$(IMPORT_PATH_INFO),\
	$(eval $(call IMPORT_PATH_PREPROCS_DEF,$(subst $(ROOT_DIR),.,$(word 3,$(subst ;, ,$(i)))),$(i))))

################################################################################
##                                  INFO                                      ##
################################################################################
info:
	$(info Project Import Path....$(ROOT_IMPORT_PATH))
	$(info Project Name...........$(ROOT_IMPORT_NAME))
	$(info OS / Arch..............$(GOOS)_$(GOARCH))
	$(info Vendored...............$(VENDORED))
ifneq (,$(strip $(SRCS)))
	$(info Sources................$(patsubst ./%,%,$(firstword $(SRCS))))
	$(foreach s,$(patsubst ./%,%,$(wordlist 2,$(words $(SRCS)),$(SRCS))),\
		$(info $(5S)$(5S)$(5S)$(5S)$(SPACE)$(SPACE)$(SPACE)$(s)))
endif
ifneq (,$(strip $(TEST_SRCS)))
	$(info Test Sources...........$(patsubst ./%,%,$(firstword $(TEST_SRCS))))
	$(foreach s,$(patsubst ./%,%,$(wordlist 2,$(words $(TEST_SRCS)),$(TEST_SRCS))),\
		$(info $(5S)$(5S)$(5S)$(5S)$(SPACE)$(SPACE)$(SPACE)$(s)))
endif
#ifneq (,$(strip $(EXT_DEPS_SRCS)))
#	$(info Dependency Sources.....$(patsubst ./%,%,$(firstword $(EXT_DEPS_SRCS))))
#	$(foreach s,$(patsubst ./%,%,$(wordlist 2,$(words $(EXT_DEPS_SRCS)),$(EXT_DEPS_SRCS))),\
#		$(info $(5S)$(5S)$(5S)$(5S)$(SPACE)$(SPACE)$(SPACE)$(s)))
#endif

################################################################################
##                               DEPENDENCIES                                 ##
################################################################################
GOGET_LOCK := goget.lock
GLIDE_LOCK := glide.lock
GLIDE_YAML := glide.yaml
GLIDE_LOCK_D := glide.lock.d

EXT_DEPS := $(sort $(EXT_DEPS))
EXT_DEPS_SRCS := $(sort $(EXT_DEPS_SRCS))
TEST_EXT_DEPS := $(sort $(TEST_EXT_DEPS))
TEST_EXT_DEPS_SRCS := $(sort $(TEST_EXT_DEPS_SRCS))
ALL_EXT_DEPS := $(sort $(EXT_DEPS) $(TEST_EXT_DEPS))
ALL_EXT_DEPS_SRCS := $(sort $(EXT_DEPS_SRCS) $(TEST_EXT_DEPS_SRCS))

ifneq (1,$(VENDORED))
ifneq (,$(wildcard $(GLIDE_YAML)))
$(ALL_EXT_DEPS_SRCS): $(GLIDE_LOCK_D)

$(GLIDE_LOCK_D): $(GLIDE_LOCK)
	glide up && touch $@

$(GLIDE_LOCK): $(GLIDE_YAML)
	touch $@

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
endif

################################################################################
##                               PROJECT BUILD                                ##
################################################################################

define IMPORT_PATH_BUILD_DEF

ifneq (,$$(strip $$(SRCS_$1)))
ifneq (1,$$(C_$1))

DEPS_SRCS_$1 := $$(foreach d,$$(INT_DEPS_$1),$$(SRCS_.$$(subst $$(ROOT_IMPORT_PATH),,$$(d))))

$$(PKG_D_$1): $$(filter-out %_generated.go,$$(SRCS_$1))
	$$(file >$$@,$$(PKG_A_$1) $$(PKG_D_$1): $$(filter-out %_generated.go,$$(DEPS_SRCS_$1)))

-include $$(PKG_D_$1)

$$(PKG_D_$1)-clean:
	rm -f $$(PKG_D_$1)
GO_CLEAN += $$(PKG_D_$1)-clean

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
endif

################################################################################
##                               PROJECT TESTS                                ##
################################################################################
ifneq (,$$(strip $$(TEST_SRCS_$1)))
ifneq (1,$$(TEST_C_$1))

TEST_DEPS_SRCS_$1 := $$(foreach d,$$(TEST_INT_DEPS_$1),$$(SRCS_.$$(subst $$(ROOT_IMPORT_PATH),,$$(d))))

$$(PKG_TD_$1): $$(filter-out %_generated.go,$$(TEST_SRCS_$1))
	$$(file >$$@,$$(PKG_TA_$1) $$(PKG_TD_$1): $$(filter-out %_generated.go,$$(TEST_DEPS_SRCS_$1)))

$$(PKG_TD_$1)-clean:
	rm -f $$(PKG_TD_$1)
GO_CLEAN += $$(PKG_TD_$1)-clean

-include $$(PKG_TD_$1)

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
	go test -cover -coverpkg '$$(TEST_COVERPKG_$1)' -c -o $$@ $1
$$(PKG_TA_$1)-clean:
	rm -f $$(PKG_TA_$1)
GO_PHONY += $$(PKG_TA_$1)-clean
GO_CLEAN += $$(PKG_TA_$1)-clean

$$(PKG_TC_$1): $$(PKG_TA_$1)
	$$(PKG_TA_$1) -test.coverprofile $$@ $$(GO_TEST_FLAGS)
TEST_PROFILES += $$(PKG_TC_$1)

$$(PKG_TC_$1)-clean:
	rm -f $$(PKG_TC_$1)
GO_PHONY += $$(PKG_TC_$1)-clean

GO_TEST += $$(PKG_TC_$1)
GO_BUILD_TESTS += $$(PKG_TA_$1)
GO_CLEAN += $$(PKG_TC_$1)-clean

endif
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

################################################################################
##                                 EXECUTORS                                  ##
################################################################################
EXECUTOR := $(shell go list -f '{{.Target}}' ./cli/lsx/lsx-$(GOOS))
EXECUTOR_LINUX := $(shell env GOOS=linux go list -f '{{.Target}}' ./cli/lsx/lsx-linux)
EXECUTOR_DARWIN := $(shell env GOOS=darwin go list -f '{{.Target}}' ./cli/lsx/lsx-darwin)
EXECUTOR_WINDOWS := $(shell env GOOS=windows go list -f '{{.Target}}' ./cli/lsx/lsx-windows)
build-executor-linux: $(EXECUTOR_LINUX)
build-executor-darwin: $(EXECUTOR_DARWIN)
build-executor-windows: $(EXECUTOR_WINDOWS)

EXECUTORS_GENERATED := ./api/server/executors/executors_generated.go
API_SERVER_EXECUTORS_A := $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ROOT_IMPORT_PATH)/api/server/executors.a

define EXECUTOR_RULES
LSX_EMBEDDED_$2 := ./api/server/executors/bin/$$(notdir $1)

ifneq ($2,$$(GOOS))
$1:
	env GOOS=$2 GOARCH=amd64 $$(MAKE) -j $$@
$1-clean:
	rm -f $1
GO_PHONY += $1-clean
GO_CLEAN += $1-clean
endif

$$(LSX_EMBEDDED_$2): $1
	@mkdir -p $$(@D) && cp -f $$? $$@

EXECUTORS_EMBEDDED += $$(LSX_EMBEDDED_$2)
endef

$(eval $(call EXECUTOR_RULES,$(EXECUTOR_LINUX),linux))
$(eval $(call EXECUTOR_RULES,$(EXECUTOR_DARWIN),darwin))
#$(eval $(call EXECUTOR_RULES,$(EXECUTOR_WINDOWS),windows))

$(EXECUTORS_GENERATED): $(EXECUTORS_EMBEDDED)
	$(GOPATH)/bin/go-bindata -md5checksum -pkg executors -prefix $(@D)/bin -o $@ $(@D)/bin/...

$(EXECUTORS_GENERATED)-clean:
	rm -fr $(dir $(EXECUTORS_GENERATED))/bin
GO_PHONY += $(EXECUTORS_GENERATED)-clean
GO_CLEAN += $(EXECUTORS_GENERATED)-clean

$(API_SERVER_EXECUTORS_A): $(EXECUTORS_GENERATED)

################################################################################
##                                    C                                       ##
################################################################################
CC := gcc -Wall -pedantic -std=c99
ifneq (,$(wildcard /usr/include))
CC += -I/usr/include
endif

################################################################################
##                               SEMAPHORE BINS                               ##
################################################################################
SEM_OPEN := ./cli/semaphores/open
SEM_WAIT := ./cli/semaphores/wait
SEM_SIGNAL := ./cli/semaphores/signal
SEM_UNLINK := ./cli/semaphores/unlink

$(SEM_OPEN): $(SEM_OPEN).c
	$(CC) $? -o $@ -lpthread
$(SEM_OPEN)-clean:
	rm -f $(SEM_OPEN)
GO_PHONY += $(SEM_OPEN)-clean
GO_CLEAN += $(SEM_OPEN)-clean

$(SEM_WAIT): $(SEM_WAIT).c
	$(CC) $? -o $@ -lpthread
$(SEM_WAIT)-clean:
	rm -f $(SEM_WAIT)
GO_PHONY += $(SEM_WAIT)-clean
GO_CLEAN += $(SEM_WAIT)-clean

$(SEM_SIGNAL): $(SEM_SIGNAL).c
	$(CC) $? -o $@ -lpthread
$(SEM_SIGNAL)-clean:
	rm -f $(SEM_SIGNAL)
GO_PHONY += $(SEM_SIGNAL)-clean
GO_CLEAN += $(SEM_SIGNAL)-clean

$(SEM_UNLINK): $(SEM_UNLINK).c
	$(CC) $? -o $@ -lpthread
$(SEM_UNLINK)-clean:
	rm -f $(SEM_UNLINK)
GO_PHONY += $(SEM_UNLINK)-clean
GO_CLEAN += $(SEM_UNLINK)-clean

sem-tools: $(SEM_OPEN) $(SEM_WAIT) $(SEM_SIGNAL) $(SEM_UNLINK)
sem-tools-clean: $(addsuffix -clean,$(SEM_OPEN) $(SEM_WAIT) $(SEM_SIGNAL) $(SEM_UNLINK))


################################################################################
##                                  C CLIENT                                  ##
################################################################################
C_LIBSTOR_DIR := ./c
C_LIBSTOR_C_DIR := $(C_LIBSTOR_DIR)/libstor-c
C_LIBSTOR_C_SO := $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ROOT_IMPORT_PATH)/c/libstor-c.so
C_LIBSTOR_C_BIN := $(GOPATH)/bin/libstor-c
C_LIBSTOR_C_BIN_SRC := $(C_LIBSTOR_DIR)/libstor-c.c
C_LIBSTOR_C_GO_DEPS :=	$(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ROOT_IMPORT_PATH)/api/types.a \
						$(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ROOT_IMPORT_PATH)/client.a \

libstor-c: $(C_LIBSTOR_C_SO) $(C_LIBSTOR_C_BIN)

$(C_LIBSTOR_C_SO):  $(EXT_DEPS_SRCS_./c/libstor-c) \
					$(SRCS_./c/libstor-c) | $(DEPS_ARKS_./c/libstor-c)
	go build -buildmode=c-shared -o $@ $(C_LIBSTOR_C_DIR)

$(C_LIBSTOR_C_SO)-clean:
	rm -f $(C_LIBSTOR_C_SO) $(basename $(C_LIBSTOR_C_SO).h)
GO_PHONY += $(C_LIBSTOR_C_SO)-clean
GO_CLEAN += $(C_LIBSTOR_C_SO)-clean

$(C_LIBSTOR_C_BIN):  $(C_LIBSTOR_C_BIN_SRC) \
				 	 $(C_LIBSTOR_C_SO) \
					 $(C_LIBSTOR_C_GO_DEPS)
	$(CC) -I$(abspath $(C_LIBSTOR_C_DIR)) \
          -I$(dir $(C_LIBSTOR_C_SO)) \
          -L$(dir $(C_LIBSTOR_C_SO)) \
          -o $@ \
          $(C_LIBSTOR_C_BIN_SRC) \
          -lstor-c

################################################################################
##                                  C SERVER                                  ##
################################################################################
C_LIBSTOR_S_DIR := $(C_LIBSTOR_DIR)/libstor-s
C_LIBSTOR_S_SO := $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ROOT_IMPORT_PATH)/c/libstor-s.so
C_LIBSTOR_S_BIN := $(GOPATH)/bin/libstor-s
C_LIBSTOR_S_BIN_SRC := $(C_LIBSTOR_DIR)/libstor-s.c
C_LIBSTOR_S_GO_DEPS := $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ROOT_IMPORT_PATH)/api/server.a

libstor-s: $(C_LIBSTOR_S_BIN) $(C_LIBSTOR_S_SO)

$(C_LIBSTOR_S_SO):  $(EXT_DEPS_SRCS_./c/libstor-s) \
					$(SRCS_./c/libstor-s) | $(DEPS_ARKS_./c/libstor-s)
	go build -buildmode=c-shared -o $@ $(C_LIBSTOR_S_DIR)

$(C_LIBSTOR_S_SO)-clean:
	rm -f $(C_LIBSTOR_S_SO) $(basename $(C_LIBSTOR_S_SO).h)
GO_PHONY += $(C_LIBSTOR_S_SO)-clean
GO_CLEAN += $(C_LIBSTOR_S_SO)-clean

$(C_LIBSTOR_S_BIN):  $(C_LIBSTOR_TYPES_H) \
					 $(C_LIBSTOR_S_BIN_SRC) \
					 $(C_LIBSTOR_S_SO) \
					 $(C_LIBSTOR_S_GO_DEPS)
	$(CC) -I$(abspath $(C_LIBSTOR_DIR)) \
          -I$(dir $(C_LIBSTOR_S_SO)) \
          -L$(dir $(C_LIBSTOR_S_SO)) \
          -o $@ \
          $(C_LIBSTOR_S_BIN_SRC) \
          -lstor-s

################################################################################
##                                  SERVERS                                   ##
################################################################################
LSS_BIN := $(shell go list -f '{{.Target}}' ./cli/lss/lss-$(GOOS))
LSS_ALL += $(LSS_BIN)
LSS_LINUX := $(shell env GOOS=linux go list -f '{{.Target}}' ./cli/lss/lss-linux)
LSS_DARWIN := $(shell env GOOS=darwin go list -f '{{.Target}}' ./cli/lss/lss-darwin)
LSS_WINDOWS := $(shell env GOOS=windows go list -f '{{.Target}}' ./cli/lss/lss-windows)
build-lss-linux: $(LSS_LINUX)
build-lss-darwin: $(LSS_DARWIN)
build-lss-windows: $(LSS_WINDOWS)

define LSS_RULES
ifneq ($2,$$(GOOS))
$1:
	env GOOS=$2 GOARCH=amd64 $$(MAKE) -j $$@
$1-clean:
	rm -f $1
GO_PHONY += $1-clean
GO_CLEAN += $1-clean
endif

LSS_ALL += $1
endef

#$(eval $(call LSS_RULES,$(LSS_LINUX),linux))
#$(eval $(call LSS_RULES,$(LSS_DARWIN),darwin))
#$(eval $(call LSS_RULES,$(LSS_WINDOWS),windows))

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

cover: $(COVERAGE)
ifneq (1,$(CODECOV_OFFLINE))
	curl -sSL https://codecov.io/bash | bash -s -- -f $?
else
	@echo codecov offline
endif

cover-debug:
	env LIBSTORAGE_DEBUG=true $(MAKE) cover


################################################################################
##                                  TARGETS                                   ##
################################################################################

build-tests: $(GO_BUILD_TESTS)

build-lsx: $(EXECUTORS_EMBEDDED)

build-lss: $(LSS_ALL)

build: sem-tools $(GO_BUILD)
	$(MAKE) -j libstor-c libstor-s
	$(MAKE) build-lss

test: $(GO_TEST)

test-debug:
	env LIBSTORAGE_DEBUG=true $(MAKE) test

clean: $(GO_CLEAN)

clobber: clean $(GO_CLOBBER)

run: $(GOPATH)/bin/libstorage-mock-server
	env LIBSTORAGE_RUN_HOST='tcp://127.0.0.1:7979' $?

run-debug:
	env LIBSTORAGE_DEBUG=true $(MAKE) run

run-tls:
	env LIBSTORAGE_RUN_TLS='true' $(MAKE) run

run-tls-debug:
	env LIBSTORAGE_RUN_TLS='true' $(MAKE) run-debug

.PHONY: info clean clobber \
		run run-debug run-tls run-tls-debug \
		$(GO_PHONY)
