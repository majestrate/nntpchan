#!/usr/bin/env bash
set -e
root=$(readlink -e "$(dirname "$0")")

prefix="/opt/nntpchan"

help_text="usage: $0 [--prefix /opt/nntpchan] [-q|--quiet] [-r|--rebuild] [--disable-redis]"

# check for help flags first
for arg in "$@" ; do
    case $arg in
        -h|--help)
            echo "$help_text"
            exit 0
            ;;
    esac
done

_next=""
want_rebuild="0"
want_quiet="0"
build_args=""

# check for main flags
for arg in "$@" ; do
    case $arg in
        -q|--quiet)
            want_quiet="1"
            ;;
        -r|--rebuild)
            want_rebuild="1"
            ;;
        --prefix)
            _next="prefix"
            ;;
        --prefix=*)
            prefix=$(echo "$arg" | cut -d'=' -f2)
            ;;
        --disable-redis)
            build_args="$build_args --disable-redis"
            ;;
        *)
            if [ "X$_next" == "Xprefix" ] ; then
                # set prefix
                prefix="$arg"
                _next=""
            fi
            ;;
    esac
done

_cmd() {
    if [ "X$want_quiet" == "X1" ] ; then
        "$@" &> /dev/null
    else
        "$@"
    fi
}

if [ "X$want_rebuild" == "X1" ] ; then
    _cmd echo "rebuilding daemon";
    _cmd "$root/build.sh" $build_args
fi

if [ ! -e "$root/srndv2" ] ; then
    _cmd echo "building daemon"
    # TODO: use different GOPATH for root?
    _cmd "$root/build.sh" "$build_args"
fi


_cmd mkdir -p "$prefix"
_cmd mkdir -p "$prefix/webroot/thm"
_cmd mkdir -p "$prefix/webroot/img"
_cmd cp -f "$root/srndv2" "$prefix/srndv2"
_cmd cp -rf "$root/"{doc,contrib,certs} "$prefix/"
_cmd echo "installed to $prefix"
