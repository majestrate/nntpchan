#include <nntpchan/event.hpp>
#include <sys/types.h>
#include <sys/event.h>

#include <iostream>

namespace nntpchan 
{
namespace ev 
{
    template<size_t bufsz>
    struct KqueueLoop : public Loop
    {
        int kfd;
        size_t conns;
        char readbuf[bufsz];


        KqueueLoop() : kfd(kqueue()), conns(0)
        {

        };

        virtual ~KqueueLoop()
        {
            ::close(kfd);
        }

        virtual bool TrackConn(ev::io * handler)
        {
            struct kevent event;
            short filter = 0;
            if(handler->readable() || handler->acceptable())
            {   
                filter |= EVFILT_READ;
            }
            if(handler->writeable())
            {
                filter |= EVFILT_WRITE;
            }
            EV_SET(&event, handler->fd, filter, EV_ADD | EV_CLEAR, 0, 0, handler);
            int ret = kevent(kfd, &event, 1, nullptr, 0, nullptr);
            if(ret == -1) return false;
            if(event.flags & EV_ERROR) 
            {
                std::cerr << "KqueueLoop::TrackConn() kevent failed: " << strerror(event.data) << std::endl;
                return false;
            }
            ++conns;
            return true;
        }

        virtual void UntrackConn(ev::io * handler)
        {
            struct kevent event;
            short filter = 0;
            if(handler->readable() || handler->acceptable())
            {   
                filter |= EVFILT_READ;
            }
            if(handler->writeable())
            {
                filter |= EVFILT_WRITE;
            }
            EV_SET(&event, handler->fd, filter, EV_DELETE, 0, 0, handler);
            int ret = kevent(kfd, &event, 1, nullptr, 0, nullptr);
            if(ret == -1 || event.flags & EV_ERROR) 
            {
                std::cerr << "KqueueLoop::UntrackConn() kevent failed: " << strerror(event.data) << std::endl;
                return false;
            }
            --conns;
            return true;
        }

        virtual void Run()
        {
            struct kevent events[512];
            struct kevent * event;
            io * handler;
            int ret, idx;
            do
            {
                idx = 0;
                ret = kevent(kfd, nullptr, 0, &event, 512, nullptr);
                if(ret > 0)
                {
                    while(idx < ret)
                    {
                        event = &events[idx++];
                        handler = static_cast<io *>(ev->udata);
                        if(event->flags & EV_EOF)
                        {
                            handler->close();
                            delete handler;
                            continue;
                        }
                        if(event->filter & EVFILT_READ && handler->acceptable())
                        {
                            int backlog = event->data;
                            while(backlog)
                            {
                                handler->accept();
                                --backlog;
                            }
                        }

                        if(event->filter & EVFILT_READ && handler->readable())
                        {
                            int readed = 0;
                            int readnum = event->data;
                            while(readnum > sizeof(readbuf))
                            {
                                int r = handler->read(readbuf, sizeof(readbuf));
                                if(r > 0)
                                {
                                    readnum -= r;
                                    readed += r;
                                }
                                else
                                    readnum = 0;
                            }
                            if(readnum && readed != -1)
                            {
                               int r = handler->read(readbuf, readnum);
                               if(r > 0)
                                 readed += r;
                               else
                                 readed = r;
                            }
                        }
                        if(event->filter & EVFILT_WRITE && handler->writeable())
                        {
                            int writespace = event->data;
                            int written = handler->write(writespace);
                            if(written > 0)
                            {

                            }
                        }
                        if(!handler->keepalive())
                        {
                            handler->close();
                            delete handler;
                        }
                    }
                }
            }
            while(ret != -1);
        }
    };
}
}