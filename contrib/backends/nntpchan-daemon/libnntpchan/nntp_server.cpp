
#include <cassert>
#include <cstring>
#include <iostream>
#include <nntpchan/net.hpp>
#include <nntpchan/nntp_auth.hpp>
#include <nntpchan/nntp_handler.hpp>
#include <nntpchan/nntp_server.hpp>
#include <sstream>

namespace nntpchan
{

NNTPServer::NNTPServer(Mainloop & loop) : Server(loop), m_frontend(nullptr) {}

NNTPServer::~NNTPServer() {}

IServerConn *NNTPServer::CreateConn(int f)
{
  CredDB_ptr creds;

  std::ifstream i;
  i.open(m_logindbpath);
  if (i.is_open())
    creds = std::make_shared<HashedFileDB>(m_logindbpath);

  NNTPServerHandler *handler = new NNTPServerHandler(m_storagePath);
  if (creds)
    handler->SetAuth(creds);

  return new NNTPServerConn(f, this, handler);
}

void NNTPServer::SetLoginDB(const std::string path) { m_logindbpath = path; }

void NNTPServer::SetStoragePath(const std::string &path) { m_storagePath = path; }

void NNTPServer::SetInstanceName(const std::string &name) { m_servername = name; }

void NNTPServer::SetFrontend(Frontend *f) { m_frontend.reset(f); }

std::string NNTPServer::InstanceName() const { return m_servername; }

void NNTPServer::OnAcceptError(int status) { std::cerr << "nntpserver::accept() " << strerror(status) << std::endl; }


void NNTPServerConn::Greet()
{
  IConnHandler *handler = GetHandler();
  handler->Greet();
}
}
