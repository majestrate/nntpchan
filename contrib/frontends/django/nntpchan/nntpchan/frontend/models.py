
from django.conf import settings
from django.db import models
from django.core.urlresolvers import reverse

from . import util

import mimetypes
from datetime import datetime

import os

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
    
    def thumb(self, root=settings.MEDIA_URL):
        return '{}thumb-{}.jpg'.format(root, self.path())

    def source(self, root=settings.MEDIA_URL):
        return '{}{}'.format(root, self.path())

    def remove(self):
        """
        remove from filesystem and delete self
        """
        os.unlink(os.path.join(settings.MEDIA_ROOT, self.thumb('')))
        os.unlink(os.path.join(settings.MEDIA_ROOT, self.source('')))
        self.delete()
    
class Newsgroup(models.Model):
    """
    synonym for board
    """
    name = models.CharField(max_length=256, primary_key=True, editable=False)
    posts_per_page = models.IntegerField(default=10)
    max_pages = models.IntegerField(default=10)
    banned = models.BooleanField(default=False)

    def get_absolute_url(self):
        if self.name == 'ctl':
            return reverse('frontend:modlog')
        return reverse('frontend:board-front', args=[self.name[9:]])
    
class Post(models.Model):
    """
    a post made anywhere on the boards
    """
    
    msgid = models.CharField(max_length=256, primary_key=True, editable=False)
    posthash = models.CharField(max_length=256, editable=False)
    reference = models.CharField(max_length=256, default='')
    message = models.TextField(default='')
    subject = models.CharField(max_length=256, default='None')
    name = models.CharField(max_length=256, default='Anonymous')
    pubkey = models.CharField(max_length=64, default='')
    signature = models.CharField(max_length=64, default='')
    newsgroup = models.ForeignKey(Newsgroup)
    attachments = models.ManyToManyField(Attachment)
    posted = models.IntegerField(default=0)
    placeholder = models.BooleanField(default=False)
    last_bumped = models.IntegerField(default=0)

    def has_attachment(self, filehash):
        """
        return True if we own a file attachment by its hash
        """
        for att in self.attachments.all():
            if att.filehash in filehash:
                return True
        return False
    
    def get_all_replies(self):
        """
        get all replies to this thread
        """
        if self.is_op():
            return Post.objects.filter(reference=self.msgid).order_by('posted')
        
    def get_board_replies(self, truncate=5):
        """
        get replies to this thread
        truncate to last N replies
        """
        rpls = self.get_all_replies()
        l = len(rpls)
        if l > truncate:
            rpls = rpls[l-truncate:]
        return rpls
        
    def is_op(self):
        return self.reference == '' or self.reference == self.msgid

    def shorthash(self):
        return self.posthash[:10]

    def postdate(self):
        return datetime.fromtimestamp(self.posted)
    
    def get_absolute_url(self):
        """
        self explainitory
        """
        if self.is_op():
            op = util.hashid(self.msgid)
            return reverse('frontend:thread', args=[op])
        else:
            op = util.hashid(self.reference)
            frag = util.hashid(self.msgid)
            return reverse('frontend:thread', args=[op]) + '#{}'.format(frag)

    def bump(self, last):
        """
        bump thread
        """
        if self.is_op():
            self.last_bumped = last

    def remove(self):
        """
        remove post and all attachments
        """
        for att in self.attachments.all():
            att.remove()
        self.delete()

class ModPriv(models.Model):
    """
    a record that permits moderation actions on certain boards or globally
    """

    """
    absolute power :^DDDDDDD (does not exist)
    """
    GOD = 0

    """
    node admin
    """
    ADMIN = 1

    """
    can ban, delete and edit posts
    """
    MOD = 2
    
    """
    can only delete
    """
    JANITOR = 3

    """
    lowest access level for login 
    """
    LOWEST = JANITOR
    
    """
    what board this priviledge is for or 'all' for global
    """
    board = models.CharField(max_length=128, default='all')

    """
    what level of priviledge is granted
    """
    level = models.IntegerField(default=3)

    """
    public key of mod mod user
    """
    pubkey = models.CharField(max_length=256, editable=False)

    @staticmethod
    def has_access(level, pubkey, board_name=None):
        # check global priviledge
        global_priv = ModPriv.objects.filter(pubkey=pubkey, board='all')
        for priv in global_priv:
            if priv.level <= level:
                return True
        # check board level priviledge
        if board_name:
            board_priv = ModPriv.objects.filter(pubkey=pubkey, board=board_name)
            for priv in board_priv:
                if priv.level <= level:
                    return True
        # no access allowed
        return False

    @staticmethod
    def try_delete(pubkey, post):
        """
        try deleting a post, return True if it was deleted otherwise return False
        """
        if ModPriv.has_access(ModPriv.JANITOR, pubkey, post.newsgroup.name):
            # we can do it
            post.remove()
            return True
        return False

    @staticmethod
    def try_edit(pubkey, post, newbody):
        """
        try editing a post by replacing its body with a new one
        returns True if this was done otherwise return False
        """
        if ModPriv.has_access(ModPriv.MOD, pubkey, post.newsgroup.name):
            post.message = newbody
            post.save()
            return True
        return False
