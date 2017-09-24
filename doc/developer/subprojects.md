# subprojects

This is a list of subprojects related to the main nntpchan daemon

Mostly rewrite attempts

## nntpchand

Golang refactor of srndv2

Build Requirements:

* go 1.9

* GNU Make

To build:

    $ make beta
    
To run tests:
    
    $ make test-beta
    
To clean:
    
    $ make clean-beta
    
## nntpd

Native C++ rewrite of nntpchan daemon

Build Requirements:

* GNU Make

* clang c++ compiler

* pkg-config

* libuv-1.x

* libsodium-1.x

To build:

    $ make native
    
To run tests:

    $ make native-test
    
To clean:

    $ make clean-native
    
    
## Tests

    $ make test-full
