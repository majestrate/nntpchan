from django.http import HttpResponse, Http404
from django.shortcuts import render, get_object_or_404
from django.views import generic

from .models import Post, Newsgroup

class BoardView(generic.View):
    template_name = 'frontend/board.html'
    context_object_name = 'threads'
    model = Post

    def get(self, request, name, page):
        newsgroup = 'overchan.{}'.format(name)
        page = int(page or "0")
        try:
            group = Newsgroup.objects.get(name=newsgroup)
        except Newsgroup.DoesNotExist:
            raise Http404("no such board")
        else:
            begin = page * group.posts_per_page
            end = begin + group.posts_per_page
            posts = self.model.objects.filter(newsgroup=group, reference='').order_by('-posted')[begin:end]
            return render(request, self.template_name, {'threads': posts, 'page': page, 'name': newsgroup})
        
        
class ThreadView(generic.ListView):
    template_name = 'frontend/thread.html'
    model = Post
    context_object_name = 'op'

    def get_queryset(self):
        return get_object_or_404(self.model, posthash=self.kwargs['op'])
    


class FrontPageView(generic.ListView):
    template_name = 'frontend/frontpage.html'
    model = Post

    def get_queryset(self):
        return self.model.objects.order_by('posted')[:10]
    
    
def modlog(request, page):
    if page is None:
        page = 0
    return HttpResponse('mod log page {}'.format(page))
