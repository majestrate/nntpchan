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
                ('id', models.AutoField(serialize=False, verbose_name='ID', primary_key=True, auto_created=True)),
                ('filehash', models.CharField(editable=False, max_length=256)),
                ('filename', models.CharField(max_length=256)),
                ('mimetype', models.CharField(max_length=256, default='text/plain')),
                ('width', models.IntegerField(default=0)),
                ('height', models.IntegerField(default=0)),
                ('banned', models.BooleanField(default=False)),
            ],
        ),
        migrations.CreateModel(
            name='Newsgroup',
            fields=[
                ('name', models.CharField(editable=False, max_length=256, serialize=False, primary_key=True)),
                ('posts_per_page', models.IntegerField(default=10)),
                ('max_pages', models.IntegerField(default=10)),
                ('banned', models.BooleanField(default=False)),
            ],
        ),
        migrations.CreateModel(
            name='Post',
            fields=[
                ('msgid', models.CharField(editable=False, max_length=256, serialize=False, primary_key=True)),
                ('posthash', models.CharField(editable=False, max_length=256)),
                ('reference', models.CharField(max_length=256, default='')),
                ('message', models.TextField(default='')),
                ('subject', models.CharField(max_length=256, default='None')),
                ('name', models.CharField(max_length=256, default='Anonymous')),
                ('pubkey', models.CharField(max_length=64, default='')),
                ('signature', models.CharField(max_length=64, default='')),
                ('posted', models.DateTimeField()),
                ('placeholder', models.BooleanField(default=False)),
                ('attachments', models.ManyToManyField(to='frontend.Attachment')),
                ('newsgroup', models.ForeignKey(to='frontend.Newsgroup')),
            ],
        ),
    ]
