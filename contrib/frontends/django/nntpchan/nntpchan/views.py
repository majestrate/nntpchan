from django.conf import settings
from django.http import HttpResponse, HttpResponseNotAllowed, JsonResponse
from django.views.decorators.csrf import csrf_exempt   

from .frontend.models import Post, Attachment, Newsgroup
from .frontend import util

from . import thumbnail

import email
import traceback
from datetime import datetime
import mimetypes
import os

@csrf_exempt
def webhook(request):
    """
    endpoint for nntpchan daemon webhook
    """
    if request.method != 'POST':
        return HttpResponseNotAllowed(['POST'])
    
    try:
        msg = email.message_from_bytes(request.body)
        newsgroup = msg.get('Newsgroups')

        if newsgroup is None:
            raise Exception("no newsgroup specified")
        
        if not util.newsgroup_valid(newsgroup):
            raise Exception("invalid newsgroup name")
        
    
        group, created = Newsgroup.objects.get_or_create(name=newsgroup)

        if group.banned:
            raise Exception("newsgroup is banned")

        msgid = None
        for h in ('Message-ID', 'Message-Id', 'MessageId', 'MessageID'):
            if h in msg:
                msgid = msg[h]
                break

        if msgid is None:
            raise Exception("no message id specified")
        elif not util.msgid_valid(msgid):
            raise Exception("invalid message id format: {}".format(msgid))
        
        h = util.hashid(msgid)
        atts = list()
        ref = msg['References'] or ''
        posted = email.utils.parsedate_to_datetime(msg['Date'])

        f = msg['From'] or 'anon <anon@anon>'
        name = email.utils.parseaddr(f)[0]
        post, created = Post.objects.get_or_create(defaults={
            'posthash': h,
            'reference': ref,
            'posted': posted,
            'name': name,
            'subject': msg["Subject"] or '',
            'newsgroup': group}, msgid=msgid)
        m = ''

        for part in msg.walk():
            ctype = part.get_content_type()
            if ctype.startswith("text/plain"):
                m += '{} '.format(part.get_payload(decode=True).decode('utf-8'))
            else:
                print(part.get_content_type())
                payload = part.get_payload(decode=True)
                if payload is None:
                    continue
                mtype = part.get_content_type()
                ext = mimetypes.guess_extension(mtype) or ''
                fh = util.hashfile(bytes(payload))
                fn = fh + ext
                fname = os.path.join(settings.MEDIA_ROOT, fn)
                if not os.path.exists(fname):
                    with open(fname, 'wb') as f:
                        f.write(payload)
                tname = os.path.join(settings.MEDIA_ROOT, 'thumb-{}.jpg'.format(fn))
                placeholder = os.path.join(settings.ASSETS_ROOT, 'placeholder.jpg')
                if not os.path.exists(tname):
                    thumbnail.generate(fname, tname, placeholder)
                        
                att = Attachment(filehash=fh)
                att.mimetype = mtype
                att.filename = part.get_filename()
                att.save()
                atts.append(att)
        post.message = m
        post.save()

        for att in atts:
            post.attachments.add(att)
    except Exception as ex:
        traceback.print_exc()
        return JsonResponse({ 'error': '{}'.format(ex) })
    else:
        return JsonResponse({'posted': True})
