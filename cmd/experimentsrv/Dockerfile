# Deliver the experiment server as a container
#FROM alpine:latest
FROM ubuntu:16.04

EXPOSE 30001

#RUN apk --no-cache add ca-certificates
RUN apt-get -y update && apt-get -y install net-tools iproute2 curl libssl1.0.0 libgflags2v5
RUN apt-get -y update && apt-get -y install iputils-ping telnet

RUN adduser --system app --group
USER app

WORKDIR /home/app

COPY run.sh .
COPY bin/experimentsrv .
CMD ["./run.sh"]
