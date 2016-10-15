#ifndef NNTPCHAN_BASE64_HPP
#define NNTPCHAN_BASE64_HPP
#include <string>
#include <vector>

namespace nntpchan
{
  /** returns base64 encoded string */
  std::string B64Encode(const uint8_t * data, const std::size_t l);
  
  /** @brief returns true if decode was successful */
  bool B64Decode(const std::string & data, std::vector<uint8_t> & out);
  
}


#endif
