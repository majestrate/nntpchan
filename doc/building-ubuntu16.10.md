Install the initial dependencies:

    # apt-get -y --no-install-recommends install imagemagick sox git ca-certificates \
    ffmpeg build-essential tcl8.5 postgresql postgresql-contrib golang-go

Configure PostgreSQL:

    # su - postgres -c "createuser --pwprompt --encrypted srnd"
    # su - postgres -c "createdb srnd"

Install nntpchan:

    # cd /opt
    # git clone https://github.com/majestrate/nntpchan.git
    # cd nntpchan
    # make
