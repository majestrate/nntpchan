import hashlib


def hashid(msgid):
    return hashlib.sha1().hexdigest('%s' % msgid).decode('ascii')
