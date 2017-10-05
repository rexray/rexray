FROM alpine:3.6

RUN apk update
RUN apk add xfsprogs e2fsprogs ca-certificates libaio curl
RUN echo http://dl-cdn.alpinelinux.org/alpine/edge/testing >> /etc/apk/repositories && apk update && apk add numactl
RUN curl -sSL https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.25-r0/sgerrand.rsa.pub > /etc/apk/keys/sgerrand.rsa.pub
RUN curl -sSLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.25-r0/glibc-2.25-r0.apk
RUN apk add glibc-2.25-r0.apk
RUN rm -f glibc-2.25-r0.apk
RUN ln -s /lib/libc.musl-x86_64.so.1 /usr/glibc-compat/lib/
RUN ln -s /lib/libuuid.so.1 /usr/glibc-compat/lib/
RUN ln -s /usr/lib/libaio.so.1 /usr/glibc-compat/lib/
RUN ln -s /usr/lib/libnuma.so.1 /usr/glibc-compat/lib/

RUN mkdir -p /etc/rexray /run/docker/plugins /var/lib/rexray/volumes
ADD rexray /usr/bin/rexray
ADD rexray.yml /etc/rexray/rexray.yml

ADD rexray.sh /rexray.sh
RUN chmod +x /rexray.sh

CMD [ "rexray", "start", "--nopid" ]
ENTRYPOINT [ "/rexray.sh" ]
