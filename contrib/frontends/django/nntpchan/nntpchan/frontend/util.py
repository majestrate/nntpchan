from django.conf import settings

import base64
import hashlib
import re

from datetime import datetime
import time
import string
import random
import nntplib
import email.message

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
    
def randstr(l, base=string.digits):
    r = ''
    while l > 0:
        r += random.choice(base)
        l -= 1
    return r


def createPost(newsgroup, ref, form, files):
    """
    create a post and post it to a news server
    """

    msg = email.message.Message()
    msg['Content-Type'] = 'multipart/mixed'
    msg["Subject"] = form["subject"] or "None"
    msg['Date'] = email.utils.format_datetime(datetime.now())
    if ref and not msgid_valid(ref):
        return None, "invalid reference: {}".format(ref)
    if ref:
        msg["References"] = ref
    msg["Newsgroups"] = newsgroup
    msg["From"] = '{} <anon@django.nntpchan.tld>'.format(form['name'] or 'Anonymous')
    if 'attachment' in files:
        f = files['attachment']
        part =  email.message.Message()
        part['Content-Type'] = f.content_type
        part['Content-Disposition'] = 'form-data; filename="{}"; name="attachment"'.format(f.name)
        part['Content-Transfer-Encoding'] = 'base64'
        part.set_payload(base64.b64encode(f.read()))
        msg.attach(part)
    text = email.message.Message()
    m = '{}'.format(form['message'] or ' ')
    text.set_payload(m)
    text['Content-Type'] = 'text/plain'
    msg.attach(text)

    server = settings.NNTP_SERVER
    server['readermode'] = True
    response = None
    try:
        with nntplib.NNTP(**server) as nntp:
            nntp.login(**settings.NNTP_LOGIN)
            response = nntp.post(msg.as_bytes())
    except Exception as e:
        return None, 'connection to backend failed, {}'.format(e)
    if ref:
        return ref, None
    return None, None
