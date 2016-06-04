Building the NNTPChan server
============================

This document will help you setup the NNTPChan server.

##Requirements

NNTPChan can run on the following operating systems:

* Linux
* FreeBSD

Dependancies:

* libsodium _1.0_ or _higher_
* imagemagick
* ffmpeg
* sox
* go _1.6_ or _higher_ **with redis driver**
* go _1.3_ or _higher_ **without redis driver**

## debian ##

Get `go 1.6` from [here](https://golang.org/dl/) for your platform

Get the dependancies

    sudo apt-get update
    sudo apt-get --no-install-recommends install imagemagick libsodium-dev ffmpeg sox build-essential git ca-certificates


Check out the repo and build it

    git clone https://github.com/majestrate/nntpchan
    cd nntpchan
    ./build.sh

If you want to build without supporting redis then build with the `--no-redis` flag

    ./build.sh --no-redis

To run eiter run `./srndv2 setup` and browse to http://127.0.0.1:18000 or configure [by hand](database.md)
