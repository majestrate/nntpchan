# feeds.ini #

## Peering ##

In order to actually be distributed, you need another person to sync posts with, otherwise what's the point right?

Right now peering information is private, there is no link level authentication (yet) so everything is done via either a vpn tunnel or a tor hidden service.

### Peering over tor ###

Install tor

    apt-get install tor

Make a tor hidden service point from outside port 119 to port 1199
Add to /etc/tor/torrc:

    HiddenServiceDir /var/lib/tor/nntp_feed
    HiddenServicePort 119 127.0.0.1:1199

restart/reload tor then

    cat /var/lib/tor/nntp_feed/hostname

This is your in feed address

Then to peer with someone over tor add this to you feeds.ini

    [feed-PeersOnionAddress.onion:119]
    proxy-type=socks4a
    proxy-host=127.0.0.1
    proxy-port=9050

    [PeersOnionAddress.onion:119]
    overchan=1
    ctl=1


### Peering over cjdns ###

Set up cjdns, read more [here](https://github.com/cjdelisle/cjdns/blob/master/doc/configure.md#connection-interfaces)

    git clone https://github.com/cjdelisle/cjdns
    cd cjdns && ./do
    ./cjdroute --genconf >> cjdroute.conf
    ./cjdroute < cjdroute.conf

Get your ipv6 address for cjdns

    ip addr show tun0

Edit srnd.ini to bind nntp on that ipv6 address, make sure to use the square braces `[` and `]`

    [nntp]
    ...
    bind=[xxxx:xxxx:xxxx:xxx:xx....]:1199


Say you have 2 friends at fc33:3:3::aadd and fc03:9f:123::a3df. right now feeds.ini can't take raw ipv6 addresses so add them to `/etc/hosts`

    # add these lines to /etc/hosts
    fc33:3:3::aadd     bob
    fc03:9f:123::a3df  charlie

Then add to feeds.ini the following:


    [feed-bob]
    proxy-type=none

    [bob]
    overchan=1
    ctl=1
    
    [feed-charlie]
    proxy-type=none

    [charlie]
    overchan=1
    ctl=1


## Options ##

#### You need one connection and one settings block for each connection ####

Here is an example entry in feeds.ini

    [feed-aabbccddeeff2233.onion:119]
    proxy-type=socks4a
    proxy-host=127.0.0.1
    proxy-port=9050

    [aabbccddeeff2233.onion:119]
    overchan=1
    ano.paste=0
    ctl=1

But what does it mean?

    [feed-aabbccddeeff2233.onion:119]

Connection settings for a peer

    proxy-type=socks4a
    proxy-host=127.0.0.1
    proxy-port=9050
    
Proxy settings, straight forward. Supported proxy types are `socks4a` and `none`

    [aabbccddeeff2233.onion:119]

nntp synchronization settings

    overchan=1

Sync all boards, use 

    overchan.bad=0

to prevent certain boards from syncing with certain peers. It can be used to keep bad boards out or keep exclusive boards in

    ano.paste=0

This WILL be the nntpchan pastebin, but it's not implimented yet

    ctl=1

Allows you to recieve moderation notifications from other boards, it's also used for decentralized moderation
