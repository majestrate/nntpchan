Building NNPTChan
=================

This document will help you build the NNTPChan software from the source code.

## Requirements

NNTPChan can run on the following operating systems:

* Linux
    * Instructions are available for Debian and [Trisquel](#trisquel-instructions-wip).
* FreeBSD

Dependancies:

* libsodium _1.0_ or _higher_
* imagemagick
* ffmpeg
* sox
* go _1.6_ or higher
* GNU make

## Debian instructions

These are installation instructions for Debian.

### Install Go

Install the Go programming language version _1.6_ from the [Go website](https://golang.org/dl/).

### Install the dependancies

    sudo apt-get update
    sudo apt-get --no-install-recommends install imagemagick libsodium-dev ffmpeg sox build-essential git ca-certificates postgresql postgresql-client

### Get the NNTPChan source

    git clone https://github.com/majestrate/nntpchan --depth=1
    cd nntpchan

### Now compile!

Run `make`:

    make


## Trisquel instructions (WIP)

These are installation instructions for Trisquel.

### Install Go

Run this to install Go.

    sudo apt-get update
    sudo apt-get install golang-1.6

### Installing dependancies (standard)

    sudo apt-get --no-install-recommends install imagemagick libsodium-dev sox build-essential git ca-certificates postgresql postgresql-client

### Installing dependancies (`ffmpeg` is not available in Trisquel repos - there must be compiled)

This will install `ffmpeg` to `/usr/local/bin/ffmpeg`:

    git clone https://git.ffmpeg.org/ffmpeg.git ffmpeg --depth=1
    cd ffmpeg
    ./configure --disable-yasm
    make
    sudo make install
    cd ..
    rm -rf ffmpeg

### Get the NNTPChan source

    git clone https://github.com/majestrate/nntpchan --depth=1
    cd nntpchan

### Now compile!

run `make`

    make
