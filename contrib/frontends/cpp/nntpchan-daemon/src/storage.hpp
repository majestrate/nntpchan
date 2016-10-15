#ifndef NNTPCHAN_STORAGE_HPP
#define NNTPCHAN_STORAGE_HPP

#include <string>
#include "message.hpp"

namespace nntpchan
{
  class ArticleStorage
  {
  public:
    ArticleStorage();
    ArticleStorage(const std::string & fpath);
    ~ArticleStorage();

    void SetPath(const std::string & fpath);

    std::ostream & OpenWrite(const MessageID & msgid);
    std::istream & OpenRead(const MessageID & msgid);

    /**
       return true if we should accept a new message give its message id
     */
    bool Accept(const MessageID & msgid);
    
  private:

    static char GetPathSep();
    
    std::string basedir;
    
  };
}


#endif
