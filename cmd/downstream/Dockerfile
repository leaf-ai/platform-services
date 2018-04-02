# Deliver the downstream service as a container
#FROM alpine:latest
FROM ubuntu:16.04

EXPOSE 30002

#RUN apk --no-cache add ca-certificates
RUN apt-get -y update && apt-get -y install net-tools iproute2 curl libssl1.0.0 libgflags2v5
WORKDIR /root/
COPY bin/downstream .
CMD ["./downstream"]
