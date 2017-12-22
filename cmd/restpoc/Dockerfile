# Deliver the experiment manager as a container
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY bin/restpoc .
CMD ["./restpoc"]
