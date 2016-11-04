from django.db import models

from . import util

class Attachment(models.Model):
    filename = models.CharField(max_length=256)
    filepath = models.CharField(max_length=256)
    width = models.IntegerField()
    height = models.IntegerField()
    
class Post(models.Model):
    msgid = models.CharField(max_length=256, primary_key=True)
    reference = models.CharField(max_length=256)
    message = models.TextField()
    subject = models.CharField(max_length=256)
    name = models.CharField(max_length=256)
    pubkey = models.CharField(max_length=64)
    signature = models.CharField(max_length=64)

    def get_absolute_url(self):
        from django.urls import reverse
        op = self.msgid
        if self.reference != self.msgid:
            op = self.reference
        
        return reverse('frontend.views.threadpage', args=[util.hashid(op)])

class Board(models.Model):
    pass
