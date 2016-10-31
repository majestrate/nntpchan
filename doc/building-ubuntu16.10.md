Install the initial dependencies:

    # apt-get -y --no-install-recommends install imagemagick libsodium-dev sox git ca-certificates \
    libav-tools build-essential tcl8.5 postgresql postgresql-contrib golang-go
    
Configure PostgreSQL:

    # su - postgres -c "createuser --pwprompt --encrypted srnd"
    # su - postgres -c "createdb srnd"

Install nntpchan:

    # cd /opt
    # git clone https://github.com/majestrate/nntpchan.git
    # cd nntpchan
    # ./build.sh
