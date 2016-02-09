#!/usr/bin/env bash
set -e
root=$(readlink -e $(dirname $0))
cd $root
export GOPATH=$root/go
mkdir -p $GOPATH
go get -v -u github.com/majestrate/srndv2
cp -a $GOPATH/bin/srndv2 $root
