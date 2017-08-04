#ifndef NNTPCHAN_MESSAGE_HPP
#define NNTPCHAN_MESSAGE_HPP

#include <functional>
#include <string>
#include <vector>

namespace nntpchan
{
  bool IsValidMessageID(const std::string & msgid);

  namespace mime
  {
    typedef std::function<bool(std::string, std::string)> HeaderValueVisitor;

    struct PartHeader
    {
      std::string boundary;
      std::map<std::string, std::string> values;

    };

  }
}


#endif
