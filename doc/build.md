# building the daemon #


## requirements ##

* linux or freebsd
* go 1.3 or higher
* libsodium 1.0 or higher
* imagemagick
* ffmpegthumbnailer
* sox

## debian ##


Get the dependancies

    sudo apt-get update
    sudo apt-get --no-install-recommends install imagemagick libsodium-dev ffmpegthumbnailer sox build-essential git golang ca-certificates


Check out the repo and build it

    git clone https://github.com/majestrate/nntpchan
    cd nntpchan
    ./build.sh

Now configure the database. [next](database.md)
