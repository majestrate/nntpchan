Building the NNTPChan server
============================

This document will help you setup the NNTPChan server from the source code.

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

##Debian instructions

These are installation instructions for Debian.

###Install Go

Install the Go programming language version _1.6_ from the [Go website](https://golang.org/dl/).

###Install the dependancies

    sudo apt-get update
    sudo apt-get --no-install-recommends install imagemagick libsodium-dev ffmpeg sox build-essential git ca-certificates

###Get the NNTPChan source

    git clone https://github.com/majestrate/nntpchan --depth=1
    cd nntpchan

###Now compile!

If you want to compile with Redis support (recommended - Redis is easy to use) then run:

    ./build.sh

If you want to build without support for Redis then build with the `--no-redis` flag:

    ./build.sh --no-redis

##Trisquel instructions

These are installation instructions for Trisquel.

###Installing dependancies

####Standard dependancies

    sudo apt-get update
    sudo apt-get --no-install-recommends install imagemagick libsodium-dev sox build-essential git ca-certificates

####Compiling and installing `ffmpeg` (`ffmpeg` is not available in Trisquel repos)

This will install `ffmpeg` to `/sur/local/bin/ffmpeg`:

    git clone https://git.ffmpeg.org/ffmpeg.git ffmpeg --depth=1
    cd ffmpeg
    ./configure --disable-yasm
    make
    sudo make install
    rm -rf ffmpeg

###Install Go

Run this to install Go.

    sudo apt-get install golang-1.6

###Installing Redis

Run this to install Redis.

    sudo apt-get install redis-server

###Installing Postgres (WIP)

Run this to install Postgres.

    sudo apt-get install Postgres

###Get the NNTPChan source

    git clone https://github.com/majestrate/nntpchan --depth=1
    cd nntpchan

###Now compile!

If you want to compile with Redis support (recommended - Redis is easy to use) then run:

    ./build.sh

If you want to build without support for Redis then build with the `--no-redis` flag:

    ./build.sh --no-redis
