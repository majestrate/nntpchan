#ifndef NNTPCHAN_EXEC_FRONTEND_HPP
#define NNTPCHAN_EXEC_FRONTEND_HPP
#include "frontend.hpp"
#include <deque>

namespace nntpchan
{
  class ExecFrontend : public Frontend
  {
  public:

    ExecFrontend(const std::string & exe);
    
    ~ExecFrontend();

    void ProcessNewMessage(const std::string & fpath);
    bool AcceptsNewsgroup(const std::string & newsgroup);
    bool AcceptsMessage(const std::string & msgid);

  private:

    int Exec(std::deque<std::string> args);
    
  private:
    std::string m_exec;
    
  };
}

#endif
