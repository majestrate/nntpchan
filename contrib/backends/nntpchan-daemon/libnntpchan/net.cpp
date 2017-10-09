#include "net.hpp"
#include <uv.h>
#include <sstream>
#include <stdexcept>
#include <cstring>

namespace nntpchan
{
  std::string NetAddr::to_string()
  {
    std::string str("invalid");
    const size_t s = 128;
    char * buff = new char[s];
    if(uv_ip6_name(&addr, buff, s) == 0) {
      str = std::string(buff);
      delete [] buff;
    }
    std::stringstream ss;
    ss << "[" << str << "]:" << ntohs(addr.sin6_port);
    return ss.str();
  }

  NetAddr::NetAddr()
  {
    std::memset(&addr, 0, sizeof(addr));
  }
  
  NetAddr ParseAddr(const std::string & addr)
  {
    NetAddr saddr;
    auto n = addr.rfind("]:");
    if (n == std::string::npos) {
      throw std::runtime_error("invalid address: "+addr);
    }
    if (addr[0] != '[') {
      throw std::runtime_error("invalid address: "+addr);
    }
    auto p = addr.substr(n+2);
    int port = std::atoi(p.c_str());
    auto a = addr.substr(0, n);
    uv_ip6_addr(a.c_str(), port, &saddr.addr);
    return saddr;
  }
}
