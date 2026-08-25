FROM golang:1.27-alpine AS build

RUN apk add --no-cache git make
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG KEEP_SYMBOL
RUN make build && ./eru-cli --version

FROM alpine:3.22

LABEL ERU=1
COPY --from=build /src/eru-cli /usr/bin/eru-cli
