#!/usr/bin/env bash
root=$(readlink -e "$(dirname "$0")")
set -e
if [ "x" == "x$root" ] ; then
    root=$PWD/${0##*}
fi
cd "$root"

if [ -z "$GOPATH" ]; then
	export GOPATH=$root/go
	mkdir -p "$GOPATH"
fi

if [ ! -f "$GOPATH/bin/minify" ]; then
  echo "set up minifiy"
	go get -v github.com/tdewolff/minify/cmd/minify
fi

outfile=$PWD/contrib/static/nntpchan.js

lint() {
    if [ "$(which jslint)" == "" ] ; then
        # no jslint
        true
    else
        echo "jslint: $1"
        jslint --browser "$1"
    fi
}

mini() {
    echo "minify $1"
    echo "" >> $2
    echo "/* begin $1 */" >> $2
    "$GOPATH/bin/minify" --mime=text/javascript >> $2 < $1
    echo "/* end $1 */" >> $2
}

# do linting too
if [ "x$1" == "xlint" ] ; then
    echo "linting..."
    for f in ./contrib/js/*.js ; do
        lint "$f"
    done
fi

rm -f "$outfile"

echo '/*' >> $outfile
echo ' * For source code and license information please check https://github.com/majestrate/nntpchan' >> $outfile
brandingfile=./contrib/branding.txt
if [ -e "$brandingfile" ] ; then
    echo ' *' >> $outfile
    while read line; do
        echo -n ' * ' >> $outfile;
        echo $line >> $outfile;
    done < $brandingfile;
fi
echo ' */' >> $outfile

if [ -e ./contrib/js/contrib/*.js ] ; then
    for f in ./contrib/js/contrib/*.js ; do
        mini "$f" "$outfile"
    done
fi

mini ./contrib/js/entry.js "$outfile"

# local js
for f in ./contrib/js/nntpchan/*.js ; do
  mini "$f" "$outfile"
done

# vendor js
for f in ./contrib/js/vendor/*.js ; do
  mini "$f" "$outfile"
done

echo "ok"
