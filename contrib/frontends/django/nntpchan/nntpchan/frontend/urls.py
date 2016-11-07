from django.conf.urls import url

from . import views

urlpatterns = [
    url(r'^ctl-(?P<page>[0-9]+)\.html$', views.modlog, name='old-modlog'),
    url(r'^ctl/((?P<page>[0-9]+)/)?$', views.modlog, name='modlog'),
    url(r'^overchan\.(?P<name>[a-zA-Z0-9\.]+)-(?P<page>[0-9]+)\.html$', views.BoardView.as_view(), name='old-board'),
    url(r'^overchan\.(?P<name>[a-zA-Z0-9\.]+)/', views.BoardView.as_view(), name='board-alt'),
    url(r'^thread-(?P<op>[a-fA-F0-9\.]{40})\.html$', views.ThreadView.as_view(), name='old-thread'),
    url(r'^b/(?P<name>[a-zA-Z0-9\.]+[a-zA-Z0-9])/$', views.BoardView.as_view(), name='board-front'),
    url(r'^b/(?P<name>[a-zA-Z0-9\.]+[a-zA-Z0-9])/(?P<page>[0-9]+)/$', views.BoardView.as_view(), name='board'),
    url(r'^t/(?P<op>[a-fA-F0-9\.]{40})/$', views.ThreadView.as_view(), name='thread'),
    url(r'captcha.png', views.create_captcha, name='captcha'),
    url(r'^$', views.FrontPageView.as_view(), name='frontpage'),
]
