#!/usr/bin/env bash

neochan="no"
if [ "$1" == "--enable-neochan" ] ; then
    neochan="yes"
fi

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

outfile="$PWD/contrib/static/nntpchan.js"
neochan_js_outfile="$PWD/contrib/static/neochan.js"
neochan_css_outfile="$PWD/contrib/static/neochan.css"

mini() {
    echo "minify $1"
    echo "" >> $2
    echo "/* begin $1 */" >> $2
    "$GOPATH/bin/minify" --mime=text/javascript >> $2 < $1
    echo "" >>  $2
    echo "/* end $1 */" >> $2
}

css() {
    echo "minify $1"
    echo "" >> $2
    echo "/* begin $1 */" >> $2
    lessc $1 >> $2 
    echo "" >>  $2
    echo "/* end $1 */" >> $2
}

initfile() {
    
    rm -f "$1"

    echo '/*' >> "$1"
    echo ' * For source code and license information please check https://github.com/majestrate/nntpchan' >> "$1"
    brandingfile=./contrib/branding.txt
    if [ -e "$brandingfile" ] ; then
        echo ' *' >> "$1"
        while read line; do
            echo -n ' * ' >> "$1";
            echo $line >> "$1";
        done < $brandingfile;
    fi
    echo ' */' >> "$1"
}
echo
echo "building nntpchan.js ..."
echo
initfile "$outfile"


for f in ./contrib/js/contrib/*.js ; do
    mini "$f" "$outfile"
done

mini ./contrib/js/entry.js "$outfile"

# local js
for f in ./contrib/js/nntpchan/*.js ; do
  mini "$f" "$outfile"
done

# vendor js
for f in ./contrib/js/vendor/*.js ; do
  mini "$f" "$outfile"
done

if [ "$neochan" == "yes" ] ; then
    set +e
    for exe in lessc coffee ; do
        which $exe &> /dev/null
        if [ "$?" != "0" ] ; then
            echo "$exe not installed";
            exit 1
        fi
    done
    
    echo
    echo "building neochan.js ..."
    echo
    
    initfile "$neochan_js_outfile"
    for f in ./contrib/js/neochan/*.coffee ; do
        echo "compile $f"
        coffee -cs < "$f" > "$f.js"
    done
    for f in ./contrib/js/neochan/*.js ; do
        mini "$f" "$neochan_js_outfile"
    done

    echo
    echo "building neochan.css ..."
    echo
    initfile "$neochan_css_outfile"
    for f in ./contrib/js/neochan/*.less ; do
        css "$f" "$neochan_css_outfile"
    done

fi
echo
echo "ok"
