#include "message.hpp"

namespace nntpchan
{
  bool IsValidMessageID(const std::string & msgid)
  {
    if(msgid[0] != '<') return false;
    if(msgid[msgid.size()-1] != '>') return false;
    auto itr = msgid.begin() + 1;
    auto end = msgid.end() - 1;
    bool atfound = false;
    while(itr != end) {
      auto c = *itr;
      ++itr;
      if(atfound && c == '@') return false;
      if(c == '@') {
        atfound = true;
        continue;
      }
      if (c == '$' || c == '_' || c == '-' || c == '.') continue;
      if (c >= '0' && c <= '9') continue;
      if (c >= 'A' && c <= 'Z') continue;
      if (c >= 'a' && c <= 'z') continue;
      return false;
    }
    return true;
  }
}
