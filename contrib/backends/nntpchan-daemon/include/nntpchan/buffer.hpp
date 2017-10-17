#ifndef NNTPCHAN_BUFFER_HPP
#define NNTPCHAN_BUFFER_HPP
#include <string>
#include <uv.h>

namespace nntpchan
{
struct WriteBuffer
{
  uv_write_t w;
  uv_buf_t b;

  WriteBuffer(const std::string &s);
  WriteBuffer(const char *b, const size_t s);
  ~WriteBuffer();
};
}

#endif
