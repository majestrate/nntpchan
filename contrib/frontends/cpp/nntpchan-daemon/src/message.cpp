#include "message.hpp"

namespace nntpchan
{
  bool IsValidMessageID(const MessageID & msgid)
  {
    auto itr = msgid.begin();
    auto end = msgid.end();
    --end;
    if (*itr != '<') return false;
    if (*end != '>') return false;
    bool atfound = false;
    while(itr != end) {
      auto c = *itr;
      ++itr;
      if(atfound && c == '@') return false;
      if(c == '@') {
        atfound = true;
        continue;
      }
      if (c == '$' || c == '_' || c == '-') continue;
      if (c > '0' && c < '9') continue;
      if (c > 'A' && c < 'Z') continue;
      if (c > 'a' && c < 'z') continue;
      return false;
    }
    return true;
  }
}
