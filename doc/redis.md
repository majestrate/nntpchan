Configuring Redis database
==========================

These are instructions for setting up NNTPChan with Redis as the data-storage system.

##Install Redis

Redis 3.x or higher is required, [stable release](http://download.redis.io/releases/redis-stable.tar.gz) recommended

* See http://redis.io/download

##Configuration

In srnd.ini the database sections should look like this:

    [database]
    type=redis
    schema=single
    host=localhost
    port=6379
    user=
    password=

##Next step

See the [Running NNTPChan](running.md).
