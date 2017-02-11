FROM centos

LABEL drivers="${DRIVERS}"
LABEL version="${VERSION}"

RUN yum update -y
RUN yum install xfsprogs e2fsprogs -y

RUN mkdir -p /etc/rexray /run/docker/plugins /var/lib/libstorage/volumes
ADD rexray /usr/bin/rexray
ADD rexray.yml /etc/rexray/rexray.yml

ADD rexray.sh /rexray.sh
RUN chmod +x /rexray.sh

ENTRYPOINT [ "/rexray.sh", "rexray", "start", "-f" ]
