# Deliver the example time server as a container
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY ../../bin/timesrv .
CMD ["./timesrv"]
