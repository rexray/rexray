all:
	$(MAKE) deps
	$(MAKE) build


################################################################################
##                                 CONSTANTS                                  ##
################################################################################
EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
ASTERIK := *
LPAREN := (
RPAREN := )
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
##                                SYSTEM INFO                                 ##
################################################################################

GOPATH := $(shell go env | grep GOPATH | sed 's/GOPATH="\(.*\)"/\1/')
GOHOSTOS := $(shell go env | grep GOHOSTOS | sed 's/GOHOSTOS="\(.*\)"/\1/')
GOHOSTARCH := $(shell go env | grep GOHOSTARCH | sed 's/GOHOSTARCH="\(.*\)"/\1/')
ifneq (,$(TRAVIS_GO_VERSION))
GOVERSION := $(TRAVIS_GO_VERSION)
else
GOVERSION := $(shell go version | awk '{print $$3}' | cut -c3-)
endif

ifeq (1.7,$(TRAVIS_GO_VERSION))
ifeq (linux,$(TRAVIS_OS_NAME))
COVERAGE_ENABLED := 1
endif
endif

# explicitly enable vendoring for Go 1.5.x versions.
GO15VENDOREXPERIMENT := 1

ifneq (,$(strip $(findstring 1.3.,$(TRAVIS_GO_VERSION))))
PRE_GO15 := 1
endif

ifneq (,$(strip $(findstring 1.4.,$(TRAVIS_GO_VERSION))))
PRE_GO15 := 1
endif

ifneq (1,$(PRE_GO15))
export GO15VENDOREXPERIMENT
endif


################################################################################
##                                  PATH                                      ##
################################################################################
export PATH := $(GOPATH)/bin:$(PATH)


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
##                               OS/ARCH INFO                                 ##
################################################################################
ifeq ($(GOOS),windows)
	OS ?= Windows_NT
endif
ifeq ($(GOOS),linux)
	OS ?= Linux
endif
ifeq ($(GOOS),darwin)
	OS ?= Darwin
endif
ifeq ($(GOARCH),386)
	ARCH ?= i386
endif
ifeq ($(GOARCH),amd64)
	ARCH ?= x86_64
endif

export OS
export ARCH


################################################################################
##                                MAKE FLAGS                                  ##
################################################################################
ifeq (,$(MAKEFLAGS))
MAKEFLAGS := --no-print-directory
export $(MAKEFLAGS)
endif


################################################################################
##                                 GAE SDK                                    ##
################################################################################
GAE_VER := 1.9.40
GAE_ZIP := go_appengine_sdk_$(GOHOSTOS)_$(GOHOSTARCH)-$(GAE_VER).zip
GAE_URL := https://storage.googleapis.com/appengine-sdks/featured/$(GAE_ZIP)
GAE_SDK := ./.gaesdk/$(GAE_VER)
GAE_SDZ := $(GAE_SDK)/$(GAE_ZIP)
GAE_DEV := $(GAE_SDK)/go_appengine/dev_appserver.py
GAE_GOA := $(GAE_SDK)/go_appengine/goapp
export PATH := $(PATH):$(dir $(GAE_DEV))

$(GAE_SDZ):
	mkdir -p $(GAE_SDK) && \
		cd $(GAE_SDK) && \
		curl -sSLO $(GAE_URL) && \
		cd -

gae-dev: $(GAE_DEV)
$(GAE_DEV): $(GAE_SDZ)
	cd $(?D) && unzip -q $(?F) && cd - && touch $@

GO_DEPS += $(GAE_DEV)
./gae/gae.test: | $(GAE_DEV) $(GAE_GOA)
./tests/gournal.test: | $(GAE_DEV) $(GAE_GOA)


################################################################################
##                              PROJECT DETAIL                                ##
################################################################################

