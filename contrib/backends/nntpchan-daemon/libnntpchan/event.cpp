#include <nntpchan/event.hpp>

#ifdef __linux__
#include "epoll.hpp"
typedef nntpchan::ev::EpollLoop LoopImpl;
#else
#ifdef __freebsd__
#include "kqueue.hpp"
typedef nntpchan::ev::KqueueLoop LoopImpl;
#else
#error "unsupported platform"
#endif
#endif

namespace nntpchan 
{
  ev::Loop * NewMainLoop()
  {
    return new LoopImpl;
  }
}