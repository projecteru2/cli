FROM golang:alpine AS BUILD

MAINTAINER CMGS <ilskdw@gmail.com>

# make binary
RUN apk add --no-cache git ca-certificates curl make gcc libc-dev \
    && go get -d github.com/projecteru2/cli
WORKDIR /go/src/github.com/projecteru2/cli
RUN make build && ./eru-cli --version

FROM alpine:latest

MAINTAINER CMGS <ilskdw@gmail.com>

COPY --from=BUILD /etc/ssl/certs /etc/ssl/certs
COPY --from=BUILD /go/src/github.com/projecteru2/cli/eru-cli /usr/bin/eru-cli
