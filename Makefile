SHELL := /bin/bash

all:
# if docker is running, then let's use docker to build it
ifneq (,$(shell if docker version &> /dev/null; then echo -; fi))
	$(MAKE) docker-build
else
	$(MAKE) deps
	$(MAKE) build
endif

################################################################################
##                                  DOCKER                                    ##
################################################################################
ifneq (,$(shell if docker version &> /dev/null; then echo -; fi))

DPKG := github.com/emccode/libstorage
DIMG := golang:1.7.1
DGOHOSTOS := $(shell uname -s | tr A-Z a-z)
ifeq (undefined,$(origin DGOOS))
DGOOS := $(DGOHOSTOS)
endif
DGOARCH := amd64
DPRFX := build-libstorage
DNAME := $(DPRFX)
ifeq (1,$(DBUILD_ONCE))
DNAME := $(DNAME)-$(shell date +%s)
endif
DPATH := /go/src/$(DPKG)
DSRCS := $(shell git ls-files)
ifneq (,$(DGLIDE_YAML))
DSRCS := $(filter-out glide.yaml,$(DSRCS))
DSRCS := $(filter-out glide.lock,$(DSRCS))
DSRCS := $(filter-out glide.lock.d,$(DSRCS))
endif
DPROG1_NAME := lss-$(DGOOS)
DPROG1_PATH := /go/bin/$(DPROG1_NAME)
DPROG2_NAME := lsx-$(DGOOS)
DPROG2_PATH := /go/bin/$(DPROG2_NAME)
ifneq (linux,$(DGOOS))
DPROG1_PATH := /go/bin/$(DGOOS)_$(DGOARCH)/$(DPROG1_NAME)
DPROG2_PATH := /go/bin/$(DGOOS)_$(DGOARCH)/$(DPROG2_NAME)
endif
ifeq (darwin,$(DGOHOSTOS))
DTARC := -
endif
DIMG_EXISTS := docker images --format '{{.Repository}}:{{.Tag}}' | grep $(DIMG) &> /dev/null
DTO_CLOBBER := docker ps -a --format '{{.Names}}' | grep $(DPRFX)
DNETRC := $(HOME)/.netrc

# DLOCAL_IMPORTS specifics a list of imported packages to copy into the
# container build's vendor directory instead of what is specified in the
# glide.lock file. If this variable is set and the GOPATH variable is not
# then the target will fail.
ifeq (undefined,$(DLOCAL_IMPORTS))
DLOCAL_IMPORTS :=
endif
ifneq (,$(DLOCAL_IMPORTS))
ifneq (,$(GOPATH))
DLOCAL_IMPORTS_FILES := $(foreach I,$(DLOCAL_IMPORTS),$(addprefix $I/,$(shell git --git-dir=$(GOPATH)/src/$(I)/.git --work-tree=$(GOPATH)/src/$(I) ls-files)))
DLOCAL_IMPORTS_FILES += $(foreach I,$(DLOCAL_IMPORTS),$I/.git)
endif
endif

docker-init:
	@if ! $(DIMG_EXISTS); then docker pull $(DIMG); fi
	@docker run --name $(DNAME) -d $(DIMG) /sbin/init -D &> /dev/null || true && \
		docker exec $(DNAME) mkdir -p $(DPATH) && \
		tar -c $(DTARC) .git $(DSRCS) | docker cp - $(DNAME):$(DPATH)
ifneq (,$(DGLIDE_YAML))
	@docker cp $(DGLIDE_YAML) $(DNAME):$(DPATH)/glide.yaml
endif
ifneq (,$(wildcard $(DNETRC)))
	@docker cp $(DNETRC) $(DNAME):/root
endif
	docker exec -t $(DNAME) env make -C $(DPATH) deps
ifneq (,$(DLOCAL_IMPORTS))
ifeq (,$(GOPATH))
	@echo GOPATH must be set when using DLOCAL_IMPORTS && false
else
	@docker exec -t $(DNAME) rm -fr $(addprefix $(DPATH)/vendor/,$(DLOCAL_IMPORTS))
	@tar -C $(GOPATH)/src -c $(DTARC) $(DLOCAL_IMPORTS_FILES) | docker cp - $(DNAME):$(DPATH)/vendor
endif
endif
	docker exec -t $(DNAME) env GOOS=$(DGOOS) GOARCH=$(DGOARCH) DOCKER=1 make -C $(DPATH) -j build

