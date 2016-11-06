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
            print(begin, end)
            posts = self.model.objects.filter(newsgroup=group).order_by('-posted')[begin:end]
            roots = []
            for post in posts:
                if post.is_op():
                    if post not in roots:
                        roots.append(post)
                else:
                    op = self.model.objects.filter(msgid=post.reference)
                    if len(op) > 0:
                        op = op[0]
                        if op not in roots:
                            roots.append(op)
                        
            ctx = {'threads': roots,'page': page, 'name': newsgroup}
            if page < group.max_pages:
                ctx['nextpage'] = '/{}/{}/'.format(group.name, page + 1)
            if page > 0:
                ctx['prevpage'] = '/{}/{}/'.format(group.name, page - 1)
            return render(request, self.template_name, ctx)
        
        
class ThreadView(generic.ListView):
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
