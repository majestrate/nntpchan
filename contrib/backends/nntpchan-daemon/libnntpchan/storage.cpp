#include <nntpchan/storage.hpp>
#include <nntpchan/sanitize.hpp>
#include <sstream>

namespace nntpchan
{
  ArticleStorage::ArticleStorage()
  {
  }

  ArticleStorage::ArticleStorage(const fs::path & fpath) {
    SetPath(fpath);
  }

  ArticleStorage::~ArticleStorage()
  {
  }

  void ArticleStorage::SetPath(const fs::path & fpath)
  {
    basedir = fpath;
    fs::create_directories(basedir);
  }

  bool ArticleStorage::Accept(const std::string& msgid)
  {
    if (!IsValidMessageID(msgid)) return false;
    auto p = MessagePath(msgid);
    return !fs::exists(p);
  }

  fs::path ArticleStorage::MessagePath(const std::string & msgid)
  {
    return basedir / msgid;
  }

  FileHandle_ptr ArticleStorage::OpenRead(const std::string & msgid)
  {
    return OpenFile(MessagePath(msgid), eRead);
  }

  FileHandle_ptr ArticleStorage::OpenWrite(const std::string & msgid)
  {
    return OpenFile(MessagePath(msgid), eWrite);
  }

}
