FROM golang:alpine AS BUILD

# make binary
RUN apk add --no-cache git ca-certificates curl make gcc libc-dev
RUN git clone https://github.com/projecteru2/cli.git /go/src/github.com/projecteru2/cli
WORKDIR /go/src/github.com/projecteru2/cli
RUN make build && ./eru-cli --version

FROM alpine:latest

COPY --from=BUILD /etc/ssl/certs /etc/ssl/certs
COPY --from=BUILD /go/src/github.com/projecteru2/cli/eru-cli /usr/bin/eru-cli
