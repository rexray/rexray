FROM golang:@GO_VERSION@ as rexray@FNAME_SUFFIX@-builder

ENV REXRAY=rexray@FNAME_SUFFIX@

WORKDIR @WORKDIR_RR@
@COPY_RR_SRCS_CMD@

@WORKDIR_LS@
@INIT_LS_SRCS_CMD@

WORKDIR @WORKDIR_RR@
RUN @BUILD_CMD@
RUN /go/bin/$REXRAY version

FROM alpine:3.5

LABEL build="@BUILD_TYPE@"
LABEL drivers="@DRIVERS@"
LABEL version="@SEMVER@"

COPY --from=rexray@FNAME_SUFFIX@-builder /go/bin/$REXRAY /usr/bin/$REXRAY
COPY @DOCKERFILE@ /Dockerfile

RUN apk update
RUN apk add xfsprogs e2fsprogs ca-certificates

RUN mkdir -p /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN mkdir -p /etc/rexray /run/docker/plugins /var/lib/libstorage/volumes

ENTRYPOINT [ "/usr/bin/rexray@FNAME_SUFFIX@" ]
