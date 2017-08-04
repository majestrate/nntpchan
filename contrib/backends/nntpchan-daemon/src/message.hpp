#ifndef NNTPCHAN_MESSAGE_HPP
#define NNTPCHAN_MESSAGE_HPP

#include <string>
#include <vector>
#include <map>
#include <functional>

namespace nntpchan
{
  bool IsValidMessageID(const std::string & msgid);

  typedef std::pair<std::string, std::string> MessageHeader;

  typedef std::map<std::string, std::string> MIMEPartHeader;

  typedef std::function<bool(const MessageHeader &)> MessageHeaderFilter;

  typedef std::function<bool(const MIMEPartHeader &)> MIMEPartFilter;

  /**
     read MIME message from i,
     filter each header with h,
     filter each part with p,
     store result in o

     return true if we read the whole message, return false if there is remaining
   */
  bool StoreMIMEMessage(std::istream & i, MessageHeaderFilter h, MIMEPartHeader p, std::ostream & o);

}


#endif
