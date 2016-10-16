#include "line.hpp"
#include <iostream>

namespace nntpchan {

  void LineReader::OnData(const char * d, ssize_t l)
  {
    if(l <= 0) return;
    std::size_t idx = 0;
    while(l-- > 0) {
      char c = d[idx++];
      if(c == '\n') {
        OnLine(d, idx-1);
        d += idx;
      } else if (c == '\r' && d[idx] == '\n') {
        OnLine(d, idx-1);
        d += idx + 1;
      }
    }
  }

  void LineReader::OnLine(const char *d, const size_t l)
  {
    std::string line(d, l);
    HandleLine(line);
  }
  
  bool LineReader::HasNextLine()
  {
    return m_sendlines.size() > 0;
  }

  std::string LineReader::GetNextLine()
  {
    std::string line = m_sendlines[0];
    m_sendlines.pop_front();
    return line;
  }

  void LineReader::QueueLine(const std::string & line)
  {
    m_sendlines.push_back(line);
  }
}
