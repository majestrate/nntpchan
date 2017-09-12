GO ?= $(GOROOT)/bin/go
REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION=$(shell $(GO) version | cut -d' ' -f3)

all: build

build: srndv2

assert-go:
	test $(VERSION) = go1.9

srndv2: assert-go
	GOPATH=$(REPO) GOROOT=$(GOROOT) $(GO) build -v

clean:
	GOPATH=$(REPO) GOROOT=$(GOROOT) $(GO) clean -v

test:
	GOPATH=$(REPO) GOROOT=$(GOROOT) $(GO) test -v srnd
