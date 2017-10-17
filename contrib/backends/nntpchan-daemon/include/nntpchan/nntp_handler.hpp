#ifndef NNTPCHAN_NNTP_HANDLER_HPP
#define NNTPCHAN_NNTP_HANDLER_HPP
#include "line.hpp"
#include "nntp_auth.hpp"
#include "storage.hpp"
#include <deque>
#include <string>

namespace nntpchan
{
class NNTPServerHandler : public LineReader, public IConnHandler
{
public:
  NNTPServerHandler(const fs::path &storage);
  ~NNTPServerHandler();

  virtual bool ShouldClose();

  void SetAuth(CredDB_ptr creds);

  virtual void OnData(const char *, ssize_t);

  void Greet();

protected:
  void HandleLine(const std::string &line);
  void HandleCommand(const std::deque<std::string> &command);

private:
  enum State
  {
    eStateReadCommand,
    eStateStoreArticle,
    eStateQuit
  };

private:
  void EnterState(State st);

  void ArticleObtained();

  // handle quit command, this queues a reply
  void Quit();

  // switch nntp modes, this queues a reply
  void SwitchMode(const std::string &mode);

  bool PostingAllowed();

private:
  std::string m_articleName;
  FileHandle_ptr m_article;
  CredDB_ptr m_auth;
  ArticleStorage_ptr m_store;
  std::string m_mode;
  bool m_authed;
  State m_state;
};
}

#endif
