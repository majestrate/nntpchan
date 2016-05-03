#!/usr/bin/env bash
root=$(readlink -e $(dirname $0))
set -e
if [ "x" == "x$root" ] ; then
    root=$PWD/${0##*}
fi
cd $root

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

rev="QmaNuKBcG3hb5YJ4xpeipdX3t2Fw6pwZJSnpvsfn9Zj1tm"

_next=""
# check for build flags
for arg in $@ ; do
    case $arg in
        "--disable-redis")
            tags="$tags -tags disable_redis"
            ;;
        "--revision")
            _next="rev"
            ;;
        "--revision=*")
            rev=$(echo $arg | cut -d'=' -f2)
            ;;
        *)
            if [ "x$_next" == "xrev" ] ; then
                rev="$arg"
            fi
    esac
done

if [ "x$rev" == "x" ] ; then
    echo "revision not specified"
    exit 1
fi

cd $root
#echo "obtaining gx"
#go get -v github.com/whyrusleeping/gx
#go get -v github.com/whyrusleeping/gx-go
#gx install --global && go build -v
export GOPATH=$PWD/go
mkdir -p $GOPATH
go get -u -v github.com/majestrate/srndv2
cp $GOPATH/bin/srndv2 $root
echo "Built"
