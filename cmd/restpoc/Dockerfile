# Deliver the experiment manager as a container
FROM alpine:latest
RUN apk --no-cache add ca-certificates

RUN adduser --system app --disabled-password
USER app

WORKDIR /home/app
COPY bin/restpoc .
CMD ["./restpoc"]
