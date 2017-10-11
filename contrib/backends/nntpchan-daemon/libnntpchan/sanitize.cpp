#include <nntpchan/sanitize.hpp>
#include <algorithm>
#include <regex>
#include <cctype>

namespace nntpchan
{
  
  std::string NNTPSanitizeLine(const std::string & str)
  {
    if(str == ".") return " .";
    std::string sane;
    sane += str;
    const char ch =  ' ';
    std::replace_if(sane.begin(), sane.end(), [](unsigned char ch) -> bool { return iscntrl(ch); } , ch);
    return sane;
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

  static const std::regex re_ValidNewsgroup("^[a-zA-Z][a-zA-Z0-9.]{1,128}$");

  bool IsValidNewsgroup(const std::string & msgid)
  {
    return std::regex_search(msgid, re_ValidNewsgroup) == 1;
  }
}
