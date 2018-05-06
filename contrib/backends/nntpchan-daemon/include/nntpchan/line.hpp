#ifndef NNTPCHAN_LINE_HPP
#define NNTPCHAN_LINE_HPP
#include "server.hpp"
#include <stdint.h>
#include <sstream>

namespace nntpchan
{

/** @brief a buffered line reader */
class LineReader
{
public:

  /** @brief queue inbound data from connection */
  void Data(const char *data, ssize_t s);

protected:
  /** @brief handle a line from the client */
  virtual void HandleLine(const std::string line) = 0;

private:
  
  std::stringstream m_line;
  std::string m_leftover;
};
}

#endif
