FROM ubuntu:14.04
MAINTAINER <EMC{code}>

# See ./.docs/dev-guide/build-reference.md under "Build with Docker" for usage

RUN apt-get update && apt-get install -y --no-install-recommends software-properties-common
RUN add-apt-repository ppa:masterminds/glide

# gcc for cgo
RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
        debhelper \
        dpkg-dev \
        fakeroot \
        g++ \
        gcc \
        git \
        glide \
        libc6-dev \
        libfakeroot \
        make \
        rpm \
        wget \
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
ENV DOCKER_MARKER 1

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH

VOLUME ["/go/src/github.com/emccode/rexray/", "/go/pkg/", "/go/bin/"]

CMD ["/bin/bash", "-c", "make clean \
&& make deps \
&& env GOOS=linux GOARCH=amd64 make build \
&& env GOOS=darwin GOARCH=amd64 make build \
&& make tgz \
&& make RPMBUILD='fakeroot rpmbuild' rpm \
&& make deb"]