docker-build: docker-init
	@docker cp $(DNAME):$(DPROG1_PATH) $(DPROG1_NAME)
	@docker cp $(DNAME):$(DPROG2_PATH) $(DPROG2_NAME)
	@bytes=$$(stat --format '%s' $(PROG) 2> /dev/null || \
		stat -f '%z' $(DPROG1_NAME) 2> /dev/null) && mb=$$(($$bytes / 1024 / 1024)) && \
		printf "\nThe $(DPROG1_NAME) binary is $${mb}MB and located at: \n\n" && \
		printf "  ./$(DPROG1_NAME)\n\n"
	@bytes=$$(stat --format '%s' $(PROG) 2> /dev/null || \
		stat -f '%z' $(DPROG2_NAME) 2> /dev/null) && mb=$$(($$bytes / 1024 / 1024)) && \
		printf "\nThe $(DPROG2_NAME) binary is $${mb}MB and located at: \n\n" && \
		printf "  ./$(DPROG2_NAME)\n\n"
ifeq (1,$(DBUILD_ONCE))
	docker stop $(DNAME) &> /dev/null && docker rm $(DNAME) &> /dev/null
endif

docker-test: DGOOS=linux
docker-test: DTEST_ENV_VARS=TRAVIS=true LIBSTORAGE_DISABLE_STARTUP_INFO=true
docker-test: docker-init
	docker exec -t $(DNAME) env $(DTEST_ENV_VARS) make -C $(DPATH) test

docker-clean:
	-docker stop $(DNAME) &> /dev/null && docker rm $(DNAME) &> /dev/null

docker-clobber:
	-CNAMES=$$($(DTO_CLOBBER)); if [ "$$CNAMES" != "" ]; then \
		docker stop $$CNAMES && docker rm $$CNAMES; \
	fi

docker-list:
	-$(DTO_CLOBBER)

endif # ifneq (,$(shell if docker version &> /dev/null; then echo -; fi))


################################################################################
##                                 CONSTANTS                                  ##
################################################################################
ifneq (,$(shell which go 2> /dev/null)) # if go exists

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

ifeq (1.7.1,$(TRAVIS_GO_VERSION))
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
ROOT_IMPORT_PATH_NV := $(ROOT_IMPORT_PATH)
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
ROOT_IMPORT_PATH_NV := $(shell echo $(ROOT_IMPORT_PATH) | sed 's/.*vendor\///g')
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

endef
$(foreach i,\
	$(IMPORT_PATH_INFO),\
	$(eval $(call IMPORT_PATH_PREPROCS_DEF,$(subst $(ROOT_DIR),.,$(word 3,$(subst ;, ,$(i)))),$(i))))


################################################################################
##                                  INFO                                      ##
################################################################################
info:
	$(info Project Import Path.........$(ROOT_IMPORT_PATH))
ifeq (1,$(VENDORED))
	$(info No Vendor Import Path.......$(ROOT_IMPORT_PATH_NV))
endif
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
GLIDE := $(GOPATH)/bin/glide
GLIDE_VER := 0.11.1
GLIDE_TGZ := glide-v$(GLIDE_VER)-$(GOHOSTOS)-$(GOHOSTARCH).tar.gz
GLIDE_URL := https://github.com/Masterminds/glide/releases/download/v$(GLIDE_VER)/$(GLIDE_TGZ)
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
	@curl -SLO $(GLIDE_URL) && \
		tar xzf $(GLIDE_TGZ) && \
		rm -f $(GLIDE_TGZ) && \
		mkdir -p $(GOPATH)/bin && \
		mv $(GOHOSTOS)-$(GOHOSTARCH)/glide $(GOPATH)/bin && \
		rm -fr $(GOHOSTOS)-$(GOHOSTARCH)
glide: $(GLIDE)
GO_DEPS += $(GLIDE)

GO_DEPS += $(GLIDE_LOCK_D)
$(ALL_EXT_DEPS_SRCS): $(GLIDE_LOCK_D)

ifeq (,$(strip $(wildcard $(GLIDE_LOCK))))
$(GLIDE_LOCK_D): $(GLIDE_LOCK) | $(GLIDE)
	touch $@

$(GLIDE_LOCK): $(GLIDE_YAML)
	$(GLIDE) up

else #ifeq (,$(strip $(wildcard $(GLIDE_LOCK))))

$(GLIDE_LOCK_D): $(GLIDE_LOCK) | $(GLIDE)
	$(GLIDE) install && touch $@

$(GLIDE_LOCK): $(GLIDE_YAML)
	$(GLIDE) up && touch $@ && touch $(GLIDE_LOCK_D)

endif #ifeq (,$(strip $(wildcard $(GLIDE_LOCK))))

$(GLIDE_YAML):
	$(GLIDE) init

$(GLIDE_LOCK)-clean:
	rm -f $(GLIDE_LOCK)
GO_PHONY += $(GLIDE_LOCK)-clean
GO_CLOBBER += $(GLIDE_LOCK)-clean

endif #ifneq (1,$(VENDORED))


