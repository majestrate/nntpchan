#ifndef NNTPCHAN_FRONTEND_HPP
#define NNTPCHAN_FRONTEND_HPP
#include <string>

namespace nntpchan
{
  /** @brief nntpchan frontend ui interface */
  class Frontend
  {
  public:
    virtual ~Frontend() {}

    /** @brief process an inbound message stored at fpath that we have accepted. */
    virtual void ProcessNewMessage(const std::string & fpath) = 0;

    /** @brief return true if we take posts in a newsgroup */
    virtual bool AcceptsNewsgroup(const std::string & newsgroup) = 0;
    
    /** @brief return true if we will accept a message given its message-id */
    virtual bool AcceptsMessage(const std::string & msgid) = 0;
    
  };
}

#endif
