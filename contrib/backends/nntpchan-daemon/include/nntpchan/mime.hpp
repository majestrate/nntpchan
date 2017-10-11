#ifndef NNTPCHAN_MIME_HPP
#define NNTPCHAN_MIME_HPP
#include "file_handle.hpp"
#include "io_handle.hpp"
#include <functional>
#include <map>
#include <string>

namespace nntpchan
{

  typedef std::map<std::string, std::string> RawHeader;

  bool ReadHeader(const FileHandle_ptr & f, RawHeader & h);


  struct MimePart
  {
    virtual RawHeader & Header() = 0;
    virtual IOHandle_ptr OpenPart() = 0;
  };

  typedef std::unique_ptr<MimePart> MimePart_ptr;
  
  typedef std::function<bool(MimePart_ptr)> PartReader;
  
  bool ReadParts(const FileHandle_ptr & f, PartReader r);
}

#endif
