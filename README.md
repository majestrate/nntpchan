# NNTPChan #

NNTPChan (previously known as overchan) is a decentralized imageboard that uses nntp to synchronize content between many different servers. It utilizes cryptograpghicly signed posts to perform optional/opt-in decentralized moderation (currently work-in-progress)

## getting started ##

If you don't want to compile from source, you can download a precompiled binary [here](https://github.com/majestrate/srndv2/releases) when they are released.

After you [built and installed the daemon](build-daemon.md) and [set up your database](database-setup.md), clone this repository and start up the daemon

    # clone it
    git clone https://github.com/majestrate/nntpchan
    cd nntpchan

    # set up the workspace
    srndv2 setup

    # run the daemon
    srndv2 run

Then open http://127.0.0.1:18000/ukko.html in your browser.

*PLEASE* report any bugs you find while setting up or building [(here)](https://github.com/majestrate/nntpchan/issues) so that the problems get fixed (^:

For questions or support find me on [rizon](https://qchat.rizon.net/?channels=#8chan-dev) as \__uguu\__
