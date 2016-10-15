#ifndef NNTPCHAN_NNTP_HANDLER_HPP
#define NNTPCHAN_NNTP_HANDLER_HPP
#include <deque>
#include <string>
#include "line.hpp"
#include "nntp_auth.hpp"
#include "storage.hpp"

namespace nntpchan
{
  class NNTPServerHandler : public LineReader
  {
  public:
    NNTPServerHandler(const std::string & storage);
    ~NNTPServerHandler();
    
    bool Done();

    void SetAuth(NNTPCredentialDB * creds);

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
    // handle quit command, this queues a reply
    void Quit();

    // switch nntp modes, this queues a reply
    void SwitchMode(const std::string & mode);

    bool PostingAllowed();
    
  private:
    NNTPCredentialDB * m_auth;
    ArticleStorage m_store;
    std::string m_mode;
    bool m_authed;
    State m_state;
  };
}


#endif
