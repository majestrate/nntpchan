# configuring redis database backend


0) Install redis

Redis 3.x or higher is required, [stable release](http://download.redis.io/releases/redis-stable.tar.gz) recommend

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
