#!/usr/bin/env bash
#
# this script can be called via hooks for each post to purge varnish cache on new post
#

if [ "$3" == "" ] ; then
		op="$2"
else
		op="$3"
fi

# ip to bind to when doing http request
ip="127.0.1.1"

varnish="127.0.0.1:8000"

# purge thread page
curl --interface "$ip" -v -X PURGE http://$varnish/thread-$(sha1sum <<< "$op" | cut -d' ' -f1).html
# purge board page
curl --interface "$ip" -v -X PURGE http://$varnish/$1-0.html
