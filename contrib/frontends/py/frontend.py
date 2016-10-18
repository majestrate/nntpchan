#!/usr/bin/env python3
from nntpchan import message
from nntpchan import db
import logging
import os
import sys


if __name__ == "__main__":
    lvl = logging.INFO
    if 'NNTPCHAN_DEBUG' in os.environ:
        lvl = logging.DEBUG
    logging.basicConfig(level=lvl)
    l = logging.getLogger(__name__)
    cmd = sys.argv[1]
    if cmd == 'post':
        fpath = sys.argv[2]
        msg = None
        if not os.path.exists(fpath):
            print("{} does not exist".format(fpath))
            exit(1)
        with open(fpath) as f:
            msg = message.parse(f)
        if msg:
            l.debug("loaded {}".format(fpath))
    elif cmd == 'newsgroup':
        if db.allowsNewsgroup(sys.argv[2]):
            exit(0)
        else:
            exit(1)
    elif cmd == 'msgid':
        if db.allowsMessage(sys.argv[2]):
            exit(0)
        else:
            exit(1)
    elif cmd == 'init':
        db.init()