################################################################################
##                                GOBINDATA                                   ##
################################################################################
GO_BINDATA := $(GOPATH)/bin/go-bindata
go-bindata: $(GO_BINDATA)

GO_BINDATA_IMPORT_PATH := vendor/github.com/jteeuwen/go-bindata/go-bindata
ifneq (1,$(VENDORED))
GO_BINDATA_IMPORT_PATH := $(ROOT_IMPORT_PATH)/$(GO_BINDATA_IMPORT_PATH)
else
GO_BINDATA_IMPORT_PATH := $(firstword $(subst /vendor/, ,$(ROOT_IMPORT_PATH)))/$(GO_BINDATA_IMPORT_PATH)
endif

ifneq (1,$(VENDORED))
$(GO_BINDATA): $(GLIDE_LOCK_D)
endif
$(GO_BINDATA):
	GOOS=$(GOOS) GOARCH=$(GOARCH) go install $(GO_BINDATA_IMPORT_PATH)
	@touch $@
GO_DEPS += $(GO_BINDATA)


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
##                                  VERSION                                   ##
################################################################################

# figure out the git dirs
GIT_WORK:=.
GIT_ROOT:=.git
ifeq (1,$(VENDORED))
ifneq (,$(wildcard $(HOME)/.glide))
ROOT_IMPORT_PATH_DASH:=$(subst /,-,$(ROOT_IMPORT_PATH_NV))
VGIT_WORK:=$(shell find $(HOME)/.glide -name "*$(ROOT_IMPORT_PATH_DASH)" -type d)
ifneq (,$(wildcard $(VGIT_WORK)))
GIT_WORK:=$(VGIT_WORK)
ifneq (,$(wildcard $(VGIT_WORK)/.git))
GIT_ROOT:=$(VGIT_WORK)/.git
endif
endif
endif
endif

# parse a semver
SEMVER_PATT := ^[^\d]*(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z].+?))?(?:-(\d+)-g(.+?)(?:-(dirty))?)?$$
PARSE_SEMVER = $(shell echo $(1) | perl -pe 's/$(SEMVER_PATT)/$(2)/gim')

# describe the git information and create a parsing function for it
GIT_DESCRIBE := $(shell git --git-dir="$(GIT_ROOT)" --work-tree="$(GIT_WORK)" describe --long --dirty)
PARSE_GIT_DESCRIBE = $(call PARSE_SEMVER,$(GIT_DESCRIBE),$(1))

# parse the version components from the git information
V_MAJOR := $(call PARSE_GIT_DESCRIBE,$$1)
V_MINOR := $(call PARSE_GIT_DESCRIBE,$$2)
V_PATCH := $(call PARSE_GIT_DESCRIBE,$$3)
V_NOTES := $(call PARSE_GIT_DESCRIBE,$$4)
V_BUILD := $(call PARSE_GIT_DESCRIBE,$$5)
V_SHA_SHORT := $(call PARSE_GIT_DESCRIBE,$$6)
V_DIRTY := $(call PARSE_GIT_DESCRIBE,$$7)

V_OS := $(OS)
V_ARCH := $(ARCH)
V_OS_ARCH := $(V_OS)-$(V_ARCH)

# the long commit hash
V_SHA_LONG := $(shell git --git-dir="$(GIT_ROOT)" --work-tree="$(GIT_WORK)" show HEAD -s --format=%H)

# the branch name, possibly from travis-ci
ifeq ($(origin TRAVIS_BRANCH), undefined)
	TRAVIS_BRANCH := $(shell git --git-dir="$(GIT_ROOT)" --work-tree="$(GIT_WORK)" branch | grep '*')
else
ifeq (,$(strip $(TRAVIS_BRANCH)))
	TRAVIS_BRANCH := $(shell git --git-dir="$(GIT_ROOT)" --work-tree="$(GIT_WORK)" branch | grep '*')
endif
endif
TRAVIS_BRANCH := $(subst $(ASTERIK) ,,$(TRAVIS_BRANCH))
TRAVIS_BRANCH := $(subst $(LPAREN)HEAD detached at ,,$(TRAVIS_BRANCH))
TRAVIS_BRANCH := $(subst $(RPAREN),,$(TRAVIS_BRANCH))

ifeq ($(origin TRAVIS_TAG), undefined)
	TRAVIS_TAG := $(TRAVIS_BRANCH)
else
	ifeq ($(strip $(TRAVIS_TAG)),)
		TRAVIS_TAG := $(TRAVIS_BRANCH)
	endif
endif
V_BRANCH := $(TRAVIS_TAG)

# the build date as an epoch
V_EPOCH := $(shell date +%s)

# the build date
V_BUILD_DATE := $(shell perl -e 'use POSIX strftime; print strftime("%a, %d %b %Y %H:%M:%S %Z", localtime($(V_EPOCH)))')

