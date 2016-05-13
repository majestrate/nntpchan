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

rev="QmZWEqyb7hs2z1pNUqdfNyDgG3cj31ZZgP148RFdRExCQr"
ipfs="no"
_next=""
# check for build flags
for arg in $@ ; do
    case $arg in
        "--ipfs")
            ipfs="yes"
            ;;
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
unset GOPATH 
export GOPATH=$PWD/go
mkdir -p $GOPATH

if [ "x$ipfs" == "xyes" ] ; then
    if [ ! -e $GOPATH/bin/gx ] ; then
        echo "obtaining gx"
        go get -u -v github.com/whyrusleeping/gx
    fi
    if [ ! -e $GOPATH/bin/gx-go ] ; then
        echo "obtaining gx-go"
        go get -u -v github.com/whyrusleeping/gx-go
    fi
    echo "building stable revision, this will take a bit. to speed this part up install and run ipfs locally"
    $GOPATH/bin/gx install
    go get -d -v 
    go build -v .
    mv nntpchan srndv2
else
    go get -u -v github.com/majestrate/srndv2  
    cp $GOPATH/bin/srndv2 $root
fi
echo "Built"
