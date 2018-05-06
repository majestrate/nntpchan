#include <cstring>
#include <nntpchan/net.hpp>
#include <sstream>
#include <stdexcept>
#include <arpa/inet.h>


namespace nntpchan
{
std::string NetAddr::to_string()
{
  std::string str("invalid");
  const size_t s = 128;
  char *buff = new char[s];
  if (inet_ntop(AF_INET6, &addr, buff, sizeof(sockaddr_in6)))
  {
    str = std::string(buff);
    delete[] buff;
  }
  std::stringstream ss;
  ss << "[" << str << "]:" << ntohs(addr.sin6_port);
  return ss.str();
}

NetAddr::NetAddr() { std::memset(&addr, 0, sizeof(addr)); }

NetAddr ParseAddr(const std::string &addr)
{
  NetAddr saddr;
  auto n = addr.rfind("]:");
  if (n == std::string::npos)
  {
    throw std::runtime_error("invalid address: " + addr);
  }
  if (addr[0] != '[')
  {
    throw std::runtime_error("invalid address: " + addr);
  }
  auto p = addr.substr(n + 2);
  int port = std::atoi(p.c_str());
  auto a = addr.substr(0, n);
  saddr.addr.sin6_port = htons(port);
  saddr.addr.sin6_family = AF_INET6;
  inet_pton(AF_INET6, a.c_str(), &saddr.addr);
  return saddr;
}
}
