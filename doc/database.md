
Postgres on Debian:


    # install
    apt-get install postgresql postgresql-client


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
The username has to be the same as the database name.
