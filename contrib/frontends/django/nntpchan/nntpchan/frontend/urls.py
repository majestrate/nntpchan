from django.conf.urls import url

from . import views

urlpatterns = [
    url(r'^ctl-(?P<page>[0-9])\.html$', views.modlog, name='old-modlog'),
    url(r'^ctl/((?P<page>[0-9])/)?$', views.modlog, name='modlog'),
    url(r'^overchan\.(?P<name>[a-zA-z0-9\.]+)-(?P<page>[0-9])\.html$', views.boardpage, name='old-board'),
    url(r'^(?P<name>[a-zA-z0-9\.]+)/((?P<page>[0-9])/)?$', views.boardpage, name='board'),
    url(r'^thread-(?P<op>[a-fA-F0-9\.]{40})\.html$', views.threadpage, name='old-thread'),
    url(r'^t/(?P<op>[a-fA-F0-9\.]{40})\.html$', views.redirect_thread, name='redirect-thread'),
    url(r'^t/(?P<op>[a-fA-F0-9\.]{40})/$', views.threadpage, name='thread'),
    url(r'^$', views.frontpage, name='index'),
]
