REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all: clean build

build: nntpchand

nntpchand:
	GOPATH=$(REPO) go build -v

clean:
	GOPATH=$(REPO) go clean -v
