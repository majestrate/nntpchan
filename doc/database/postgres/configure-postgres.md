Configuring Postgres database
=============================

These are instructions for setting up NNTPChan with Postgres as the data-storage system.

##Configuring Postgres

Setting up postgres (as root):

    # become postgres user
    su postgres
    # spawn postgres admin shell
    psql 

You'll get a prompt, enter the following:

    CREATE ROLE srnd WITH LOGIN PASSWORD 'srnd';
    CREATE DATABASE srnd WITH ENCODING 'UTF8' OWNER srnd;
    \q

For demo purposes we'll use these credentials.
These are default values, please change them later.

###Important

These credentials assume you are going to run using a user called `srnd`, if your username you plan to run the daemon as is different please change `srnd` to your username.
