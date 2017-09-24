REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all: clean build

build: nntpchand

nntpchand:
	GOROOT=$(GOROOT) GOPATH=$(REPO) go build -v

test:
	GOROOT=$(GOROOT) GOPATH=$(REPO) go test ./...

clean:
	GOPATH=$(REPO) go clean -v
