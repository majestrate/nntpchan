GOROOT ?= $(shell go env GOROOT)
GO ?= $(GOROOT)/bin/go
REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION=$(shell $(GO) version | cut -d' ' -f3)

all: build

build: srndv2

srndv2:
	GOPATH=$(REPO) GOROOT=$(GOROOT) $(GO) build -ldflags "-X srnd.GitVersion=-$(shell git rev-parse --short HEAD)" -v

srndv2-lua:
	GOPATH=$(REPO) GOROOT=$(GOROOT) $(GO) build -ldflags "-X srnd.GitVersion=-$(shell git rev-parse --short HEAD)" -tags lua -v

clean:
	GOPATH=$(REPO) GOROOT=$(GOROOT) $(GO) clean -v

test:
	GOPATH=$(REPO) GOROOT=$(GOROOT) $(GO) test srnd
