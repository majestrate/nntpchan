#include "buffer.hpp"
#include <cstring>

namespace nntpchan
{
  WriteBuffer::WriteBuffer(const char * b, const size_t s)
  {
    char * buf = new char[s];
    std::memcpy(buf, b, s);
    this->b = uv_buf_init(buf, s);
    w.data = this;
  };

  WriteBuffer::WriteBuffer(const std::string & s) : WriteBuffer(s.c_str(), s.size()) {}

  WriteBuffer::~WriteBuffer()
  {
    delete [] b.base;
  }
}

