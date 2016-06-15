# Debain 8.5 step-by-step

Everything in this guide is done as root.

## FFMPEG install
FFMPEG is non-free or something. There is a guide on how to get it installed [here](https://www.assetbank.co.uk/support/documentation/install/ffmpeg-debian-squeeze/ffmpeg-debian-jessie/) or you can just follow these instructions:

Create a new list for ffmpeg:

```
# touch /etc/apt/sources.list.d/ffmpeg.list
```

And add the repos:

```
# echo "deb http://www.deb-multimedia.org jessie main non-free" >> /etc/apt/sources.list.d/ffmpeg.list
# echo "deb-src http://www.deb-multimedia.org jessie main non-free" >> /etc/apt/sources.list.d/ffmpeg.list
```

Update the repos:

```
# apt-get update
```

Get the key ring:

```
# apt-get -y install deb-multimedia-keyring
```

Update the repos, again:

```
# apt-get update
```

Make sure ffmpeg is not already installed:

```
# apt-get remove ffmpeg
```

Get the library packages and build tools:

```
# apt-get install -y build-essential libmp3lame-dev libvorbis-dev libtheora-dev libspeex-dev yasm pkg-config libfaac-dev libopenjpeg-dev libx264-dev
```


Make some temporary folders we will use to organize the ffmpeg stuff:

```
# mkdir -p /opt/ffmpeg-build/{software,src} && cd /opt/ffmpeg-build/software
```

Download the ffmpeg ball:

```
# wget http://ffmpeg.org/releases/ffmpeg-2.7.2.tar.bz2
```

Extract the ball:

```
# tar -C ../src -xvjf ffmpeg-2.7.2.tar.bz2
```

Change to the source directory:

```
# cd ../src/ffmpeg-2.7.2
```

Configure:

```
# ./configure --enable-gpl --enable-postproc --enable-swscale --enable-avfilter --enable-libmp3lame --enable-libvorbis --enable-libtheora --enable-libx264 --enable-libspeex --enable-shared --enable-pthreads --enable-libopenjpeg --enable-libfaac --enable-nonfree
```

Make:

```
# make && make install
```

Do this thing:

```
# /sbin/ldconfig
```

# Install redis

If you use your package manager, you will install an old version of redis and you will wonder why nothing works. The full set of instructions is [here](http://redis.io/topics/quickstart).

Let's make a build directory:

```
# mkdir /opt/redis-build && cd /opt/redis-build
```

Download the stable ball:

```
# wget http://download.redis.io/redis-stable.tar.gz
```

Make:

```
# make
```

Install tcl, so we can make test:

```
# apt-get install -y install tcl
```

Make test:

```
# make test
```

Then install:

```
# make install
```

Now, run the install script. Select the defaults unless you want to diverge from this step-by-step:

```
# cd utils && ./install_server.sh
```

Make redis start on boot:

```
# update-rc.d redis_6379 defaults
```

__It is strongly recommended that you secure redis:__

Generate a cute little sha512sum to use as the password. Use your own sufficiently good seed.

```
# echo "requirepass" $(echo "the ass was fat and i loved the way she cooks her memes" | sha512sum) >> /etc/redis/6379.conf
```

Restart redis:

```
# service redis_6379 restart
```

# Install golang

[Get the golang binary](https://golang.org/doc/install). Make sure you get an up-to-date version. As of 2016-06-10, the stable version of golang is 1.6.2.

Download the ball:

```
# wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
```

Extract the ball:

```
# tar -C /usr/local/ -xvzf go1.6.2.linux-amd64.tar.gz
```

Add the go path to the global profile:

```
# echo 'export PATH="$PATH:/usr/local/go/bin"' >> /etc/.profile
```

# Now install nntpchan

Do some things:

```
# sudo apt-get -y --no-install-recommends install imagemagick libsodium-dev sox git ca-certificates
```

Clone the repo in /opt/nntpchan:

```
# cd /opt && git clone https://github.com/majestrate/nntpchan.git
# cd /opt/nntpchan
```

Build:

```
# ./build.sh
```

Run the setup:

```
# ./srndv2 setup
```

Follow the instructions. The ffmpeg command should be changed to `/usr/local/bin/ffmpeg`.

Have fun.
