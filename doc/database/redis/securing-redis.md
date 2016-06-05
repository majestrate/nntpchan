Securing Redis
==============

This document provides a good tip for securing your Redis server, just to be 100% happy with the security.

##Adding an authentication password for commands

This will allow you to add a password to your Redis server that must be used before any other commands can be issued to your Redis server.

* Remember choose a strong password with lower-case, upper-case, numbers and other symbols.
* Make sure there are no spaces. (need to still test this #6969)

Then take your password, `x` and run this command (with `sudo` if needed).

    echo "requirepass x" >> /path/to/your/redis.conf
