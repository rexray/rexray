#!/bin/bash

COVER_PROFILE=coverage.txt

echo "mode: set" > $COVER_PROFILE
FAIL=0

go test -cover ./rexray/cli || FAIL=1

if [ "$FAIL" -ne 0 ]; then
    exit 1
fi

#COVER_PKG="github.com/emccode/rexray","github.com/emccode/rexray/core"
#go test -coverpkg=$COVER_PKG  -coverprofile=profile.out ./test || FAIL=1
#if [ -f profile.out ]; then
#    cat profile.out | grep -v "mode: set" >> $COVER_PROFILE
#    rm -f profile.out
#fi

#if [ "$FAIL" -ne 0 ]; then
#    exit 1
#fi

if [ "$1" = "main" ]; then
    rm -f $COVER_PROFILE
    exit 0
fi

for DIR in $(find . -type d \
             -not -path '*/.*' \
             -not -path './.git*' \
             -not -path '*/_*' \
             -not -path './vendor/*' \
             -not -path './rexray/*' \
             -not -path './test/*' \
             -not -path './core' \
             -not -path '.'); do

    if ls $DIR/*.go &> /dev/null; then
        go test -coverprofile=profile.out $DIR || FAIL=1
        if [ -f profile.out ]; then
            cat profile.out | grep -v "mode: set" >> $COVER_PROFILE
            rm -f profile.out
        fi
    fi

done

if [ -n "$COVERALLS" -a "$FAIL" -eq "0" ]; then
    goveralls -v -coverprofile=$COVER_PROFILE
fi
if [ -n "$CODECOV" -a "$FAIL" -eq "0" ]; then
    bash <(curl -s https://codecov.io/bash)
fi

rm -f $COVER_PROFILE

exit $FAIL
