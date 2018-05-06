#include <cassert>
#include <nntpchan/crypto.hpp>
#include <nntpchan/sanitize.hpp>
#include <nntpchan/storage.hpp>
#include <sstream>

namespace nntpchan
{

const fs::path posts_skiplist_dir = "posts";
const fs::path threads_skiplist_dir = "threads";

ArticleStorage::ArticleStorage(const fs::path &fpath) { SetPath(fpath); }

ArticleStorage::~ArticleStorage() {}

void ArticleStorage::SetPath(const fs::path &fpath)
{
  basedir = fpath;
  fs::create_directories(basedir);
  assert(init_skiplist(posts_skiplist_dir));
  assert(init_skiplist(threads_skiplist_dir));
  errno = 0;
}

bool ArticleStorage::init_skiplist(const std::string &subdir) const
{
  fs::path skiplist = skiplist_root(subdir);
  const auto subdirs = {
      'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p',
      'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '2', '3', '4', '5', '6', '7',
  };
  for (const auto &s : subdirs)
    fs::create_directories(skiplist / std::string(&s, 1));
  return true;
}

bool ArticleStorage::Accept(const std::string &msgid) const
{
  if (!IsValidMessageID(msgid))
    return false;
  auto p = MessagePath(msgid);
  bool ret = !fs::exists(p);
  errno = 0;
  return ret;
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
void ArticleStorage::EnsureSymlinks(const std::string &msgid) const
{
  std::string msgidhash = Blake2B_base32(msgid);
  auto skip = skiplist_dir(skiplist_root(posts_skiplist_dir), msgidhash) / msgidhash;
  auto path = fs::path("..") / fs::path("..") / fs::path("..") / MessagePath(msgid);
  fs::create_symlink(path, skip);
  errno = 0;
}

fs::path ArticleStorage::skiplist_root(const std::string &name) const { return basedir / name; }
fs::path ArticleStorage::skiplist_dir(const fs::path &root, const std::string &name) const
{
  return root / name.substr(0, 1);
}
}
