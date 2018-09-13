FROM golang:1.11.0-alpine3.8 AS BUILD

MAINTAINER CMGS <ilskdw@gmail.com>

# make binary
RUN apk add --no-cache git ca-certificates curl make gcc libc-dev \
    && curl https://glide.sh/get | sh \
    && go get -d github.com/projecteru2/cli
WORKDIR /go/src/github.com/projecteru2/cli
RUN make build && ./eru-cli --version

FROM alpine:3.8

MAINTAINER CMGS <ilskdw@gmail.com>

COPY --from=BUILD /etc/ssl/certs /etc/ssl/certs
COPY --from=BUILD /go/src/github.com/projecteru2/cli/eru-cli /usr/bin/eru-cli
