#!/usr/bin/env python3.4

import requests
from bs4 import BeautifulSoup as BS

import datetime
import io
import nntplib
import time
import urllib.parse

class Article:
    """
    an nntp article
    """

    timeFormat = '%a, %d %b %Y %H:%M:%S +0000' 

    def __init__(self, j, board, site):
        """
        construct article
        :param j: json object
        """
        self.j = j
        self.board = board
        self.site = site
        self.messageID = self.genMessageID(j['no'])
        self.attachments = list()
        
    def formatDate(self):
        return datetime.datetime.utcfromtimestamp(self.j['time']).strftime(self.timeFormat)
        
    def message(self):
        m = ''
        # for each line
        for p in BS(self.j['com']).find_all('p'):
            # clean it
            m += p.text
            m += '\n'
        if len(m.rstrip('\n')) > 0:
            return m

    def subject(self):
        if 'subject' in self.j:
            return self.j['subject']

    def name(self):
        return self.j['name']

    def group(self):
        return 'overchan.archive.{}.{}'.format(self.site, self.board)
        
    def genMessageID(self, no):
        return '<{}.{}@{}>'.format(self.board, no, self.site)
        
    def header(self):
        hdr = ("Subject: {}\n"+\
        "From: {} <archiver@srndv2.tools>\n"+\
        "Date: {}\n"+\
        "Newsgroups: {}\n"+\
        "Message-ID: {}\n"+\
        "Path: {}\n").format(self.subject(), self.name(), self.formatDate(), self.group(), self.messageID, self.site)
        if self.j['resto'] > 0:
            hdr += "References: {}\n".format(self.genMessageID(self.j['resto']))
        return hdr

    def bodyPlain(self):
        msg = self.message()
        if msg:
            return "{}\n{}".format(self.header(), msg)

    def bodyMultipart(self):
        pass

    def body(self):
        if len(self.attachments) > 0:
            return self.bodyMultipart()
        else:
            return self.bodyPlain()
        
class Poster:

    def __init__(self, host, port):
        self.host, self.port = host, port
        
    def post(self, articles):
        """
        post 1 or more articles
        """
        if isinstance(articles, Article):
            return self.post([articles])
        else:
            n = nntplib.NNTP(self.host, self.port)
            for article in articles:
                body = article.body()
                if body:
                    print("posting {}".format(article.messageID))
                    try:
                        body = io.BytesIO(body.encode('utf-8'))
                        n.ihave(article.messageID, body)
                    except Exception as e:
                        print('failed: {}'.format(e))
            n.quit()


def url_parse(url):
    return urllib.parse.urlparse(url)
            
class Getter:

    def __init__(self, url):
        self.url = url
        self.site = url_parse(url).hostname
        self.board = url_parse(url).path.split('/')[1]

    def get(self):
        """
        yield a bunch of articles
        """
        r = requests.get(self.url)
        if r.status_code == 200:
            try:
                j = r.json()
            except:
                pass
            else:
                for t in j['threads']:
                    posts = t['posts']
                    for post in posts:
                        yield Article(post, self.board, self.site)



def main():
    import argparse
    ap = argparse.ArgumentParser()
    ap.add_argument('--server', type=str, required=True)
    ap.add_argument('--port', type=int, required=True)
    ap.add_argument('--board', type=str, required=True)
    args = ap.parse_args()
    poster = Poster(args.server, args.port)
    for n in range(10):
        getter = Getter('https://8ch.net/{}/{}.json'.format(args.board, n))
        poster.post(getter.get())
    
    
if __name__ == "__main__":
    main()
