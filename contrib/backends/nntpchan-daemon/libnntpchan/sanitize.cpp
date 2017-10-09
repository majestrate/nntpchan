#include "sanitize.hpp"
#include <algorithm>
#include <regex>

namespace nntpchan
{
  std::string NNTPSanitize(const std::string & str)
  {
  }

  std::string ToLower(const std::string & str)
  {
    std::string lower = str;
    std::transform(lower.begin(), lower.end(), lower.begin(), [](unsigned char ch) -> unsigned char { return std::tolower(ch); } );
    return lower;
  }

  static const std::regex re_ValidMessageID("^<[a-zA-Z0-9$\\._]{2,128}@[a-zA-Z0-9\\-\\.]{2,63}>$");

  bool IsValidMessageID(const std::string & msgid)
  {
    return std::regex_search(msgid, re_ValidMessageID) == 1;
  }
}
