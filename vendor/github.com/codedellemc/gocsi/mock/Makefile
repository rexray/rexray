all: build

PROG := mock

$(PROG): *.go ../*.go ../csi/csi.pb.go
	@go install .
	go build -o "$@" .

build: $(PROG)

clean:
	go clean -i -v .
	rm -f $(PROG)

clobber: clean

.PHONY: clean clobber
