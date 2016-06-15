#How to install nntpchan on Debian 8.5 Jessie

Install the initial dependencies:

```
apt-get -y --no-install-recommends install imagemagick libsodium-dev sox git ca-certificates libav-tools build-essential tcl8.5
```

##Install redis

It is not recommended that you install redis from the default package repos because it is probably not up to date.

Download the redis stable tarball and make:

```
cd /opt
wget http://download.redis.io/redis-stable.tar.gz
tar -xzvf redis-stable.tar.gz
cd redis-stable
make && make test && make install
```

The `utils/` directory has a bash script that automates redis configuration. The default settings work just fine, so run the script:

```
cd utils && ./install_server.sh
```

Make redis start during system boot up:

```
update-rc.d redis_6379 defaults
```

It is *strongly recommended* that you use a password for redis. I am generating an sha512sum for a random string:

```
"good old fashioned memes will end global warming and restore our freedom of speech" | sha512sum
```

Edit `/etc/redis/6379.conf` and append the file with `requirepass YOUR_LONG_PASSWORD_HERE`.

## Install golang

Download the golang tarball, extract it to `/usr/local`, and add it to the global profile:

```
cd /opt
wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
tar -C /usr/local/ -xvzf go1.6.2.linux-amd64.tar.gz
echo 'export PATH="$PATH:/usr/local/go/bin"' >> /etc/profile
```

Your `PATH` is set at login, so log out and back in before proceeding. 

## Install nntpchan

```
cd /opt
git clone https://github.com/majestrate/nntpchan.git
cd nntpchan
./build.sh
```

Now you can proceed with [setting up NNTPChan](setting-up.md). When you get to the "set paths to external programs" step, you should change the ffmpeg path to `/usr/bin/avconv`.

Run `./srndv2 setup` and follow the instructions [here](setting-up.md).
