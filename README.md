# NNTPChan #

NNTPChan (previously known as overchan) is a decentralized imageboard that uses nntp to synchronize content between many different servers. It utilizes cryptograpghicly signed posts to perform optional/opt-in decentralized moderation.

This repository contains resources used by the core daemon which is located [here](https://github.com/majestrate/srdnv2) along with general documentation, [here](doc/)

## getting started ##

After you [built and installed the daemon](doc/build.md) and [set up your database](doc/database.md), clone this repository and start up the daemon

    # clone it
    git clone https://github.com/majestrate/nntpchan
    cd nntpchan

    # set up the workspace
    srndv2 setup

    # run the daemon
    srndv2 run

Then open http://127.0.0.1:18000/ukko.html in your browser.

*PLEASE* report any bugs you find while setting up or building [(here)](https://github.com/majestrate/nntpchan/issues) so that the problems get fixed (^:

For peering requests, questions or support find me on [rizon](https://qchat.rizon.net/?channels=#nntpchan) as \__uguu\__


Like this project? Fund it:

bitcoin: 15yuMzuueV8y5vPQQ39ZqQVz5Ey98DNrjE

