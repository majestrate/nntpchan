#!/usr/bin/env bash
root=$(readlink -e "$(dirname "$0")")
set -e
if [ "" == "$root" ] ; then
    root=$PWD/${0##*}
fi
cd "$root"

tags=""

help_text="usage: $0 [--disable-redis]"

# check for help flags first
for arg in "$@" ; do
    case $arg in
        -h|--help)
            echo "$help_text"
            exit 0
            ;;
    esac
done

rev="QmPAqM7anxdr1ngPmJz9J9AAxDLinDz2Eh9aAzLF9T7LNa"
ipfs="no"
rebuildjs="yes"
_next=""
unstable="no"
# check for build flags
for arg in "$@" ; do
    case $arg in
        "--unstable")
            unstable="yes"
            ;;
        "--no-js")
            rebuildjs="no"
            ;;
        "--ipfs")
            ipfs="yes"
            ;;
        "--cuckoo")
            cuckoo="yes"
            ;;
        "--disable-redis")
            tags="$tags -tags disable_redis"
            ;;
        "--revision")
            _next="rev"
            ;;
        "--revision=*")
            rev=$(echo "$arg" | cut -d'=' -f2)
            ;;
        *)
            if [ "x$_next" == "xrev" ] ; then
                rev="$arg"
            fi
    esac
done

if [ "$rev" == "" ] ; then
    echo "revision not specified"
    exit 1
fi

cd "$root"
if [ "$rebuildjs" == "yes" ] ; then
    echo "rebuilding generated js..."
    ./build-js.sh
fi
unset GOPATH
export GOPATH=$PWD/go
mkdir -p "$GOPATH"

if [ "$ipfs" == "yes" ] ; then
    if [ ! -e "$GOPATH/bin/gx" ] ; then
        echo "obtaining gx"
        go get -u -v github.com/whyrusleeping/gx
    fi
    if [ ! -e "$GOPATH/bin/gx-go" ] ; then
        echo "obtaining gx-go"
        go get -u -v github.com/whyrusleeping/gx-go
    fi
    echo "building stable revision, this will take a bit. to speed this part up install and run ipfs locally"
    mkdir -p "$GOPATH/src/gx/ipfs"
    cd "$GOPATH/src/gx/ipfs"
    "$GOPATH/bin/gx" get "$rev"
    cd "$root"
    go get -d -v
    go build -v .
    mv nntpchan srndv2
    echo -e "Built\n"
    echo "Now configure NNTPChan with ./srndv2 setup"
else
    if [ "$unstable" == "yes" ] ; then
        go get -u -v github.com/majestrate/srndv2/cmd/nntpchan
        cp "$GOPATH/bin/nntpchan" "$root"
        echo "built unstable, if you don't know what to do, run without --unstable"
    else
        go get -u -v github.com/majestrate/srndv2
        cp "$GOPATH/bin/srndv2" "$root"
        echo -e "Built\n"
        echo "Now configure NNTPChan with ./srndv2 setup"
    fi
fi
