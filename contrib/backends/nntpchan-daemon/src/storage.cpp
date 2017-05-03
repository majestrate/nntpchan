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

  bool ArticleStorage::Accept(const std::string& msgid)
  {
    if (!IsValidMessageID(msgid)) return false;
    auto s = MessagePath(msgid);
    FILE * f = fopen(s.c_str(), "r");
    if ( f == nullptr) return errno == ENOENT;
    fclose(f);
    return false;
  }

  std::string ArticleStorage::MessagePath(const std::string & msgid)
  {
    return basedir + GetPathSep() + msgid;
  }

  std::fstream * ArticleStorage::OpenRead(const std::string & msgid)
  {
    return OpenMode(msgid, std::ios::in);
  }

  std::fstream * ArticleStorage::OpenWrite(const std::string & msgid)
  {
    return OpenMode(msgid, std::ios::out);
  }

  char ArticleStorage::GetPathSep()
  {
    return '/';
  }


}
