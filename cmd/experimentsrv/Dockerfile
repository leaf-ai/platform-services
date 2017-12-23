# Deliver the experiment server as a container
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY bin/experimentsrv .
CMD ["./experimentsrv"]
