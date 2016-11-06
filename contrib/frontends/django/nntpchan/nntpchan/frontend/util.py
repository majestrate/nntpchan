import base64
import hashlib
import re

import time

def hashid(msgid):
    h = hashlib.sha1()
    m = '{}'.format(msgid).encode('ascii')
    h.update(m)
    return h.hexdigest()

def newsgroup_valid(name):
    return re.match('overchan\.[a-zA-Z0-9\.]+[a-zA-Z0-9]$', name) is not None or name == 'ctl'

def hashfile(data):
    h = hashlib.sha512()
    h.update(data)
    return base64.b32encode(h.digest()).decode('ascii')

def msgid_valid(msgid):
    return re.match("<[a-zA-Z0-9\$\._\-\|]+@[a-zA-Z0-9\$\._\-\|]+>$", msgid) is not None

def time_int(dtime):
    return time.mktime(dtime.timetuple())
    
