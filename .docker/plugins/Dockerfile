FROM alpine:3.6

RUN apk update
RUN apk add xfsprogs e2fsprogs ca-certificates

RUN mkdir -p /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN mkdir -p /etc/rexray /run/docker/plugins /var/lib/rexray/volumes
ADD rexray /usr/bin/rexray
ADD rexray.yml /etc/rexray/rexray.yml

ADD rexray.sh /rexray.sh
RUN chmod +x /rexray.sh

CMD [ "rexray", "start", "--nopid" ]
ENTRYPOINT [ "/rexray.sh" ]
