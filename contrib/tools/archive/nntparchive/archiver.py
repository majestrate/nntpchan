#!/usr/bin/env python3.4

import requests
from bs4 import BeautifulSoup as BS

import datetime
import io
import nntplib
import time
import urllib.parse
import base64
import random

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
        if 'filename' in self.j:
            hdr += 'Mime-Version: 1.0\n'
            hdr += 'Content-Type: multipart/mixed; boundary="{}"\n'.format(self.boundary)
        return hdr

    def bodyPlain(self):
        msg = self.message()
        if msg:
            return "{}\n{}".format(self.header(), msg)

    def bodyMultipart(self):
        self.boundary = '========{}'.format(random.randint(0, 10000000))
        msg = self.header() + '\n'
        msg += '--{}\n'.format(self.boundary)
        msg += 'Content-Type: text/plain; encoding=UTF-8\n'
        msg += '\n'
        msg += self.message() + '\n'
        msg += '--{}\n'.format(self.boundary)
        msg += 'Content-Type: image/{}\n'.format(self.j['ext'])
        msg += 'Content-Disposition: form-data; filename="{}{}"; name="import"\n'.format(self.j['filename'], self.j['ext'])
        msg += 'Content-Transfer-Encoding: base64\n'        
        msg += '\n'
        url = 'https://{}/{}/src/{}{}'.format(self.site, self.board, self.j['tim'], self.j['ext'])
        print ('obtain {}'.format(url))
        
        r = requests.get(url)
        if r.status_code == 200:
            msg += base64.b64encode(r.content).decode('ascii')
            msg += '\n'
        else:
            print ('failed to obtain attachment: {} != 200'.format(r.status_code))
            return
        msg += '\n--{}--\n'.format(self.boundary)
        return msg

    def body(self):
        if 'filename' in self.j:
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
                if 'threads' in j:
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
