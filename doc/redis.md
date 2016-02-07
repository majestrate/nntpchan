# configuring redis database backend


0) Install redis

0.A) debian/ubuntu

    # apt update
    # apt install redis-server

0.B) redhat

    # yum install redis

0.C) from source

* see http://redis.io/download

    

1) Configuration

In srnd.ini the database sections should look like this:

    [database]
    type=redis
    schema=single
    host=localhost
    port=6379
    user=
    password=

2) Run the daemon

* see the [next step](running.md)
