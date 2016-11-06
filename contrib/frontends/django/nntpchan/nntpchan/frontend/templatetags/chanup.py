from django import template
from django.template.defaultfilters import stringfilter

from django.utils.html import conditional_escape
from django.utils.safestring import mark_safe


from nntpchan.frontend.models import Newsgroup, Post

import re
from urllib.parse import urlparse
from html import unescape

register = template.Library()

re_postcite = re.compile('>> ?([0-9a-fA-F]+)')
re_boardlink = re.compile('>>> ?/([a-zA-Z0-9\.]+[a-zA-Z0-9])/')
re_redtext = re.compile('== ?(.+) ?==')
re_psytext = re.compile('@@ ?(.+) ?@@')

def greentext(text, esc):
    return_text = ''
    f = False
    for line in text.split('\n'):
        line = line.strip()
        if len(line) == 0:
            continue
        if line[0] == '>' and line[1] != '>':
            return_text += '<span class="greentext">%s </span>' % esc ( line ) + '\n'
            f = True
        else:
            return_text += esc(line) + '\n'
    return return_text, f

def blocktext(text, esc, delim='', css='', tag='span'):
    parts = text.split(delim)
    f = False
    if len(parts) > 1:
        parts.reverse()
        return_text = ''
        while len(parts) > 0:
            return_text += esc(parts.pop())
            if len(parts) > 0:
                f = True
                return_text += '<{} class="{}">%s</{}>'.format(tag,css,tag) % esc(parts.pop())
        return return_text, f
    else:
        return text, f

redtext = lambda t, e : blocktext(t, e, '==', 'redtext')
psytext = lambda t, e : blocktext(t, e, '@@', 'psy')
codeblock = lambda t, e : blocktext(t, e, '[code]', 'code', 'pre')


def postcite(text, esc):
    return_text = ''
    filtered = False
    for line in text.split('\n'):
        for word in line.split(' '):
            match = re_postcite.match(unescape(word))
            if match:
                posthash = match.groups()[0]
                posts = Post.objects.filter(posthash__startswith=posthash)
                if len(posts) > 0:
                    filtered = True
                    return_text +=  '<a href="%s" class="postcite">&gt;&gt%s</a> ' % ( posts[0].get_absolute_url(), posthash)
                else:
                    return_text += '<span class="greentext">&gt;&gt;%s</span> ' % match.string
            elif filtered:
                return_text += word + ' '
            else:
                return_text += esc(word) + ' '
        return_text += '\n'
    return return_text, filtered


def boardlink(text, esc):
    return_text = ''
    filtered = False
    for line in text.split('\n'):
        for word in line.split(' '):
            match = re_boardlink.match(unescape(word))
            if match:
                name = match.groups()[0]
                group = Newsgroup.objects.filter(name=name)
                if len(group) > 0:
                    filtered = True
                    return_text += '<a href="%s" class="boardlink">%s</a> ' % ( group[0].get_absolute_url(), esc(match.string ) )
                else:
                    return_text += '<span class="greentext">%s</span> ' % esc (match.string)
            else:
                return_text += esc(word) + ' '
        return_text += '\n'
    return return_text, filtered
    

def urlify(text, esc):
    return_text = ''
    filtered = False
    for line in text.split('\n'):
        for word in line.split(' '):
            u = urlparse(word)
            if u.scheme != '' and u.netloc != '':
                return_text +=  '<a href="%s">%s</a> ' % ( u.geturl(), esc(word) )
                filtered = True
            else:
                return_text += esc(word) + ' '
        return_text += '\n'
    return return_text, filtered

line_funcs = [
    greentext,
    redtext,
    urlify,
    psytext,
    codeblock,
    postcite,
    boardlink,
]

@register.filter(needs_autoescape=True, name='memepost')
def memepost(text, autoescape=True):
    text, _ = line_funcs[0](text, conditional_escape)
    for f in line_funcs[1:]:
        text, _ = f(text, lambda x : x)
    return mark_safe(text)


@register.filter(name='truncate')
@stringfilter
def truncate(text, truncate=500):
    if len(text) > truncate:
        return text[:truncate] + '...'
    return text
