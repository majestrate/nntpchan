#ifndef NNTPCHAN_NNTP_HANDLER_HPP
#define NNTPCHAN_NNTP_HANDLER_HPP
#include <deque>
#include <string>
#include "line.hpp"
#include "nntp_auth.hpp"
#include "storage.hpp"

namespace nntpchan
{
  class NNTPServerHandler : public LineReader, public IConnHandler
  {
  public:
    NNTPServerHandler(const std::string & storage);
    ~NNTPServerHandler();

    virtual bool ShouldClose();

    void SetAuth(NNTPCredentialDB * creds);

    virtual void OnData(const char *, ssize_t);

    void Greet();

  protected:
    void HandleLine(const std::string & line);
    void HandleCommand(const std::deque<std::string> & command);
  private:

    enum State {
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
    void SwitchMode(const std::string & mode);

    bool PostingAllowed();

  private:
    std::string m_articleName;
    std::fstream * m_article;
    NNTPCredentialDB * m_auth;
    ArticleStorage m_store;
    std::string m_mode;
    bool m_authed;
    State m_state;
  };
}


#endif
