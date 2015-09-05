#!/usr/bin/env bash
#
# shell script for regenerating thumbnails
#

if [ "$1" == "" ] ; then
    echo "usage: $0 webroot_dir"
else
    cd $1
    echo "regenerate missing thumbs in $(pwd)"
    find img/ \
         -type f \
         -regextype posix-extended \
         -iregex '.*\.(png|jpg|gif)$' \
         -not -execdir test -f '../thm/{}' \; \
         -exec echo 'generating missing thumb for {}' \; \
         -exec mogrify \
         -define jpeg:size=500x500 \
         -thumbnail '250>x250>' \
         -path 'thm/{}.jpg' \
         -strip \
         '{}' \;
fi
