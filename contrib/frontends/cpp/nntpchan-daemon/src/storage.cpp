#include "storage.hpp"
#include <errno.h>
#include <sys/stat.h>
#include <sstream>

namespace nntpchan
{
  ArticleStorage::ArticleStorage()
  {
  }

  ArticleStorage::ArticleStorage(const std::string & fpath) {
    SetPath(fpath);
  }
  
  ArticleStorage::~ArticleStorage()
  {
  }

  void ArticleStorage::SetPath(const std::string & fpath)
  {
    basedir = fpath;
    // quiet fail
    // TODO: check for errors
    mkdir(basedir.c_str(), 0700);
  }

  bool ArticleStorage::Accept(const MessageID & msgid)
  {
    if (!IsValidMessageID(msgid)) return false;
    std::stringstream ss;
    ss << basedir << GetPathSep() << msgid;
    auto s = ss.str();
    FILE * f = fopen(s.c_str(), "r");
    if ( f == nullptr) return errno == ENOENT;
    fclose(f);
    return false;
  }

  char ArticleStorage::GetPathSep()
  {
    return '/';
  }

  
}
