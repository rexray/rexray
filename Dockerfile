FROM ubuntu:14.04
MAINTAINER <EMC{code}>

# to build this dockerfile first ensure that it is named "Dockerfile"
# make sure that a directory "docker_resources" is also present in the same directory
# as "Dockerfile", and that "docker_resources" contains the files
# "go-wrapper" and "get_go-bindata_md5.sh"

# Assuming:
# Your dockerhub username: dhjanedoe
# Your github username: ghjanedoe
# Your REX-Ray fork is checked out in $HOME/go/src/github.com/ghjanedoe/rexray/

# To build a Docker image using this Dockerfile:
# docker build -t dhjanedoe/golang-glide:0.1.0 .

# To build REX-Ray using this Docker image:
# docker pull dhjanedoe/golang-glide:0.1.0
# docker run -v $HOME/go/src/github.com/ghjanedoe/rexray/:/go/src/github.com/emccode/rexray/ \
# -v $HOME/build/rexray/pkg/:/go/pkg/ \
# -v $HOME/build/rexray/bin/:/go/bin/ \
# -w=/go/src/github.com/emccode/rexray/ dhjanedoe/golang-glide:0.1.0

RUN apt-get update && apt-get install -y --no-install-recommends software-properties-common
RUN add-apt-repository ppa:masterminds/glide

# gcc for cgo
RUN apt-get update && apt-get install -y --no-install-recommends \
        git \
        curl \
        g++ \
        gcc \
        libc6-dev \
        make \
        glide \
    && rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.6.2
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 e40c36ae71756198478624ed1bb4ce17597b3c19d243f3f0899bb5740d56212a

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
    && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
    && tar -C /usr/local -xzf golang.tar.gz \
    && rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH

ADD docker_resources/go-wrapper /usr/local/bin/
ADD docker_resources/get_go-bindata_md5.sh /home/

CMD ["/bin/bash", "-c", "/home/./get_go-bindata_md5.sh; make clean; make deps; make build-all"]
