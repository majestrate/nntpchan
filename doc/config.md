Configuring NNTPChan server
===========================

This document provides a step-by-step guide to configurin your NNTPChan node.

##Configuring via web-interface

You can configure NNTPChan via the web-interface by navigating your browser to http://127.0.0.1:18000.

###Selecting yoru data-storage system

![Image 1](http://i.imgur.com/l9iiXxB.png)

First your will be asked what data-storage system you would like to use. We support Redis and PostgreSQL.

<hr>

####Redis configuration

![Image 2](http://i.imgur.com/HDp4Ddf.png)

If you have chosen Redis then fill in the fields below:

* **Hostname or IP Address** - This is the hostname or IP address of your Redis server (I would run it locally to be safe).
* **Port** - The port that your Redis server is running on.
* **Password** - Optional authentication password for Redis ([Setting up a Redis password](securing-redis.md)).
