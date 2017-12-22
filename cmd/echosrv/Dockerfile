# Deliver the example echo server as a container
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY bin/echosrv .
CMD ["./echosrv"]
