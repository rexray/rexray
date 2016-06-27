#!/bin/bash
# gets akutz's go-bindata with md5checksum feature 

mkdir -p $GOPATH/src/github.com/jteeuwen/
cd $GOPATH/src/github.com/jteeuwen/
git clone https://github.com/akutz/go-bindata.git
cd $GOPATH/src/github.com/jteeuwen/go-bindata/
git checkout feature/md5checksum
cd $GOPATH/src/github.com/jteeuwen/go-bindata/go-bindata/
go build .
cp go-bindata $GOPATH/bin/
cp go-bindata $(which go | xargs dirname)
