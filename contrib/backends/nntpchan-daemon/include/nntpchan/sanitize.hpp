#ifndef NNTPCHAN_SANITIZE_HPP
#define NNTPCHAN_SANITIZE_HPP
#include <string>

namespace nntpchan
{
  std::string NNTPSanitizeLine(const std::string & str);
  std::string ToLower(const std::string & str);
  std::string StripWhitespaces(const std::string & str);
  bool IsValidMessageID(const std::string & msgid);
  bool IsValidNewsgroup(const std::string & group);
}

#endif
