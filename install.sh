#!/bin/bash

REPO="${1:-staged}"
URL=https://dl.bintray.com/emccode/rexray/$REPO/latest
ARCH=$(uname -m)

# how to detect the linux distro was taken from http://bit.ly/1JkNwWx
if [ -e "/etc/redhat-release" -o \
     -e "/etc/redhat-version" ]; then
    
    #echo "installing rpm"
    sudo rpm -ih --quiet $URL/rexray-latest-$ARCH.rpm > /dev/null
    
elif [ "$ARCH" = "x86_64" -a \
       -e "/etc/debian-release" -o \
       -e "/etc/debian-version" -o \
       -e "/etc/lsb-release" ]; then

    #echo "installing deb"
    curl -sSLO $URL/rexray-latest-$ARCH.deb && \
        sudo dpkg -i rexray-latest-$ARCH.deb && \
        rm -f rexray-latest-$ARCH.deb

else

    #echo "installing binary"
    curl -sSLO $URL/rexray-$(uname -s)-$ARCH.tar.gz && \
        sudo tar xzf rexray-$(uname -s)-$ARCH.tar.gz -C /usr/bin && \
        rm -f rexray-$(uname -s)-$ARCH.tar.gz && \
        sudo chmod 0755 /usr/bin/rexray && \
        sudo chown 0 /usr/bin/rexray && \
        sudo chgrp 0 /usr/bin/rexray && \
        sudo /usr/bin/rexray install

fi

echo
echo "REX-Ray has been installed to /usr/bin/rexray"
echo
/usr/bin/rexray version
echo