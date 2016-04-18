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

rev="QmSLsdX9BQ1f9sBeuMeLnrmFEWQiNJG9nRmgi4Pua2Ui3y"

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
export GOPATH=$root/go
mkdir -p $GOPATH
echo "obtaining gx"
go get -v github.com/whyrusleeping/gx
go get -v github.com/whyrusleeping/gx-go
export GO15VENDOREXPERIMENT=1
$GOPATH/bin/gx install
go build -v
echo "Built"
