ifneq (1,$(IS_GOMK_GOBINDATA_LOADED))

# note that the file is loaded
IS_GOMK_GOBINDATA_LOADED := 1

GOBINDATA := $(GOPATH)/bin/go-bindata

$(GOBINDATA):
	go get -u github.com/jteeuwen/go-bindata/...

endif
