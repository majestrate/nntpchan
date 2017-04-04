#How to install nntpchan on Debian 8.5 Jessie

Install the initial dependencies:

```
apt-get -y --no-install-recommends install imagemagick libsodium-dev sox git ca-certificates \
libav-tools build-essential tcl8.5 postgresql postgresql-contrib
```

##Configure postgresql

```
su - postgres -c "createuser --pwprompt --encrypted srnd"
su - postgres -c "createdb srnd"
```
Don't forget the password you make for the srnd user, you will need it for configuration.
## Install golang

Download the golang tarball, extract it to `/usr/local`, and add it to the global profile:

```
cd /opt
wget https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz
tar -C /usr/local/ -xvzf go1.8.linux-amd64.tar.gz
echo 'export PATH="$PATH:/usr/local/go/bin"' >> /etc/profile
```

Your `PATH` is set at login, so log out and back in before proceeding.

## Install nntpchan

```
cd /opt
git clone https://github.com/majestrate/nntpchan.git
cd nntpchan
make
```

Now you can proceed with [setting up NNTPChan](setting-up.md). When you get to the "set paths to external programs" step, you should change the ffmpeg path to `/usr/bin/avconv`.

Run `./srndv2 setup` and follow the instructions [here](setting-up.md).
