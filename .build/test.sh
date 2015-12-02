#!/bin/bash

echo "mode: set" > acc.out
FAIL=0

ROOT_PKG="github.com/emccode/libstorage"
COVER_PKG="$ROOT_PKG"
COVER_PKG=$COVER_PKG,"$ROOT_PKG/api"
COVER_PKG=$COVER_PKG,"$ROOT_PKG/client"
COVER_PKG=$COVER_PKG,"$ROOT_PKG/context"
COVER_PKG=$COVER_PKG,"$ROOT_PKG/driver"
COVER_PKG=$COVER_PKG,"$ROOT_PKG/service"
COVER_PKG=$COVER_PKG,"$ROOT_PKG/service/server"
COVER_PKG=$COVER_PKG,"$ROOT_PKG/service/server/handlers"

go test -coverpkg=$COVER_PKG  -coverprofile=profile.out || FAIL=1
if [ -f profile.out ]; then
    cat profile.out | grep -v "mode: set" >> acc.out
    rm -f profile.out
fi

if [ "$FAIL" -ne 0 ]; then
    rm -f profile.out acc.out
    exit 1
fi

if [ -n "$COVERALLS" -a "$FAIL" -eq "0" ]; then
    goveralls -v -coverprofile=acc.out
fi

rm -f profile.out acc.out
exit $FAIL
