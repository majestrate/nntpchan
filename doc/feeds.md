`feeds.ini`
===========

##Peering

In order to actually be distributed, you need another person to sync posts with, otherwise what's the point right?

###Peering over the regular internet with TLS

Requirements:

* Each Side's server must have a domain name
* Each Side must have each other's certificates (.crt files) in the `certs` folder
* Each Side's `srnd.ini` crypto section must have entry `tls-hostname = domain.tld` where domain.tld is the domain name of the server it is on
* Each Side's `srnd.ini` nntp section must have entry `require_tls = 1`


If Alice owns `nntp.alice.net` and Bob owns `nntp.bob.com` and are both using port `1199` the configs for each side are as follows:

    # alice's srnd.ini crypto section
    ...
    [crypto]
    tls-hostname = nntp.alice.net
    tls-trust-dir = certs
    tls-keyname = overchan-alice


    # alice's feeds.ini
    [feed-bob]
    host = nntp.bob.com
    port = 1199
    
    [bob]
    overchan.* = 1
    ctl = 1



    # bob's srnd.ini crypto section
    ...
    [crypto]
    tls-hostname = nntp.bob.com
    tls-trust-dir = certs
    tls-keyname = overchan-bob



    # bob's feeds.ini
    [feed-alice]
    host = nntp.alice.net
    port = 1199

    [alice]
    overchan.* = 1
    ctl = 1

Each side's `certs` directory contains 2 files:

* overchan-alice-nntp.alice.net.crt (alice's certificate)
* overchan-bob-nntp.bob.com.crt (bob's certificate)

Alice keeps `overchan-alice-nntp.alice.net.key` secret and never shares it

Bob keeps `overchan-bob-nntp.bob.com.key` secret and never shares it


###Peering Authentication with passwords

adding / removing credentials via the command line:

    # add an nntp login via command line
    srndv2 tool nntp add-login user-name-here pass-word-here
    
    # remove an nntp login via command line
    srndv2 tool nntp del-login user-name-here

Example `feeds.ini`:

    # section pair in feeds.ini
    # connects to nntp.something.tld:1199 and authenticates with a username and password
    # sync = 1 makes you download all applicable posts from the remote server on startup

    [feed-authenticated]
    host = nntp.something.tld
    port = 1199
    username = user-user-here
    password = pass-word-here
    sync = 1

    [authenticated]
    overchan.* = 1
    ctl = 1
     

###Peering over Tor

Install Tor

    apt-get install tor

Make a tor hidden service point from outside port 119 to port 1199
Add to /etc/tor/torrc:

    HiddenServiceDir /var/lib/tor/nntp_feed
    HiddenServicePort 119 127.0.0.1:1199

restart/reload tor then

    cat /var/lib/tor/nntp_feed/hostname

This is your in feed address

If you use an onion with tls, `srnd.ini` crypto section should have the entry `tls-hostname = youroniongoeshere.onion`. If you don't use tls NEVER disclose the onion address to anyone not trusted.

Then to peer with someone over tor add this to you feeds.ini

    [feed-ourpeer.onion]
    host=PeersOnionAddress.onion
    port=119
    proxy-type=socks4a
    proxy-host=127.0.0.1
    proxy-port=9050

    [ourpeer.onion]
    overchan=1
    ctl=1


##Options

####You need one connection and one settings block for each connection

Here is an example entry in feeds.ini

    [feed-them.onion]
    host=aabbccddeeff2233.onion
    port=119
    proxy-type=socks4a
    proxy-host=127.0.0.1
    proxy-port=9050
    username=somerandomusername
    password=somerandompassword

    [them.onion]
    overchan=1
    ano.paste=0
    ctl=1

But what does it mean?

    [feed-them.onion]

Connection settings for a peer

    host=aabbccddeeff2233.onion
    port=119
    proxy-type=socks4a
    proxy-host=127.0.0.1
    proxy-port=9050
    
Proxy settings, straight forward. Supported proxy types are `socks4a` and `none`

    [them.onion]

NNTP synchronization settings

    overchan=1

Sync all boards, use 

    overchan.bad=0

to prevent certain boards from syncing with certain peers. It can be used to keep bad boards out or keep exclusive boards in

    ano.paste=0

This WILL be the nntpchan pastebin, but it's not implemented yet

    ctl=1

Allows you to recieve moderation notifications from other boards, it's also used for decentralized moderation

##Alternative config location

If you would like to have your feeds.ini somewhere other than in the working directory, you can set the `SRND_FEEDS_INI_PATH` environment variable. For example, if you would like to use `/etc/nntpchan/meems.ini`, edit `~/.profile` and add `export SRND_FEEDS_INI_PATH=/etc/nntpchan/meems.ini`. 
