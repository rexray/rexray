#!/bin/sh

dir=/tmp/rexray
src=/go/src
nme=build-rexray
img=golang:1.7.1
pkg=github.com/emccode/rexray
srcpkg=$src/$pkg
cmd="/sbin/init -D"

function b {
    mkdir -p $dir && cd $dir
    git clone https://$pkg .
    docker run -d -w $srcpkg -v $1:/go -v $dir:$srcpkg --name $nme $img $cmd
    docker exec -i $nme make
    sudo chown $2:$2 /var/lib/rexray/volumes/$1/data/bin/rexray
    sudo cp /var/lib/rexray/volumes/$1/data/bin/rexray .
    docker stop $nme && docker rm $nme
}

b builds $(whoami)

if [ ! -x "rexray" ]; then exit 1; fi
