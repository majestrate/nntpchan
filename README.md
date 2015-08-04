# nntpchan #

## requirements ##

* linux or freebsd
* go 1.4 or higher
* libsodium 1.0 or higher
* imagemagick
* postgresql

## setting up ##

### debian jessie/wheezy ###

Debian doesn't has go 1.3, we need 1.4 or higher to build the nntpchan daemon.

If you don't want to compile from source download a precompiled binary [here](#)

## building the daemon ##

after you have satisfied the dependancies, run:

    go install github.com/majestrate/srndv2