GO_LIST_IMPORT_PATHS_INFO_CMD := go list -f '{{with $$ip:=.}}{{if $$ip.ImportPath | le "$(ROOT_IMPORT_PATH)"}}{{if or (lt $$ip.ImportPath "$(ROOT_IMPORT_PATH)/vendor") (ge $$ip.ImportPath "$(ROOT_IMPORT_PATH)/x")}}{{printf "%s;%s;%s;%s;%v;0,%s,%s,%s,%s;0,%s;0,%s;0,%s" $$ip.ImportPath $$ip.Name $$ip.Dir $$ip.Target $$ip.Stale (join $$ip.GoFiles ",") (join $$ip.CgoFiles ",") (join $$ip.CFiles ",") (join $$ip.HFiles ",") (join $$ip.TestGoFiles ",") (join $$ip.Imports ",") (join $$ip.TestImports ",")}};{{end}}{{end}}{{end}}' ./...
IMPORT_PATH_INFO := $(shell $(GO_LIST_IMPORT_PATHS_INFO_CMD))

# this runtime ruleset acts as a pre-processor, processing the import path
# information completely before creating the build targets for the project
define IMPORT_PATH_PREPROCS_DEF
ifeq (,$$(findstring examples,$1))

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
EXT_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)%,$$(DEPS_$1))
EXT_DEPS_$1 += $$(EXT_VENDORED_DEPS_$1)
EXT_DEPS += $$(EXT_DEPS_$1)
EXT_DEPS_SRCS_$1 := $$(addprefix $$(GOPATH)/src/,$$(addsuffix /*.go,$$(EXT_DEPS_$1)))
EXT_DEPS_SRCS_$1 := $$(subst $$(GOPATH)/src/$$(ROOT_IMPORT_PATH)/vendor/,./vendor/,$$(EXT_DEPS_SRCS_$1))
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

-include $1/gotest.mk
TEST_GO_$1 ?= go

TEST_DEPS_$1 := $$(subst $$(COMMA), ,$$(word 9,$$(IMPORT_PATH_INFO_$1)))
TEST_DEPS_$1 := $$(wordlist 2,$$(words $$(TEST_DEPS_$1)),$$(TEST_DEPS_$1))
TEST_DEPS_$1 := $$(filter-out $$(GO_STDLIB),$$(TEST_DEPS_$1))

TEST_INT_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)/vendor/%,$$(TEST_DEPS_$1))
TEST_INT_DEPS_$1 := $$(filter $$(ROOT_IMPORT_PATH)%,$$(TEST_INT_DEPS_$1))

TEST_EXT_VENDORED_DEPS_$1 := $$(filter $$(ROOT_IMPORT_PATH)/vendor/%,$$(TEST_DEPS_$1))
TEST_EXT_DEPS_$1 := $$(filter-out $$(ROOT_IMPORT_PATH)%,$$(TEST_DEPS_$1))
TEST_EXT_DEPS_$1 := $$(filter-out $$(GOPATH)/src/C/%,$$(TEST_EXT_DEPS_$1))
TEST_EXT_DEPS_$1 += $$(TEST_EXT_VENDORED_DEPS_$1)
TEST_EXT_DEPS += $$(TEST_EXT_DEPS_$1)
TEST_EXT_DEPS_SRCS_$1 := $$(addprefix $$(GOPATH)/src/,$$(addsuffix /*.go,$$(TEST_EXT_DEPS_$1)))
TEST_EXT_DEPS_SRCS_$1 := $$(subst $$(GOPATH)/src/$$(ROOT_IMPORT_PATH)/vendor/,./vendor/,$$(TEST_EXT_DEPS_SRCS_$1))
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

endif
endef
$(foreach i,\
	$(IMPORT_PATH_INFO),\
	$(eval $(call IMPORT_PATH_PREPROCS_DEF,$(subst $(ROOT_DIR),.,$(word 3,$(subst ;, ,$(i)))),$(i))))


################################################################################
##                                  INFO                                      ##
################################################################################
info:
	$(info Project Import Path.........$(ROOT_IMPORT_PATH))
	$(info Project Name................$(ROOT_IMPORT_NAME))
	$(info OS / Arch...................$(GOOS)_$(GOARCH))
	$(info Vendored....................$(VENDORED))
	$(info GOPATH......................$(GOPATH))
	$(info GOHOSTOS....................$(GOHOSTOS))
	$(info GOHOSTARCH..................$(GOHOSTARCH))
	$(info GOVERSION...................$(GOVERSION))
ifneq (,$(strip $(SRCS)))
	$(info Sources.....................$(patsubst ./%,%,$(firstword $(SRCS))))
	$(foreach s,$(patsubst ./%,%,$(wordlist 2,$(words $(SRCS)),$(SRCS))),\
		$(info $(5S)$(5S)$(5S)$(5S)$(5S)$(SPACE)$(SPACE)$(SPACE)$(s)))
endif
ifneq (,$(strip $(TEST_SRCS)))
	$(info Test Sources................$(patsubst ./%,%,$(firstword $(TEST_SRCS))))
	$(foreach s,$(patsubst ./%,%,$(wordlist 2,$(words $(TEST_SRCS)),$(TEST_SRCS))),\
		$(info $(5S)$(5S)$(5S)$(5S)$(5S)$(SPACE)$(SPACE)$(SPACE)$(s)))
endif
ifneq (,$(strip $(EXT_DEPS_SRCS)))
	$(info Dependency Sources..........$(patsubst ./%,%,$(firstword $(EXT_DEPS_SRCS))))
	$(foreach s,$(patsubst ./%,%,$(wordlist 2,$(words $(EXT_DEPS_SRCS)),$(EXT_DEPS_SRCS))),\
		$(info $(5S)$(5S)$(5S)$(5S)$(5S)$(SPACE)$(SPACE)$(SPACE)$(s)))
endif
ifneq (,$(strip $(TEST_EXT_DEPS_SRCS)))
	$(info Test Dependency Sources.....$(patsubst ./%,%,$(firstword $(TEST_EXT_DEPS_SRCS))))
	$(foreach s,$(patsubst ./%,%,$(wordlist 2,$(words $(TEST_EXT_DEPS_SRCS)),$(TEST_EXT_DEPS_SRCS))),\
		$(info $(5S)$(5S)$(5S)$(5S)$(5S)$(SPACE)$(SPACE)$(SPACE)$(s)))
endif


################################################################################
##                               DEPENDENCIES                                 ##
################################################################################
ifneq (1,$(PRE_GO15))

GLIDE := $(GOPATH)/bin/glide
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
$(GLIDE):
	go get -u github.com/Masterminds/glide
glide: $(GLIDE)
GO_DEPS += $(GLIDE)

GO_DEPS += $(GLIDE_LOCK_D)
$(ALL_EXT_DEPS_SRCS): $(GLIDE_LOCK_D)

ifeq (,$(strip $(wildcard $(GLIDE_LOCK))))
$(GLIDE_LOCK_D): $(GLIDE_LOCK) | $(GLIDE)
	touch $@

$(GLIDE_LOCK): $(GLIDE_YAML)
	$(GLIDE) up
else
$(GLIDE_LOCK_D): $(GLIDE_LOCK) | $(GLIDE)
	$(GLIDE) install && touch $@

$(GLIDE_LOCK): $(GLIDE_YAML)
	touch $@
endif

$(GLIDE_YAML):
	$(GLIDE) init

$(GLIDE_LOCK)-clean:
	rm -f $(GLIDE_LOCK)
GO_PHONY += $(GLIDE_LOCK)-clean
GO_CLOBBER += $(GLIDE_LOCK)-clean
endif

else
GOGET_D := .go.get.d

$(GOGET_D):
	go get -t -d ./... && touch $@
GO_DEPS += $(GOGET_D)

$(GOGET_D)-clean:
	rm -f $(GOGET_D)
GO_PHONY += $(GOGET_D)-clean

endif

################################################################################
##                               GOMETALINTER                                 ##
################################################################################
ifeq (1,$(PRE_GO15))
GOMETALINTER_DISABLED := 1
endif

ifneq (1,$(GOMETALINTER_DISABLED))
GOMETALINTER := $(GOPATH)/bin/gometalinter

$(GOMETALINTER): | $(GOMETALINTER_TOOLS)
	go get -u github.com/alecthomas/gometalinter
gometalinter: $(GOMETALINTER)
GO_DEPS += $(GOMETALINTER)

GOMETALINTER_TOOLS_D := .gometalinter.tools.d
$(GOMETALINTER_TOOLS_D): $(GOMETALINTER)
	$(GOMETALINTER) --install --update && touch $@
GO_DEPS += $(GOMETALINTER_TOOLS_D)

GOMETALINTER_ARGS := --vendor \
					 --fast \
					 --tests \
					 --cyclo-over=16 \
					 --deadline=30s \
					 --enable=gofmt \
					 --enable=goimports \
					 --enable=misspell \
					 --enable=lll \
					 --disable=gotype \
					 --severity=gofmt:error \
					 --severity=goimports:error \
					 --exclude=_generated.go \
					 --linter='gofmt:gofmt -l ./*.go:^(?P<path>[^\n]+)$''

gometalinter-warn: | $(GOMETALINTER_TOOLS_D) $(GLIDE)
	-$(GOMETALINTER) $(GOMETALINTER_ARGS) $(shell $(GLIDE) nv)

gometalinter-error: | $(GOMETALINTER_TOOLS_D) $(GLIDE)
	$(GOMETALINTER) $(GOMETALINTER_ARGS) --errors $(shell $(GLIDE) nv)

gometalinter-all:
ifeq (1,$(GOMETALINTER_WARN_ENABLED))
	$(MAKE) gometalinter-warn
endif
	$(MAKE) gometalinter-error
else
gometalinter-all:
	@echo gometalinter disabled
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
	GOOS=$(GOOS) GOARCH=$(GOARCH) go install $1

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
	$$(TEST_GO_$1) test -cover -coverpkg '$$(TEST_COVERPKG_$1)' -c -o $$@ $1
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
ifeq (1,$(COVERAGE_ENABLED))
	curl -sSL https://codecov.io/bash | bash -s -- -f $?
else
	@echo codecov enabled only for linux+go1.6.3
endif
else
	@echo codecov offline
endif

.coverage.tools.d:
ifeq (1,$(COVERAGE_ENABLED))
	go get github.com/onsi/gomega \
           github.com/onsi/ginkgo \
           golang.org/x/tools/cmd/cover && touch $@
else
	go get golang.org/x/tools/cmd/cover && touch $@
endif
GO_DEPS += .coverage.tools.d


################################################################################
##                                BENCHMARKS                                  ##
################################################################################
benchmark: ./benchmarks/benchmarks.test
	$? -test.run Benchmark -test.bench . -test.benchmem 2> /dev/null

# make the benchmark coverage file order-dependent upon the GAE coverage file.
# otherwise parallel make builds aren't happy when two GAE dev servers try to
# spin up at the same time
./benchmarks/benchmarks.test.out: | ./gae/gae.test.out

################################################################################
##                                 EXAMPLES                                   ##
################################################################################
run-example-1: build
	go run ./examples/01/main.go

run-example-2: build
	go run ./examples/02/main.go

run-example-3: build
	go run ./examples/03/main.go

run-example-4: build
	go run ./examples/04/main.go

run-examples:
	@echo EXAMPLE 1 && $(MAKE) run-example-1 && echo
	@echo EXAMPLE 2 && $(MAKE) run-example-2 && echo
	@echo EXAMPLE 3 && $(MAKE) run-example-3 && echo
	@echo EXAMPLE 4 && $(MAKE) run-example-4 && echo


################################################################################
##                                  TARGETS                                   ##
################################################################################
deps: $(GO_DEPS)

build-tests: $(GO_BUILD_TESTS)

build: $(GO_BUILD)

test: $(GO_TEST)

clean: $(GO_CLEAN)

clobber: clean $(GO_CLOBBER)

.PHONY: info clean clobber \
		$(GO_PHONY)
