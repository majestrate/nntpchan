#include <nntpchan/file_handle.hpp>


namespace nntpchan
{
  FileHandle_ptr OpenFile(const fs::path & fname, FileMode mode)
  {
    std::fstream * f = new std::fstream;
    if(mode == eRead)
    {
      f->open(fname, std::ios::in);
    }
    else if (mode == eWrite)
    {
      f->open(fname, std::ios::out);
    }
    if(f->is_open())
      return FileHandle_ptr(f);
    delete f;
    return nullptr;
  }
}
