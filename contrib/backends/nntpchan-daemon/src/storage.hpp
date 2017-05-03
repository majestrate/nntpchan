#ifndef NNTPCHAN_STORAGE_HPP
#define NNTPCHAN_STORAGE_HPP

#include <fstream>
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

    std::fstream * OpenWrite(const std::string & msgid);
    std::fstream * OpenRead(const std::string & msgid);

    /**
       return true if we should accept a new message give its message id
     */
    bool Accept(const std::string & msgid);

  private:

    template<typename Mode>
    std::fstream * OpenMode(const std::string & msgid, const Mode & m)
    {
      if(IsValidMessageID(msgid))
      {
        std::fstream * f = new std::fstream;
        f->open(MessagePath(msgid), m);
        if(f->is_open())
          return f;
        delete f;
        return nullptr;
      }
      else
        return nullptr;
    };

    std::string MessagePath(const std::string & msgid);

    static char GetPathSep();

    std::string basedir;

  };
}


#endif
