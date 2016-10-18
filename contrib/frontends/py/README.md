# python nntpchan demo frontend #

## usage ##

### srndv2 unstable ###

add to nntpchan.json hooks section:

    {
      "name": "pyfront",
      "exec": "/path/to/frontend.py"
    }

### nntpd ###

add this to nntpchan.ini


    ...
    [frontend]
    type=exec
    exec=/path/to/frontend.py
