FROM golang:1.18.3-alpine3.16 AS builder

WORKDIR /usr/src/app/

ADD go.mod go.sum /usr/src/app/
ADD *.go /usr/src/app/
RUN go build

FROM alpine:3.16

WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/metadater /usr/local/bin/metadater

VOLUME ["/etc/metadater"]

USER nobody
CMD ["metadater"]
