#ifndef NNTPCHAN_EVENT_HPP
#define NNTPCHAN_EVENT_HPP

#include <unistd.h>
#include <cstdint>
#include <string>
#include <sys/socket.h>

namespace nntpchan
{
namespace ev
{
  struct io
  {
    int fd;

    io(int f) : fd(f) {};
    virtual ~io() {};
    virtual bool readable() const { return true; };
    virtual int read(char * buf, size_t sz) = 0;
    virtual bool writeable() const { return true; };
    virtual int write(size_t avail) = 0;
    virtual bool keepalive() = 0;
    virtual void close() 
    { 
      if(fd!=-1)
      {
        ::close(fd);
      }
    };
    virtual bool acceptable() const { return false; };
    virtual int accept() { return -1; };
  };

  struct Loop
  {
  public:
    virtual ~Loop() {};

    bool BindTCP(const sockaddr * addr, ev::io * handler);
    virtual bool TrackConn(ev::io * handler) = 0;
    virtual void UntrackConn(ev::io * handler) = 0;
    virtual void Run() = 0;
    bool SetNonBlocking(ev::io *handler);
  };
}

ev::Loop * NewMainLoop();

}

#endif
