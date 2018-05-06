#ifndef NNTPCHAN_MESSAGE_HPP
#define NNTPCHAN_MESSAGE_HPP

#include <memory>
#include <nntpchan/model.hpp>

namespace nntpchan
{
struct MessageDB
{
  using BoardPage = nntpchan::model::BoardPage;
  using Thread = nntpchan::model::Thread;
  virtual ~MessageDB() {};
  virtual bool LoadBoardPage(BoardPage &board, const std::string &newsgroup, uint32_t perpage, uint32_t page) const = 0;
  virtual bool FindThreadByHash(const std::string &hashhex, std::string &msgid) const = 0;
  virtual bool LoadThread(Thread &thread, const std::string &rootmsgid) const = 0;
};

typedef std::unique_ptr<MessageDB> MessageDB_ptr;
}

#endif
