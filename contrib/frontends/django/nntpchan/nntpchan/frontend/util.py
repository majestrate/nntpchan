from django.conf import settings

import base64
import hashlib
import re

import nacl.signing
from binascii import hexlify

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
    return int(time.mktime(dtime.timetuple()))
    
def randstr(l, base=string.digits):
    r = ''
    while l > 0:
        r += random.choice(base)
        l -= 1
    return r


def createPost(newsgroup, ref, form, files, secretKey=None):
    """
    create a post and post it to a news server
    """

    msg = email.message.Message()
    if 'subject' in form:
        msg["Subject"] = form["subject"] or "None"
    else:
        msg["Subject"] = "None"
    msg['Date'] = email.utils.format_datetime(datetime.now())
    if ref and not msgid_valid(ref):
        return None, "invalid reference: {}".format(ref)
    if ref:
        msg["References"] = ref
    msg["Newsgroups"] = newsgroup
    name = "Anonymous"
    if 'name' in form:
        name = form['name'] or name
    msg["From"] = '{} <anon@django.nntpchan.tld>'.format(name)
    if 'attachment' in files:
        msg['Content-Type'] = 'multipart/mixed'
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
    else:
        msg['Content-Type'] = 'text/plain; charset=UTF-8'
        m = '{}'.format(form['message'] or ' ')
        msg.set_payload(m)
    msg['Message-Id'] = '<{}${}@signed.{}>'.format(randstr(5), int(time_int(datetime.now())), settings.FRONTEND_NAME)
    if secretKey:
        msg['Path'] = settings.FRONTEND_NAME
        # sign
        keypair = nacl.signing.SigningKey(secretKey, nacl.signing.encoding.HexEncoder)
        pubkey = hexlify(keypair.verify_key.encode()).decode('ascii')
        outerMsg = email.message.Message()
        h = hashlib.sha512()
        body = msg.as_bytes()
        h.update(body)
        sig = hexlify(keypair.sign(h.digest()).signature).decode('ascii')
        data = '''Content-Type: message/rfc822; charset=UTF-8
Message-ID: {}
Content-Transfer-Encoding: 8bit
Newsgroups: {}
X-Pubkey-Ed25519: {}
X-Signature-Ed25519-Sha512: {}
From: {}
Date: {}
Subject: {}

{}'''.format(msg["Message-ID"], newsgroup, pubkey, sig, msg["From"], msg["Date"], msg['Subject'], msg.as_string())
        data = data.encode('utf-8')
    else:
        data = msg.as_bytes()
    server = settings.NNTP_SERVER
    server['readermode'] = True
    response = None
    try:
        with nntplib.NNTP(**server) as nntp:
            nntp.login(**settings.NNTP_LOGIN)
            response = nntp.ihave(msg['Message-ID'], data)

    except Exception as e:
        raise e
        return None, 'connection to backend failed, {}'.format(e)
    if ref:
        return ref, None
    return None, None