# the release date as required by bintray
V_RELEASE_DATE := $(shell perl -e 'use POSIX strftime; print strftime("%Y-%m-%d", localtime($(V_EPOCH)))')

# init the semver
V_SEMVER := $(V_MAJOR).$(V_MINOR).$(V_PATCH)
ifneq ($(V_NOTES),)
	V_SEMVER := $(V_SEMVER)-$(V_NOTES)
endif

# get the version file's version
V_FILE := $(strip $(shell cat VERSION 2> /dev/null))

# append the build number and dirty values to the semver if appropriate
ifneq ($(V_BUILD),)
	ifneq ($(V_BUILD),0)
		# if the version file's version is different than the version parsed from the
		# git describe information then use the version file's version
		ifneq ($(V_SEMVER),$(V_FILE))
			V_MAJOR := $(call PARSE_SEMVER,$(V_FILE),$$1)
			V_MINOR := $(call PARSE_SEMVER,$(V_FILE),$$2)
			V_PATCH := $(call PARSE_SEMVER,$(V_FILE),$$3)
			V_NOTES := $(call PARSE_SEMVER,$(V_FILE),$$4)
			V_SEMVER := $(V_MAJOR).$(V_MINOR).$(V_PATCH)
			ifneq ($(V_NOTES),)
				V_SEMVER := $(V_SEMVER)-$(V_NOTES)
			endif
		endif
		V_SEMVER := $(V_SEMVER)+$(V_BUILD)
	endif
endif
ifeq ($(V_DIRTY),dirty)
	V_SEMVER := $(V_SEMVER)+$(V_DIRTY)
endif

define API_GENERATED_CONTENT
package api

import (
	"time"

	"github.com/emccode/libstorage/api/types"
)

func init() {
	Version = &types.VersionInfo{}
	Version.Arch = "$(V_OS_ARCH)"
	Version.Branch = "$(V_BRANCH)"
	Version.BuildTimestamp = time.Unix($(V_EPOCH), 0)
	Version.SemVer = "$(V_SEMVER)"
	Version.ShaLong = "$(V_SHA_LONG)"
}

endef
export API_GENERATED_CONTENT

PRINTF_VERSION_CMD += @printf "SemVer: %s\nBinary: %s\nBranch: %s\nCommit:
PRINTF_VERSION_CMD += %s\nFormed: %s\n" "$(V_SEMVER)" "$(V_OS_ARCH)"
PRINTF_VERSION_CMD += "$(V_BRANCH)" "$(V_SHA_LONG)" "$(V_BUILD_DATE)"
API_GENERATED_SRC := ./api/api_generated.go
$(API_GENERATED_SRC):
	echo generating $@
	@echo "$$API_GENERATED_CONTENT" > $@

$(API_GENERATED_SRC)-clean:
	rm -f $(API_GENERATED_SRC)
GO_CLEAN += $(API_GENERATED_SRC)-clean
GO_PHONE += $(API_GENERATED_SRC)-clean

API_A := $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ROOT_IMPORT_PATH)/api.a
$(API_A): $(API_GENERATED_SRC)

version:
	$(PRINTF_VERSION_CMD)

GO_PHONY += version


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
	env GOOS=$2 GOARCH=amd64 $$(MAKE) $$@
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
	$(GO_BINDATA) -md5checksum -pkg executors -prefix $(@D)/bin -o $@ $(@D)/bin/...

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
	env GOOS=$2 GOARCH=amd64 $$(MAKE) $$@
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

cover-debug:
	env LIBSTORAGE_DEBUG=true $(MAKE) cover


################################################################################
##                                  TARGETS                                   ##
################################################################################
deps: $(GO_DEPS)

build-tests: $(GO_BUILD_TESTS)

build-lsx: $(EXECUTORS_EMBEDDED)

build-lss: $(LSS_ALL)

build-libstorage: $(GO_BUILD)

build-generated:
	$(MAKE) build-lsx
	$(MAKE) $(API_GENERATED_SRC)

build:
	$(MAKE) build-generated
	$(MAKE) build-libstorage
ifeq ($(GOOS),$(GOHOSTOS))
	$(MAKE) libstor-c libstor-s
endif
	$(MAKE) build-lss

parallel-test: $(filter-out ./drivers/storage/vfs/%,$(GO_TEST))
vfs-test: $(filter ./drivers/storage/vfs/%,$(GO_TEST))
test:
	$(MAKE) vfs-test
	$(MAKE) -j parallel-test

test-debug:
	env LIBSTORAGE_DEBUG=true $(MAKE) test

clean: $(GO_CLEAN)

clobber: clean $(GO_CLOBBER)

.PHONY: info clean clobber $(GO_PHONY)

endif # ifneq (,$(shell which go 2> /dev/null))
