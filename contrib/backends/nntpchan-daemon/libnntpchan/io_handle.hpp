#ifndef NNTPCHAN_IO_HANDLE_HPP
#define NNTPCHAN_IO_HANDLE_HPP
#include <memory>
#include <iostream>

namespace nntpchan
{
  typedef std::unique_ptr<std::iostream> IOHandle_ptr;
  
}

#endif
