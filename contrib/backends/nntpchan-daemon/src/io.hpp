#ifndef NNTPCHAN_IO_HPP
#define NNTPCHAN_IO_HPP
#include <iostream>
#include <vector>


namespace nntpchan
{
  namespace io
  {
    template<size_t bufferSize>
    ssize_t copy(std::ostream * dst, std::istream * src)
    {
      char buff[bufferSize];
      ssize_t n = 0;
      while(!src->eof() && src->good())
      {
        std::streamsize sz;
        src->read(buff, bufferSize);
        sz = src->gcount();
        n += sz;
        dst->write(buff, sz);
        if(!dst->good())
          break;
      }
      return n;
    }

    std::basic_ostream<char> * multiWriter(const std::vector<std::basic_ostream<char> *> & writers);
    std::basic_istream<char> * teeReader(std::basic_istream<char> * i, std::basic_ostream<char> * o);
  }
}

#endif
