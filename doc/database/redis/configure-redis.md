Configuring Redis database
==========================

These are instructions for setting up NNTPChan with Redis as the data-storage system.

##Configuring Redis

In `srnd.ini` the database sections should look like this:

    [database]
    type=redis
    schema=single
    host=localhost
    port=6379
    user=
    password=

##Securing Redis (optional)

Read [Securing Redis](securing-redis.md) for adding password authentication to your Redis server.
