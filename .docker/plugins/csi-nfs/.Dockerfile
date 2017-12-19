FROM centos:7.4.1708

RUN yum install -y nfs-utils && yum clean all && rm -rf /var/cache/yum

RUN mkdir -p /etc/rexray /run/docker/plugins /var/lib/rexray/volumes
ADD rexray /usr/bin/rexray
ADD rexray.yml /etc/rexray/rexray.yml

ADD .rexray.sh /rexray.sh
RUN chmod +x /rexray.sh

CMD [ "rexray", "start", "--nopid" ]
ENTRYPOINT [ "/rexray.sh" ]
