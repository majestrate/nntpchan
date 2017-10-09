#ifndef NNTPCHAN_NET_HPP
#define NNTPCHAN_NET_HPP

#include <sys/types.h>
#include <netinet/in.h>
#include <string>

namespace nntpchan
{
  struct NetAddr
  {
    NetAddr();
    
    sockaddr_in6 addr;
    operator sockaddr * () { return (sockaddr *) &addr; }
    operator const sockaddr * () const { return (sockaddr *) &addr; }
    std::string to_string();
  };

  NetAddr ParseAddr(const std::string & addr);
}

#endif 
