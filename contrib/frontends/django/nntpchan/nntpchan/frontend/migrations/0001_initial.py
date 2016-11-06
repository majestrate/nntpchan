# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
    ]

    operations = [
        migrations.CreateModel(
            name='Attachment',
            fields=[
                ('id', models.AutoField(verbose_name='ID', serialize=False, auto_created=True, primary_key=True)),
                ('filehash', models.CharField(editable=False, max_length=256)),
                ('filename', models.CharField(max_length=256)),
                ('mimetype', models.CharField(default='text/plain', max_length=256)),
                ('width', models.IntegerField(default=0)),
                ('height', models.IntegerField(default=0)),
                ('banned', models.BooleanField(default=False)),
            ],
        ),
        migrations.CreateModel(
            name='Newsgroup',
            fields=[
                ('name', models.CharField(primary_key=True, serialize=False, editable=False, max_length=256)),
                ('posts_per_page', models.IntegerField(default=10)),
                ('max_pages', models.IntegerField(default=10)),
                ('banned', models.BooleanField(default=False)),
            ],
        ),
        migrations.CreateModel(
            name='Post',
            fields=[
                ('msgid', models.CharField(primary_key=True, serialize=False, editable=False, max_length=256)),
                ('posthash', models.CharField(editable=False, max_length=256)),
                ('reference', models.CharField(default='', max_length=256)),
                ('message', models.TextField(default='')),
                ('subject', models.CharField(default='None', max_length=256)),
                ('name', models.CharField(default='Anonymous', max_length=256)),
                ('pubkey', models.CharField(default='', max_length=64)),
                ('signature', models.CharField(default='', max_length=64)),
                ('posted', models.IntegerField(default=0)),
                ('placeholder', models.BooleanField(default=False)),
                ('last_bumped', models.IntegerField(default=0)),
                ('attachments', models.ManyToManyField(to='frontend.Attachment')),
                ('newsgroup', models.ForeignKey(to='frontend.Newsgroup')),
            ],
        ),
    ]
