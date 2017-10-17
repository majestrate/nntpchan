#ifndef NNTPCHAN_IO_HANDLE_HPP
#define NNTPCHAN_IO_HANDLE_HPP
#include <iostream>
#include <memory>

namespace nntpchan
{
typedef std::unique_ptr<std::iostream> IOHandle_ptr;
}

#endif
