#include <nntpchan/line.hpp>

namespace nntpchan
{

LineReader::LineReader(size_t limit) : m_close(false), lineLimit(limit) {}

void LineReader::Data(const char *data, ssize_t l)
{
  if (l <= 0)
    return;
  m_line << m_leftover;
  m_leftover = "";
  m_line << std::string(data, l);

  for(std::string line; std::getline(m_line, line); )
  {
    line.erase(std::remove(line.begin(), line.end(), '\r'), line.end());
    HandleLine(line);
  }
  if(m_line)
    m_leftover = m_line.str();
  m_line.clear();
}


bool LineReader::ShouldClose() { return m_close; }
}
