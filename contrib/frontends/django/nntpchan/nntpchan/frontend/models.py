from django.db import models

from . import util

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
    
    def is_op(self):
        return self.reference is None
    
    def get_absolute_url(self):
        from django.urls import reverse
        
        if self.is_op():
            op = util.hashid(self.msgid)
            return reverse('nntpchan.frontend.views.threadpage', args[op])
        else:
            op = util.hashid(self.reference.msgid)
            frag = util.hashid(self.msgid)
            return reverse('nntpchan.frontend.views.threadpage', args=[op]) + '#{}'.format(frag)
