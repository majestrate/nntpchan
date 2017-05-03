#include "nntp_handler.hpp"
#include "message.hpp"
#include <algorithm>
#include <cctype>
#include <cstring>
#include <string>
#include <sstream>
#include <iostream>

namespace nntpchan
{
  NNTPServerHandler::NNTPServerHandler(const std::string & storage) :
    LineReader(1024),
    m_article(nullptr),
    m_auth(nullptr),
    m_store(storage),
    m_authed(false),
    m_state(eStateReadCommand)
  {
  }

  NNTPServerHandler::~NNTPServerHandler()
  {
    if(m_auth) delete m_auth;
  }

  void NNTPServerHandler::HandleLine(const std::string &line)
  {
    if(m_state == eStateReadCommand)
    {
      std::deque<std::string> command;
      std::istringstream s;
      s.str(line);
      for (std::string part; std::getline(s, part, ' '); ) {
          if(part.size()) command.push_back(std::string(part));
      }
      if(command.size())
        HandleCommand(command);
      else
        QueueLine("501 Syntax error");
    }
    else if(m_state == eStateStoreArticle)
    {
      std::string l = line + "\r\n";
      OnData(l.c_str(), l.size());
    }
    else
    {
      std::cerr << "invalid state" << std::endl;
    }
  }

  void NNTPServerHandler::OnData(const char * data, ssize_t l)
  {
    if(l <= 0 ) return;
    if(m_state == eStateStoreArticle)
    {
      const char * end = strstr(data, "\r\n.\r\n");
      if(end)
      {
        std::size_t diff = end - data ;
        if(m_article)
          m_article->write(data, diff+2);
        ArticleObtained();
        diff += 5;
        Data(end+5, l-diff);
        return;
      }
      if(m_article)
        m_article->write(data, l);
    }
    else
      Data(data, l);
  }

  void NNTPServerHandler::HandleCommand(const std::deque<std::string> & command)
  {
    auto cmd = command[0];
    std::transform(cmd.begin(), cmd.end(), cmd.begin(), ::toupper);
    std::size_t cmdlen = command.size();
    for(const auto & part : command)
      std::cerr << " " << part;
    std::cerr << std::endl;
    if (cmd == "QUIT") {
      Quit();
      return;
    }
    else if (cmd[0] == '5')
    {
        return;
    }
    else if (cmd == "MODE" ) {
      if(cmdlen == 2) {
        // set mode
        SwitchMode(command[1]);
      } else if(cmdlen) {
        // too many arguments
        QueueLine("500 too many arguments");
      } else {
        // get mode
        QueueLine("500 wrong arguments");
      }
    } else if(cmd == "CAPABILITIES") {
      QueueLine("101 I support the following:");
      QueueLine("READER");
      QueueLine("IMPLEMENTATION nntpchan-daemon");
      QueueLine("VERSION 2");
      QueueLine("STREAMING");
      QueueLine(".");
    } else if (cmd == "CHECK") {
      if(cmdlen == 2) {
        const std::string & msgid = command[1];
        if(IsValidMessageID(msgid) && m_store.Accept(msgid))
        {
          QueueLine("238 "+msgid);
          return;
        }
        QueueLine("438 "+msgid);
      }
      else
        QueueLine("501 syntax error");
    } else if (cmd == "TAKETHIS") {
      if (cmdlen == 2)
      {
        const std::string & msgid = command[1];
        if(m_store.Accept(msgid))
        {
          m_article = m_store.OpenWrite(msgid);
        }
        m_articleName = msgid;
        EnterState(eStateStoreArticle);
        return;
      }
      QueueLine("501 invalid syntax");
    } else {
      // unknown command
      QueueLine("500 Unknown Command");
    }
  }

  void NNTPServerHandler::ArticleObtained()
  {
    if(m_article)
    {
      m_article->flush();
      m_article->close();
      delete m_article;
      m_article = nullptr;
      QueueLine("239 "+m_articleName);
      std::cerr << "stored " << m_articleName << std::endl;
    }
    else
      QueueLine("439 "+m_articleName);
    m_articleName = "";
    EnterState(eStateReadCommand);
  }

  void NNTPServerHandler::SwitchMode(const std::string & mode)
  {
    std::string m = mode;
    std::transform(m.begin(), m.end(), m.begin(), ::toupper);
    if (m == "READER") {
      m_mode = m;
      if(PostingAllowed()) {
        QueueLine("200 Posting is permitted yo");
      } else {
        QueueLine("201 Posting is not permitted yo");
      }
    } else if (m == "STREAM") {
      m_mode = m;
      if (PostingAllowed()) {
        QueueLine("203 Streaming enabled");
      } else {
        QueueLine("483 Streaming Denied");
      }
    } else {
      // unknown mode
      QueueLine("500 Unknown mode");
    }
  }

  void NNTPServerHandler::EnterState(State st)
  {
    std::cerr << "enter state " << st << std::endl;
    m_state = st;
  }

  void NNTPServerHandler::Quit()
  {
    EnterState(eStateQuit);
    QueueLine("205 quitting");
  }

  bool NNTPServerHandler::ShouldClose()
  {
    return m_state == eStateQuit;
  }

  bool NNTPServerHandler::PostingAllowed()
  {
    return m_authed || m_auth == nullptr;
  }

  void NNTPServerHandler::Greet()
  {
    if(PostingAllowed())
      QueueLine("200 Posting allowed");
    else
      QueueLine("201 Posting not allowed");
  }

  void NNTPServerHandler::SetAuth(NNTPCredentialDB *creds)
  {
    if(m_auth) delete m_auth;
    m_auth = creds;
  }
}
