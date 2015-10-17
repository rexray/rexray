#!/bin/sh

echo "mode: set" > acc.out
FAIL=0

for DIR in $(find . -type d \
             -not -path '*/.*' \
             -not -path './.git*' \
             -not -path '*/_*' \
             -not -path './vendor/*'); do

    if ls $DIR/*.go &> /dev/null; then
        go test -coverprofile=profile.out $DIR || FAIL=1
        if [ -f profile.out ]; then
            cat profile.out | grep -v "mode: set" >> acc.out
            rm -f profile.out
        fi
    fi

done

if [ -n "$COVERALLS" -a "$FAIL" -eq 0 ]; then
    goveralls -v -coverprofile=acc.out $COVERALLS
fi

rm -f acc.out

exit $FAIL