REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all: clean build

build: srndv2

srndv2:
	GOPATH=$(REPO) go build -v

clean:
	GOPATH=$(REPO) go clean -v
