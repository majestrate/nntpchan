# building the daemon #


## requirements ##

* linux or freebsd
* go 1.4 or higher
* libsodium 1.0 or higher
* imagemagick
* ffmpegthumbnailer
* sox

## debian ##

Debian Jessie has go 1.3, we need 1.4 or higher to build the nntpchan daemon so let's do that first, assumes 64bit linux:

    #
    #      --->>  DO NOT RUN AS ROOT <<---
    # (you probably will break stuff really bad if you do)
    #

    # make directory for go's packages
    mkdir -p $HOME/go

    # set up a directory for our go distribution
    mkdir -p $HOME/local
    cd $HOME/local
    
    # obtain and unpack go binary distribution
    wget https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz -O go-stable.tar.gz
    tar -xzvf go-stable.tar.gz

    # set up environmental variables for go
    export GOROOT=$HOME/local/go
    export GOPATH=$HOME/go
    export PATH=$GOROOT/bin:$GOPATH/bin:$PATH

    # put environmental variables in bash_alises for later
    echo 'export GOROOT=$HOME/local/go' >> $HOME/.bash_aliases
    echo 'export GOPATH=$HOME/go' >> $HOME/.bash_aliases
    echo 'export PATH=$GOROOT/bin:$GOPATH/bin:$PATH' >> $HOME/.bash_aliases


We'll also need to install some dependancies that come with debian:

    # as root

    apt update
    apt-get --no-install-recommends install imagemagick libsodium-dev ffmpegthumbnailer sox build-essential

Now build the daemon:

    go get -u github.com/majestrate/srndv2

It will create an executable at $GOPATH/bin/srndv2 which is already in our $PATH so it can be run by typing ``srndv2``

Now configure the database. [next](build.md)
