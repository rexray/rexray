#!/bin/sh

FILE=rexray-$(uname -s)-$(uname -m).tar.gz
curl -L "https://dl.bintray.com/emccode/rexray/staged/latest/$FILE" -o $FILE
tar xzf $FILE
sudo cp rexray /usr/bin
if [ "$(uname -s)" != "Darwin" ]; then sudo /usr/bin/rexray service install; fi
echo "REX-Ray has been installed!"
/usr/bin/rexray version