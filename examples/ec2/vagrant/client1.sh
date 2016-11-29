#!/bin/sh

function b {
    docker run -d -v $1:/go --name bb busybox tail -f /dev/null
    sudo cp /var/lib/rexray/volumes/$1/data/bin/rexray .
    docker stop bb && docker rm bb
    sudo rexray volume rm --volumeid \
        $(rexray volume ls --volumename $1 -l error | \
            grep 'id:' | \
            awk '{print $2}')
}

b builds

if [ ! -x "rexray" ]; then exit 1; fi
