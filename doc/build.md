# building the daemon #


## requirements ##

* linux or freebsd
* libsodium 1.0 or higher
* imagemagick
* ffmpeg
* sox

## supported go versions ##

* `go 1.6` or higher with redis driver

* `go 1.3` or higher without redis driver

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
