#ifndef NNTPCHAN_SANITIZE_HPP
#define NNTPCHAN_SANITIZE_HPP
#include <string>

namespace nntpchan
{
  std::string NNTPSanitize(const std::string & str);
  std::string ToLower(const std::string & str);
  bool IsValidMessageID(const std::string & msgid);
}

#endif
