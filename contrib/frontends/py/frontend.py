#!/usr/bin/env python3
import nntpchan
import sys

if __name__ == "__main__":
    msgid = sys.argv[1]
    group = sys.argv[2]
    if nntpchan.addArticle(msgid, group):
        nntpchan.regenerate(msgid, group)
