from django.core.urlresolvers import reverse
from django.http import HttpResponse, Http404
from django.shortcuts import render, get_object_or_404
from django.views import generic


from .models import Post, Newsgroup

class Postable:
    """
    postable view
    checks captcha etc
    """

    def post(self, request, **kwargs):
        ctx = {
            'error' : None
        }

        
        
        return render(request, 'frontend/postresult.html', ctx)

        
class BoardView(generic.View, Postable):
    template_name = 'frontend/board.html'
    context_object_name = 'threads'
    model = Post

    def get(self, request, name):
        page = 0
        if 'p' in request.GET:
            page = request.GET['p']
        newsgroup = 'overchan.{}'.format(name)
        try:
            page = int(page or "0")
        except:
            page = 0
        if page < 0:
            page = 0
        try:
            group = Newsgroup.objects.get(name=newsgroup)
        except Newsgroup.DoesNotExist:
            raise Http404("no such board")
        else:
            begin = page * group.posts_per_page
            end = begin + group.posts_per_page - 1
            roots = self.model.objects.filter(newsgroup=group, reference='').order_by('-last_bumped')[begin:end]      
            ctx = {'threads': roots, 'page': page, 'name': newsgroup}
            if page < group.max_pages:
                ctx['nextpage'] = reverse('board', args=[name]) + '?p={}'.format(page + 1)
            if page > 0:
                ctx['prevpage'] = reverse('board', args=[name]) + '?p={}'.format(page - 1)
            return render(request, self.template_name, ctx)
        
class ThreadView(generic.ListView, Postable):
    template_name = 'frontend/thread.html'
    model = Post
    context_object_name = 'op'

    def get_queryset(self):
        return get_object_or_404(self.model, posthash=self.kwargs['op'])
    
class FrontPageView(generic.View):
    template_name = 'frontend/frontpage.html'
    model = Post

    def get(self, request, truncate=5):
        if truncate <= 0:
            truncate = 5
        posts = self.model.objects.order_by('-posted')[:truncate]
        ctx = {'posts' : posts}
        return render(request, self.template_name, ctx)
    
    
def modlog(request, page):
    if page is None:
        page = 0
    return HttpResponse('mod log page {}'.format(page))
