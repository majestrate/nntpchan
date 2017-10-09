
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
  }

  IServerConn * NNTPServer::CreateConn(uv_stream_t * s)
  {
    CredDB_ptr creds;

    std::ifstream i;
    i.open(m_logindbpath);
    if(i.is_open()) creds = std::make_shared<HashedFileDB>(m_logindbpath);

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

  void NNTPServer::SetInstanceName(const std::string & name)
  {
    m_servername = name;
  }

  void NNTPServer::SetFrontend(Frontend * f)
  {
    m_frontend.reset(f);
  }

  std::string NNTPServer::InstanceName() const
  {
    return m_servername;
  }

  void NNTPServer::OnAcceptError(int status)
  {
    std::cerr << "nntpserver::accept() " << uv_strerror(status) << std::endl;
  }

  void NNTPServerConn::SendNextReply()
  {
    IConnHandler * handler = GetHandler();
    while(handler->HasNextLine()) {
      auto line = handler->GetNextLine();
      SendString(line + "\r\n");
    }
  }


  void NNTPServerConn::Greet()
  {
    IConnHandler * handler = GetHandler();
    handler->Greet();
    SendNextReply();
  }



}
