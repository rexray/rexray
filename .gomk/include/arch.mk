ifneq (1,$(IS_GOMK_ARCH_LOADED))

# note that the file is loaded
IS_GOMK_ARCH_LOADED := 1

# set the go os and architecture types as well the sed command to use based on
# the os and architecture types
ifeq ($(OS),Windows_NT)
	SYS_GOOS := windows
	export GOOS ?= windows
	ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
		SYS_GOARCH := amd64
		export GOARCH ?= amd64
	endif
	ifeq ($(PROCESSOR_ARCHITECTURE),x86)
		SYS_GOARCH := 386
		export GOARCH ?= 386
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		SYS_GOOS := linux
		export GOOS ?= linux
	endif
	ifeq ($(UNAME_S),Darwin)
		SYS_GOOS := darwin
		export GOOS ?= darwin
		SYS_GOARCH := amd64
		export GOARCH ?= amd64
	endif

	UNAME_P := $(shell uname -p)
	ifeq (,$(SYS_GOARCH))
		ifeq ($(UNAME_P),x86_64)
			SYS_GOARCH := amd64
		endif
		ifneq ($(filter %86,$(UNAME_P)),)
			SYS_GOARCH := 386
		endif
	endif

	ifeq ($(origin GOARCH), undefined)
		ifeq ($(UNAME_P),x86_64)
			export GOARCH = amd64
		endif
		ifneq ($(filter %86,$(UNAME_P)),)
			export GOARCH = 386
		endif
	endif
endif
endif
