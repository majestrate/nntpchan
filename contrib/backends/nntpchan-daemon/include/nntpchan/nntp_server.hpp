#ifndef NNTPCHAN_NNTP_SERVER_HPP
#define NNTPCHAN_NNTP_SERVER_HPP
#include "frontend.hpp"
#include "server.hpp"
#include <deque>
#include <string>

namespace nntpchan
{

class NNTPServer : public Server
{
public:
  NNTPServer(ev::Loop * loop);

  virtual ~NNTPServer();

  void SetStoragePath(const std::string &path);

  void SetLoginDB(const std::string path);

  void SetInstanceName(const std::string &name);

  std::string InstanceName() const;

  virtual IServerConn *CreateConn(int fd);

  virtual void OnAcceptError(int status);

  void SetFrontend(Frontend *f);

private:
  std::string m_logindbpath;
  std::string m_storagePath;
  std::string m_servername;

  Frontend_ptr m_frontend;
};

class NNTPServerConn : public IServerConn
{
public:
  NNTPServerConn(int fd, Server *parent, IConnHandler *h) : IServerConn(fd, parent, h) {}

  virtual bool IsTimedOut() { return false; };

  virtual void Greet();
};
}

#endif
