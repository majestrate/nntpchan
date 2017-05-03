#include "line.hpp"

namespace nntpchan {

  LineReader::LineReader(size_t limit) : m_close(false), lineLimit(limit) {}

  void LineReader::Data(const char * data, ssize_t l)
  {
    if(l <= 0) return;
    // process leftovers
    std::size_t idx = 0;
    std::size_t pos = 0;
    while(l-- > 0) {
      char c = data[idx++];
      if(c == '\n') {
        OnLine(data, pos);
        pos = 0;
        data += idx;
      } else if (c == '\r' && data[idx] == '\n') {
        OnLine(data, pos);
        data += idx + 1;
        pos = 0;
      } else {
        pos ++;
      }
    }
  }

  void LineReader::OnLine(const char *d, const size_t l)
  {
    std::string line(d, l);
    HandleLine(line);
  }

  bool LineReader::ShouldClose()
  {
    return m_close;
  }
}
