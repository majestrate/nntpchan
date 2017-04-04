Configuring Postgres database
=============================

These are instructions for setting up NNTPChan with Postgres as the data-storage system.

## Configuring Postgres
A user with sufficient privileges to run su is required (hint: you can use root). This command switches to the Postgres user, creates a Postgres role called `srnd`, and prompts for a password. For illustrative purposes, we will use `srnd` as the password.

    # su - postgres -c "createuser --pwprompt --createdb --encrypted srnd"

### Important

It's easiest to connect to Postgres using role-based authentication. In this case, our Linux user `srnd` matches up with our Postgres role `srnd`, so role-based authentication can take place. If you're running SRNDv2 as a different user (e.g. `nntpchan`), you will need to create a role that matches that user using the command above.
