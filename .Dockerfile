FROM golang:@GO_VERSION@ as rexray@FNAME_SUFFIX@-builder

ENV REXRAY rexray@FNAME_SUFFIX@
ENV GOPATH @DGOPATH@

RUN mkdir -p @DGOPATH@/{pkg,src,bin}
COPY fbf.sh /usr/local/bin

WORKDIR @WORKDIR_RR@
@COPY_RR_SRCS_CMD@

@WORKDIR_LS@
@INIT_LS_SRCS_CMD@

WORKDIR @WORKDIR_RR@
RUN @BUILD_CMD@

WORKDIR ${GOPATH}
RUN apt-get update && apt-get install -y --no-install-recommends file \
	&& rm -rf /var/lib/apt/lists/*
RUN file bin/@GOOS_GOARCH_DIR@rexray@FNAME_SUFFIX@
RUN fbf.sh

FROM alpine:3.5

LABEL build="@BUILD_TYPE@"
LABEL drivers="@DRIVERS@"
LABEL version="@SEMVER@"

COPY --from=rexray@FNAME_SUFFIX@-builder @DGOPATH@/bin/@GOOS_GOARCH_DIR@$REXRAY /usr/bin/$REXRAY
COPY @DOCKERFILE@ /Dockerfile

RUN apk update
RUN apk add xfsprogs e2fsprogs ca-certificates

RUN mkdir -p /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN mkdir -p /etc/rexray /run/docker/plugins /var/lib/libstorage/volumes

ENTRYPOINT [ "/usr/bin/rexray@FNAME_SUFFIX@" ]
