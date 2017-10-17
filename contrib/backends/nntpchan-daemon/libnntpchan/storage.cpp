#include <cassert>
#include <nntpchan/sanitize.hpp>
#include <nntpchan/storage.hpp>
#include <sstream>

namespace nntpchan
{

ArticleStorage::ArticleStorage(const fs::path &fpath) { SetPath(fpath); }

ArticleStorage::~ArticleStorage() {}

void ArticleStorage::SetPath(const fs::path &fpath)
{
  basedir = fpath;
  fs::create_directories(basedir);
  assert(init_skiplist("posts_skiplist"));
}

bool ArticleStorage::init_skiplist(const std::string &subdir) const
{
  fs::path skiplist = basedir / fs::path(subdir);
  fs::create_directories(skiplist);
  const auto subdirs = {"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"};
  for (const auto &s : subdirs)
    fs::create_directories(skiplist / s);
  return true;
}

bool ArticleStorage::Accept(const std::string &msgid) const
{
  if (!IsValidMessageID(msgid))
    return false;
  auto p = MessagePath(msgid);
  return !fs::exists(p);
}

fs::path ArticleStorage::MessagePath(const std::string &msgid) const { return basedir / msgid; }

FileHandle_ptr ArticleStorage::OpenRead(const std::string &msgid) const { return OpenFile(MessagePath(msgid), eRead); }

FileHandle_ptr ArticleStorage::OpenWrite(const std::string &msgid) const
{
  return OpenFile(MessagePath(msgid), eWrite);
}

bool ArticleStorage::LoadBoardPage(BoardPage &board, const std::string &newsgroup, uint32_t perpage,
                                   uint32_t page) const
{
  (void)board;
  (void)newsgroup;
  (void)perpage;
  (void)page;
  return false;
}
bool ArticleStorage::FindThreadByHash(const std::string &hashhex, std::string &msgid) const
{
  (void)hashhex;
  (void)msgid;
  return false;
}
bool ArticleStorage::LoadThread(Thread &thread, const std::string &rootmsgid) const
{
  (void)thread;
  (void)rootmsgid;
  return false;
}

/** ensure symlinks are formed for this article by message id */
void ArticleStorage::EnsureSymlinks(const std::string &msgid) const { (void)msgid; }
}
