REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all: srndv2

srndv2:
	GOPATH=$(REPO) go build -v -tags libsodium

pure:
	GOPATH=$(REPO) go build -v

clean:
	GOPATH=$(REPO) go clean -v

test:
	GOPATH=$(REPO) go test -v -tags libsodium srnd

test-pure:
	GOPATH=$(REPO) go test -v srnd
