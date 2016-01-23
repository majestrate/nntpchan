# NNTPChan #

NNTPChan (previously known as overchan) is a decentralized imageboard that uses nntp to synchronize content between many different servers. It utilizes cryptographically signed posts to perform optional/opt-in decentralized moderation.

This repository contains resources used by the core daemon which is located [here](https://github.com/majestrate/srndv2) along with general documentation, [here](doc/)

## getting started ##

Get the dependancies

    sudo apt-get update
    sudo apt-get --no-install-recommends install imagemagick libsodium-dev ffmpeg sox build-essential git golang ca-certificates

Check out this repo and build it

    git clone https://github.com/majestrate/nntpchan
    cd nntpchan
    ./build.sh

Now configure the database. [Next](doc/database.md)


---

*PLEASE* report any bugs you find while setting up or building [(here)](https://github.com/majestrate/nntpchan/issues) so that the problems get fixed :^)

For peering requests, questions or support find me on [rizon](https://qchat.rizon.net/?channels=#nntpchan) as \__uguu\__


Like this project? Fund it:

bitcoin: 15yuMzuueV8y5vPQQ39ZqQVz5Ey98DNrjE

