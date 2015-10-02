#!/bin/bash

REPO="${1:-stable}"
URL=https://dl.bintray.com/emccode/rexray/$REPO/latest
OS=$(uname -s)
ARCH=$(uname -m)
SUDO=$(which sudo)
BIN_DIR=/usr/bin
BIN_FILE=$BIN_DIR/rexray

function sudo() {
    if [[ $(id -u) -eq 0 ]]; then $@; else $SUDO $@; fi
}

function is_coreos() {
    grep -q DISTRIB_ID=CoreOS /etc/lsb-release &> /dev/null
    if [[ $? -eq 0 ]]; then echo 0; else echo 1; fi
}

IS_COREOS=$(is_coreos)

# how to detect the linux distro was taken from http://bit.ly/1JkNwWx
if [[ -e /etc/redhat-release || \
      -e /etc/redhat-version ]]; then
    
    #echo "installing rpm"
    sudo rpm -ih --quiet $URL/rexray-latest-$ARCH.rpm > /dev/null
    
elif [[ $ARCH = x86_64 && \
       (-e /etc/debian-release || \
        -e /etc/debian-version || \
        -e /etc/lsb-release) &&
        $IS_COREOS -eq 1 ]]; then

    #echo "installing deb"
    curl -sSLO $URL/rexray-latest-$ARCH.deb && \
        sudo dpkg -i rexray-latest-$ARCH.deb && \
        rm -f rexray-latest-$ARCH.deb

else 
    if [[ $IS_COREOS -eq 0 ]]; then
        BIN_DIR=/opt/bin
        BIN_FILE=$BIN_DIR/rexray
    elif [[ $OS = Darwin ]]; then
        BIN_DIR=/usr/local/bin
        BIN_FILE=$BIN_DIR/rexray
    fi

    sudo mkdir -p $BIN_DIR && \
      curl -sSLO $URL/rexray-$OS-$ARCH.tar.gz && \
      sudo tar xzf rexray-$OS-$ARCH.tar.gz -C $BIN_DIR && \
      rm -f rexray-$OS-$ARCH.tar.gz && \
      sudo chmod 0755 $BIN_FILE && \
      sudo chown 0 $BIN_FILE && \
      sudo chgrp 0 $BIN_FILE && \
      sudo $BIN_FILE install
fi

echo
echo "REX-Ray has been installed to $BIN_FILE"
echo
$BIN_FILE version
echo