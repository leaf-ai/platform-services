# Deliver the experiment server as a container
#FROM alpine:latest
FROM ubuntu:16.04

EXPOSE 30001

#RUN apk --no-cache add ca-certificates
RUN apt-get -y update && apt-get -y install net-tools iproute2 curl libssl1.0.0 libgflags2v5
RUN apt-get -y update && \
    echo "wireshark-common wireshark-common/install-setuid boolean true" | debconf-set-selections && \
    DEBIAN_FRONTEND=noninteractive apt-get -y install wireshark-common wireshark && \
    dpkg-reconfigure wireshark-common && \
    gpasswd -a root wireshark && \
    setcap 'CAP_NET_RAW+eip CAP_NET_ADMIN+eip' /usr/bin/dumpcap && \
    chgrp wireshark /usr/bin/dumpcap

WORKDIR /root/
COPY bin/experimentsrv .
COPY run.sh .
CMD ["./run.sh"]
