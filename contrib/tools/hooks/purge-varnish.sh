#!/usr/bin/env bash
#
# this script can be called via hooks for each post to purge varnish cache on new post
#

# ip to bind to when doing http request
ip="127.0.1.1"

varnish="127.0.0.1:8000"

# purge thread page
curl --interface "$ip" -X PURGE http://$varnish/thread-$(echo -n "$3" | sha1sum | cut -d' ' -f1).html &> /dev/null
# purge board page
curl --interface "$ip" -X PURGE http://$varnish/$1-0.html &> /dev/null
