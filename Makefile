.PHONY: golang python deps build test

REPO_PATH := github.com/projecteru2/cli
REVISION := $(shell git rev-parse HEAD || unknown)
BUILTAT := $(shell date +%Y-%m-%dT%H:%M:%S)
VERSION := $(shell git describe --tags $(shell git rev-list --tags --max-count=1))
GO_LDFLAGS ?= -s -X $(REPO_PATH)/versioninfo.REVISION=$(REVISION) \
			  -X $(REPO_PATH)/versioninfo.BUILTAT=$(BUILTAT) \
			  -X $(REPO_PATH)/versioninfo.VERSION=$(VERSION)

binary:
	go build -ldflags "$(GO_LDFLAGS)" -a -tags "netgo osusergo" -installsuffix netgo -o eru-cli

deps:
	env GO111MODULE=on go mod download
	env GO111MODULE=on go mod vendor

build: deps binary

test: deps
	go vet `go list ./... | grep -v '/vendor/'`
	go test -v `go list ./... | grep -v '/vendor/'`
