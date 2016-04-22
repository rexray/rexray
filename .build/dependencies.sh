#!/bin/bash

cwd=$(pwd)

# create a temp directory
tmp=$HOME/.tmp && mkdir -p $tmp

########################################################################
##                           coverage                                 ##
########################################################################

go get -v github.com/onsi/gomega
go get -v github.com/onsi/ginkgo
go get -v github.com/axw/gocov/gocov
go get -v github.com/mattn/goveralls
go get -v golang.org/x/tools/cmd/cover

# the remaining steps are specific to linux
if [[ $(uname -s) != Linux ]]; then exit 0; fi

########################################################################
##                             alien                                  ##
########################################################################

alien_home=$HOME/.opt/alien/8.86
alien_pkg=alien_8.86_all.deb
alien_url=http://archive.ubuntu.com/ubuntu/pool/main/a/alien/$alien_pkg

if [[ ! -e $alien_home ]]; then
    cd $tmp &> /dev/null
    if [[ ! -e $alien_pkg ]]; then wget $alien_url; fi
    mkdir -p $alien_home
    dpkg -X $alien_pkg $alien_home
    cd $cwd &> /dev/null

    echo
    which alien && alien --version
    echo
fi

export PATH=$alien_home/usr/bin:$PATH
export PERL5LIB=$alien_home/usr/share/perl5:$PERL5LIB

########################################################################
##                              make                                  ##
########################################################################

make_home=$HOME/.opt/make/4.1
make_pkg=make-4.1.tar.gz
make_url=http://ftp.gnu.org/gnu/make/$make_pkg

if [[ ! -e $make_home ]]; then
    cd $tmp &> /dev/null
    if [[ ! -e $make_pkg ]]; then wget $make_url; fi
    tar xzf $make_pkg
    cd ${make_pkg%.tar.gz} &> /dev/null
    ./configure --prefix=$make_home
    make install
    cd $cwd &> /dev/null

    echo
    which make && make --version
    echo
fi

export PATH=$make_home/bin:$PATH
