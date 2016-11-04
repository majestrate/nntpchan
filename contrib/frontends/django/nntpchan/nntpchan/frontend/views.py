from django.http import HttpResponse
from django.shortcuts import render
from django.views import generic

from .models import Post, Board

class IndexView(generic.DetailView):
    template_name = 'frontend/index.html'

class BoardView(generic.ListView):
    template_name = 'frontend/board.html'

class ThreadView(generic.ListView):
    template_name = 'frontend/thread.html'
    
def frontpage(request):
    return HttpResponse('ayyyy')

def boardpage(request, name, page):
    if page is None:
        page = 0
    name = 'overchan.{}'.format(name)
    return HttpResponse('{} page {}'.format(name, page))

def threadpage(request, op):
    return HttpResponse('thread {}'.format(op))

def redirect_thread(request, op):
    pass

def modlog(request, page):
    if page is None:
        page = 0
    return HttpResponse('mod log page {}'.format(page))
