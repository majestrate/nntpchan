#include <cassert>
#include <nntpchan/event.hpp>
#include <sys/epoll.h>
#include <unistd.h>
#include <netinet/in.h>
#include <sys/un.h>
#include <fcntl.h>
#include <signal.h>
#include <sys/signalfd.h>

#include <iostream>

namespace nntpchan
{
namespace ev
{
  struct EpollLoop : public Loop
  {
    size_t conns;
    int epollfd;
    char readbuf[128];
    EpollLoop() : conns(0), epollfd(epoll_create1(EPOLL_CLOEXEC))
    {
    }
    ~EpollLoop()
    {
        ::close(epollfd);
    }

    virtual bool BindTCP(const sockaddr * addr, ev::io * handler)
    {
      assert(handler->acceptable());
      socklen_t slen;
      switch(addr->sa_family)
      {
        case AF_INET:
        slen = sizeof(sockaddr_in);
        break;
        case AF_INET6:
        slen = sizeof(sockaddr_in6);
        break;
        case AF_UNIX:
        slen = sizeof(sockaddr_un);
        break;
        default:
        return false;
      }
      int fd = socket(addr->sa_family, SOCK_STREAM | SOCK_NONBLOCK, 0);
      if(fd == -1)
      { 
        return false;
      }
      handler->fd = fd;
   
      if(bind(fd, addr, slen) == -1)
        return false;
  
      if (listen(fd, 5) == -1)
        return false;

      return TrackConn(handler);
    }

    virtual bool TrackConn(ev::io * handler)
    {
        epoll_event ev;
        ev.data.ptr = handler;
        ev.events = EPOLLET;
        if(handler->readable() || handler->acceptable())
        {
            ev.events |= EPOLLIN;
        }
        if(handler->writeable())
        {
            ev.events |= EPOLLOUT;
        }
        if ( epoll_ctl(epollfd, EPOLL_CTL_ADD, handler->fd, &ev) == -1)
        {
            return false;
        }
        ++conns;
        return true;
        }

    virtual void UntrackConn(ev::io * handler)
    {
        if(epoll_ctl(epollfd, EPOLL_CTL_DEL, handler->fd, nullptr) != -1)
            --conns;
    }


    virtual void Run() 
    {
        epoll_event evs[512];
        epoll_event * ev;
        ev::io * handler;
        int res = -1;
        int idx ;

        sigset_t mask;
        
        sigemptyset(&mask);
        sigaddset(&mask, SIGWINCH);

        int sfd = signalfd(-1, &mask, SFD_NONBLOCK | SFD_CLOEXEC);
        epoll_event sig_ev;
        sig_ev.data.fd = sfd;
        sig_ev.events = EPOLLIN;
        epoll_ctl(epollfd, EPOLL_CTL_ADD, sfd, &sig_ev);
        do
        {
            res = epoll_wait(epollfd, evs, 512, -1);
            idx = 0;
            while(idx < res)
            {
            errno = 0;
            ev = &evs[idx++];
            if(ev->data.fd == sfd)
            {
                read(sfd, readbuf, sizeof(readbuf));
                continue;
            }
            
            handler = static_cast<ev::io *>(ev->data.ptr);

            if(ev->events & EPOLLERR || ev->events & EPOLLHUP)
            {
                handler->close();
                delete handler;
                continue;
            }

            if (handler->acceptable())
            {
                int acceptfd;
                bool errored = false;
                while(true)
                {
                acceptfd = handler->accept();
                if(acceptfd == -1)
                {
                    if (errno == EAGAIN || errno == EWOULDBLOCK)
                    {
                    break;
                    }
                    perror("accept()");
                    errored = true;
                    break;
                }
                }
                if(errored)
                {
                handler->close();
                delete handler;
                continue;
                }
            }
            if(ev->events & EPOLLIN && handler->readable())
            {
                bool errored = false;
                while(true)
                {
                int readed = handler->read(readbuf, sizeof(readbuf));
                if(readed == -1)
                {
                    if(errno != EAGAIN)
                    {
                    perror("read()");
                    handler->close();
                    delete handler;
                    errored = true;
                    }
                    break;
                }
                else if (readed == 0)
                {
                    handler->close();
                    delete handler;
                    errored = true;
                    break;
                }
                }
                if(errored) continue;
            }
            if(ev->events & EPOLLOUT && handler->writeable())
            {
                int written = handler->write();
                if(written < 0)
                {
                if (errno == EAGAIN || errno == EWOULDBLOCK)
                {
                    // blocking
                }
                else
                {
                    perror("write()");
                    handler->close();
                    delete handler;
                }
                }
            }
            if (!handler->keepalive())
            {
                handler->close();
                delete handler;
            }
            }
        }
        while(res != -1 && conns);
    }
};
}
}
