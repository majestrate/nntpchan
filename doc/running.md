Running (this should be titled `Manual setup`)
=======

After you have [built the daemon](build.md) and [configured the database](database.md) you can run the daemon.

Read the [Compiling NNTPChan document] for instructions on building NNTPChan if you haven't already. (This shouldn't even be here really)

    git clone https://github.com/majestrate/nntpchan
    cd nntpchan
    ./build.sh

Setup the daemon by running:

    ./srndv2 setup

generate admin keys, don't loose them.

    ./srndv2 tool keygen

add yourself as admin by adding your ``public key`` to the ``frontend`` section of ``srnd.ini``

    ...
    [frontend]
    enable=1
    admin_key=yourpublickeygoeshere
    ... # leave the rest of the config values alone for now

    

run it:

    ./srndv2 run


Now open the browser up to http://127.0.0.1:18000/

To access the mod panel go to the [mod panel](http://127.0.0.1:18000/mod/) and use your ``private key`` to log in

Now read about [peering](feeds.md)
