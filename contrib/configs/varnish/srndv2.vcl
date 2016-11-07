vcl 4.0;

backend default { 
    .host = "127.0.0.1";
    .port = "18000";
}
 

acl purge {
  # ACL we'll use later to allow purges
  "127.0.1.1";
}

sub vcl_miss {
        return (fetch) ;
}

sub vcl_recv {
        if (req.method == "PURGE") {
                if (!client.ip ~ purge) { # purge is the ACL defined at the begining
                        # Not from an allowed IP? Then die with an error.
                        return (synth(405, "This IP is not allowed to send PURGE requests."));
                }
                return (purge);
        }

        if (req.method != "GET" &&
                req.method != "HEAD" &&
                req.method != "PUT" &&
                req.method != "POST" &&
                req.method != "TRACE" &&
                req.method != "OPTIONS" &&
                req.method != "PATCH" &&
                req.method != "DELETE") {
                /* Non-RFC2616 or CONNECT which is weird. */
                return (pipe);
        }
        # Strip hash, server doesn't need it.
        if (req.url ~ "\#") {
                set req.url = regsub(req.url, "\#.*$", "");
        }
        # Strip a trailing ? if it exists
        if (req.url ~ "\?$") {
                set req.url = regsub(req.url, "\?$", "");
        }
        return (hash);
}

sub vcl_pipe {
        return (pipe);
}

sub vcl_hit {
        return (deliver);
}

sub vcl_backend_response {
        if (beresp.status == 500 || beresp.status == 502 || beresp.status == 503 || beresp.status == 504) {
                return (abandon);
        }
        if (bereq.url == "/") {
                set beresp.ttl = 60s;
        } else {
                if (bereq.url == "/ukko.html") {
                        set beresp.ttl = 5s;
                } else {
                        set beresp.ttl = 10s;
                }
        }
        set beresp.grace = 1h;
        return (deliver);
}

sub vcl_deliver {
        if (obj.hits > 0) {
                set resp.http.X-Cache = "HIT";
        } else {
                set resp.http.X-Cache = "MISS";
        }
        set resp.http.X-Cache-Hits = obj.hits;
        return (deliver);
}
