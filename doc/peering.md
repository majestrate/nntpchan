## peering with other nodes ##

In order to actually be distributed, you need another person to sync posts with, otherwise what's the point right?

Right now peering information is private, there is no link level authenticatio (yet) so everything is done via either a vpn tunnel or a tor hidden service.

### Peering via cjdns vpn tunnel ###

Set up cjdns, read more [here](https://github.com/cjdelisle/cjdns/blob/master/doc/configure.md#connection-interfaces)

    git clone https://github.com/cjdelisle/cjdns
    cd cjdns && ./do
    ./cjdroute --genconf >> cjdroute.conf
    ./cjdroute < cjdroute.conf

Get your ipv6 address for cjdns

    ip addr show tun0

Edit srnd.ini to bind nntp on that ipv6 address, make sure to use the square brances `[` and `]`

    [nntp]
    ...
    bind=[xxxx:xxxx:xxxx:xxx:xx....]:1199


say you have 2 friends at fc33:3:3::aadd and fc03:9f:123::a3df. right now feeds.ini can't take raw ipv6 addresses so add them to `/etc/hosts`

    # add these lines to /etc/hosts
    fc33:3:3::aadd     bob
    fc03:9f:123::a3df  charlie

then add to feeds.ini the following:


    [feed-bob]
    proxy-type=none

    [bob]
    overchan.*=1
    ctl=1
    
    [feed-charlie]
    proxy-type=none

    [charlie]
    overchan.*=1
    ctl=1

then restart srndv2

**TODO:** firewalling
