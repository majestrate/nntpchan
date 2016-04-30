#!/usr/bin/env bash
set -e
root=$(readlink -e $(dirname $0))

cd $root
if [ -z "$GOPATH" ]; then
	export GOPATH=$PWD/go
	mkdir -p $GOPATH
fi

if [ ! -f $GOPATH/bin/minify ]; then
  echo "set up minifiy"  
	go get -v github.com/tdewolff/minify/cmd/minify
fi

echo -e "//For source code and license information please check https://github.com/majestrate/nntpchan \n" > ./contrib/static/nntpchan.js

echo -n "minify "
cat ./contrib/js/main.js_ | $GOPATH/bin/minify --mime=text/javascript >> ./contrib/static/nntpchan.js
for f in ./contrib/js/*.js ; do
  echo -n "."
	cat $f | $GOPATH/bin/minify --mime=text/javascript >> ./contrib/static/nntpchan.js
done
echo "ok"
