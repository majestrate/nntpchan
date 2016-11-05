from django.db import models

from . import util

import mimetypes

class Attachment(models.Model):
    """
    a file attachment assiciated with a post
    a post may have many attachments
    """
    filehash = models.CharField(max_length=256, editable=False)
    filename = models.CharField(max_length=256)
    mimetype = models.CharField(max_length=256, default='text/plain')
    width = models.IntegerField(default=0)
    height = models.IntegerField(default=0)
    banned = models.BooleanField(default=False)

    def path(self):
        ext = self.filename.split('.')[-1]
        return '{}.{}'.format(self.filehash, ext)
    
    def thumb(self):
        return '/media/thumb-{}.jpg'.format(self.path())

    def source(self):
        return '/media/{}'.format(self.path())
    
    
class Newsgroup(models.Model):
    """
    synonym for board
    """
    name = models.CharField(max_length=256, primary_key=True, editable=False)
    posts_per_page = models.IntegerField(default=10)
    max_pages = models.IntegerField(default=10)
    banned = models.BooleanField(default=False)
    
    def get_absolute_url(self):
        from django.urls import reverse
        return reverse('nntpchan.frontend.views.boardpage', args=[self.name, '0'])
    
class Post(models.Model):
    """
    a post made
    """
    
    msgid = models.CharField(max_length=256, primary_key=True, editable=False)
    posthash = models.CharField(max_length=256, editable=False)
    reference = models.CharField(max_length=256, default='')
    message = models.TextField(default='')
    subject = models.CharField(max_length=256, default='None')
    name = models.CharField(max_length=256, default='Anonymous')
    pubkey = models.CharField(max_length=64, default='')
    signature = models.CharField(max_length=64, default='')
    newsgroup = models.ForeignKey(Newsgroup, on_delete=models.CASCADE)
    attachments = models.ManyToManyField(Attachment)
    posted = models.DateTimeField()
    placeholder = models.BooleanField(default=False)
    
    def get_all_replies(self):
        if self.is_op():
            return Post.objects.filter(reference=self.msgid).order_by('posted')
    
    def get_board_replies(self, truncate=5):
        rpls = self.get_all_replies()
        l = len(rpls)
        if l > truncate:
            rpls = rpls[l-truncate:]
        return rpls
        
    def is_op(self):
        return self.reference == ''

    def shorthash(self):
        return self.posthash[:10]
    
    def get_absolute_url(self):
        if self.is_op():
            op = util.hashid(self.msgid)
            return '/t/{}/'.format(op)
        else:
            op = util.hashid(self.reference)
            frag = util.hashid(self.msgid)
            return '/t/{}/#{}'.format(op, frag)
