# Deliver the experiment server as a container
#FROM alpine:latest
FROM ubuntu:16.04

EXPOSE 30001

#RUN apk --no-cache add ca-certificates
RUN apt-get -y update && apt-get -y install net-tools iproute2 curl libssl1.0.0 libgflags2v5
WORKDIR /root/
COPY bin/experimentsrv .
COPY run.sh .
RUN apt-get -y install lsb-core wget && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ `lsb_release -cs`-pgdg main" >> /etc/apt/sources.list.d/pgdg.list' && \
    apt-get -y update && \
    apt-get -y install postgresql-client-11
CMD ["./run.sh"]
