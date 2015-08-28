
Postgres on Debian:


    # install
    apt install postgresql postgresql-client


Setting up postgres (as root)

    # become postgres user
    su postgres
    # spawn postgres admin shell
    psql

You'll get a prompt, enter the following:

    CREATE ROLE srnd WITH LOGIN PASSWORD 'srndpassword';
    CREATE DATABASE srnd WITH ENCODING 'UTF8' OWNER srnd;
    \q

Change the username and password as desired.

RabbitMQ on Debian:

    # install
    apt install rabbitmq-server

Copy the rabbitmq configs and restart rabbitmq

    # as root
    cp ~/nntpchan/contrib/configs/rabbitmq/* /etc/rabbitmq/
    systemctl restart rabbitmq-server
