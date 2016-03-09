#!/usr/bin/env bash
set -e
root=$(readlink -e $(dirname $0))

tags=""

help_text="usage: $0 [--disable-redis]"

# check for help flags first
for arg in $@ ; do
    case $arg in
        -h|--help)
            echo $help_text
            exit 0
            ;;
    esac
done

# check for build flags
for arg in $@ ; do
    case $arg in
        "--disable-redis")
            tags="$tags -tags disable_redis"
            ;;
    esac
done

cd $root
export GOPATH=$root/go
mkdir -p $GOPATH
go get -v -u $tags github.com/majestrate/srndv2
cp -a $GOPATH/bin/srndv2 $root
echo "Built"
