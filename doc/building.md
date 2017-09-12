Building NNPTChan
=================

This document will help you build the NNTPChan software from the source code.

## Requirements

NNTPChan can run on the following operating systems:

* Linux
* FreeBSD

Dependancies:

* imagemagick
* ffmpeg
* sox
* go 1.9
* GNU make

## Debian instructions

These are installation instructions for Debian.

### Install Go

Install the Go programming language version _1.9_ from the [Go website](https://golang.org/dl/).

### Install the dependancies

    sudo apt-get update
    sudo apt-get --no-install-recommends install imagemagick ffmpeg sox build-essential git ca-certificates postgresql postgresql-client

#### Configure PostgreSQL:

    # su - postgres -c "createuser --pwprompt --encrypted srnd"
    # su - postgres -c "createdb srnd"

### Get the NNTPChan source

    git clone https://github.com/majestrate/nntpchan --depth=1
    cd nntpchan

### Now compile!

Run `make`:

    make

