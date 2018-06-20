FROM golang:1.10.3-alpine3.7 AS BUILD

MAINTAINER CMGS <ilskdw@gmail.com>

# make binary
RUN apk add --no-cache git ca-certificates curl make \
    && curl https://glide.sh/get | sh \
    && go get -d github.com/projecteru2/cli
WORKDIR /go/src/github.com/projecteru2/cli
RUN make build && ./erucli --version

FROM alpine:3.7

MAINTAINER CMGS <ilskdw@gmail.com>

COPY --from=BUILD /go/src/github.com/projecteru2/cli/erucli /usr/bin/erucli
