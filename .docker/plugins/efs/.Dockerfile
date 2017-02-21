FROM alpine:3.5

LABEL drivers="${DRIVERS}"
LABEL version="${VERSION}"

RUN apk update
RUN apk add ca-certificates nfs-utils

RUN mkdir -p /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN mkdir -p /etc/rexray /run/docker/plugins /var/lib/libstorage/volumes
ADD rexray /usr/bin/rexray
ADD rexray.yml /etc/rexray/rexray.yml

ADD rexray.sh /rexray.sh
RUN chmod +x /rexray.sh

ENTRYPOINT [ "/rexray.sh", "rexray", "start", "-f" ]
