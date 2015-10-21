Date: October 2015.

Getting the srndv2 tool

I am using debian, you should be able to use most any linux distro for this. Known to work are: debian, arch linux, <TODO: add  more>.

Most commands should be done as a normal user, but some special commands need to be done as root. I find it useful to have two terminals open. I'll denote normal user level commands with '$' and root command with '#'.

Some dependencies you will need to install (as root) are:
 
# apt-get install build-essential golang git
# apt-get install libsodium-dev ffmpegthumbnailer
# apt-get install imagemagick ffmpegthumbnailer sox

The source code is in these two repos:

* https://github.com/majestrate/nntpchan
* https://github.com/majestrate/srndv2

set up your GOPATH (notes on that here: https://golang.org/doc/code.html#GOPATH ) and then install and build it:

$ go get -u github.com/majestrate/srndv2

If that command didn't work read the errors and check if you lacked any dependencies.

Now you have the srndv2 tool which you can run, but it will not work yet: You need to step up an SQL database first.

--------------

Setting up an SQL database

* https://wiki.postgresql.org/wiki/Detailed_installation_guides

Install postgresql.

# apt-get install postgresql postgresql-client

Create a postgresql user called 'srnd' and a database 'srnd':

# su postgres
$ whoami
postgres
$ psql -f nntpchan/nntp.psql

TODO: Get correct filename here.

Test if you can log in to that SQL user this way:

$ psql -d srnd -U srnd

If there is an issue with that try the following from the debian wiki:

------------
edit pg_hba.conf in /etc/postgresql/X.Y/main/pg_hba.conf

local   all         all                               trust     # replace ident or peer with trust

reload postgresql

# /etc/init.d/postgresql reload
------------

hit Contol-D to get back your root terminal after doing this.


Once SQL setup is successful..


Now as your regular user that installed the srndv2 tool, you should be able to set up srndv2

First clone nntpchan and cd into it, then ask the srndv2 tool to setup your node:

$ git clone https://github.com/majestrate/nntpchan.git
$ cd nntpchann
ntpchan/$ srndv2 setup
ntpchan/$ srndv2 tool keygen
