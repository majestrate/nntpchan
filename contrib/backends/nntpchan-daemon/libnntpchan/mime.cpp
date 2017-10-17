#include <nntpchan/mime.hpp>

namespace nntpchan
{
bool ReadHeader(const FileHandle_ptr &file, RawHeader &header)
{
  std::string line;
  while (std::getline(*file, line) && !(line == "\r" || line == ""))
  {
    std::string k, v;
    auto idx = line.find(": ");
    auto endidx = line.size() - 1;

    while (line[endidx] == '\r')
      --endidx;

    if (idx != std::string::npos && idx + 2 < endidx)
    {
      k = line.substr(0, idx);
      v = line.substr(idx + 2, endidx);
      header[k] = v;
    }
  }
  return file->good();
}
}
