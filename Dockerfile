FROM golang:1.9.1-alpine3.6 AS BUILD

MAINTAINER CMGS <ilskdw@gmail.com>

# make binary
RUN apk add --no-cache git curl make \
    && curl https://glide.sh/get | sh \
    && go get -d github.com/projecteru2/cli
WORKDIR /go/src/github.com/projecteru2/cli
RUN make build && ./erucli --version

FROM alpine:3.6

MAINTAINER CMGS <ilskdw@gmail.com>

COPY --from=BUILD /go/src/github.com/projecteru2/cli/erucli /usr/bin/erucli
