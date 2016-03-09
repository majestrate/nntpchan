# building the daemon #


## requirements ##

* linux or freebsd
* libsodium 1.0 or higher
* imagemagick
* ffmpegthumbnailer
* sox

## supported go versions ##

* `go 1.6` or higher with redis driver

* `go 1.3` or higher without redis driver

## debian ##


Get the dependancies

    sudo apt-get update
    sudo apt-get --no-install-recommends install imagemagick libsodium-dev ffmpegthumbnailer sox build-essential git golang ca-certificates


Check out the repo and build it

    git clone https://github.com/majestrate/nntpchan
    cd nntpchan
    ./build.sh

If you want to build without supporting redis then build with the `--no-redis` flag

    ./build.sh --no-redis

Now configure the database. [next](database.md)
