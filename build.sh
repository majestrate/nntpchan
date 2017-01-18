#!/usr/bin/env bash
root=$(readlink -e "$(dirname "$0")")
set -e
if [ "" == "$root" ] ; then
    root=$PWD/${0##*}
fi
cd "$root"

tags=""

help_text="usage: $0 [--disable-neochan] [--enable-redis]"

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
neochan="no"
buildredis="no"
# check for build flags
for arg in "$@" ; do
    case $arg in
        "--enable-redis")
            buildredis="yes"
            ;;
        "--enable-neochan")
            neochan="yes"
            ;;
        "--unstable")
            unstable="yes"
            ;;
        "--no-js")
            rebuildjs="no"
            ;;
        "--ipfs")
            ipfs="yes"
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

if [ "$buildredis" == "yes" ] ; then
    tags="$tags -tags disable_redis"
fi

if [ "$rev" == "" ] ; then
    echo "revision not specified"
    exit 1
fi

cd "$root"
if [ "$rebuildjs" == "yes" ] ; then
    echo "rebuilding generated js..."
    if [ "$neochan" == "no" ] ; then
        ./build-js.sh --disable-neochan
    else
        ./build-js.sh
    fi
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
    cd "$rev/srndv2"
    echo "build..."
    go build -v .
    cp srndv2 "$root"
    echo -e "Built\n"
    echo "Now configure NNTPChan with ./srndv2 setup"
else
    if [ "$unstable" == "yes" ] ; then
        make -C contrib/backends/srndv2
        cp contrib/backends/srndv2/nntpchand "$root/nntpchan"
        echo "built unstable, if you don't know what to do, run without --unstable"
    else
        go get -u -v $tags github.com/majestrate/srndv2
        cp "$GOPATH/bin/srndv2" "$root"
        echo -e "Built\n"
        echo "Now configure NNTPChan with ./srndv2 setup"
    fi
fi
