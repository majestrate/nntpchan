
#include "nntp_server.hpp"
#include "nntp_auth.hpp"
#include "nntp_handler.hpp"
#include "net.hpp"
#include <cassert>
#include <iostream>
#include <sstream>

namespace nntpchan
{

  NNTPServer::NNTPServer(uv_loop_t * loop) : Server(loop), m_frontend(nullptr) {}

  NNTPServer::~NNTPServer()
  {
    if (m_frontend) delete m_frontend;
  }

  IServerConn * NNTPServer::CreateConn(uv_stream_t * s)
  {
    NNTPCredentialDB * creds = nullptr;

    std::ifstream i;
    i.open(m_logindbpath);
    if(i.is_open()) creds = new HashedFileDB(m_logindbpath);

    NNTPServerHandler * handler = new NNTPServerHandler(m_storagePath);
    if(creds)
      handler->SetAuth(creds);

    NNTPServerConn * conn = new NNTPServerConn(GetLoop(), s, this, handler);
    return conn;
  }

  void NNTPServer::SetLoginDB(const std::string path)
  {
    m_logindbpath = path;
  }


  void NNTPServer::SetStoragePath(const std::string & path)
  {
    m_storagePath = path;
  }

  void NNTPServer::SetFrontend(Frontend * f)
  {
    if(m_frontend) delete m_frontend;
    m_frontend = f;
  }

  void NNTPServer::OnAcceptError(int status)
  {
    std::cerr << "nntpserver::accept() " << uv_strerror(status) << std::endl;
  }

  void NNTPServerConn::SendNextReply()
  {
    IConnHandler * handler = GetHandler();
    if(handler->HasNextLine()) {
      auto line = handler->GetNextLine();
      SendString(line + "\n");
    }
  }


  void NNTPServerConn::Greet()
  {
    IConnHandler * handler = GetHandler();
    handler->Greet();
    SendNextReply();
  }



}
