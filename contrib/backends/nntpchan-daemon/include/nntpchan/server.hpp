#ifndef NNTPCHAN_SERVER_HPP
#define NNTPCHAN_SERVER_HPP
#include <deque>
#include <functional>
#include <string>
#include <nntpchan/event.hpp>

namespace nntpchan
{

class Server;

struct IConnHandler
{

  virtual ~IConnHandler(){};

  /** got inbound data */
  virtual void OnData(const char *data, ssize_t s) = 0;

  /** get next line of data to send */
  std::string GetNextLine();

  /** return true if we have a line to send */
  bool HasNextLine();

  /** return true if we should close this connection otherwise return false */
  virtual bool ShouldClose() = 0;

  /** queue a data send */
  void QueueLine(const std::string &line);

  virtual void Greet() = 0;

private:
  std::deque<std::string> m_sendlines;
};

/** server connection handler interface */
struct IServerConn : public ev::io
{
  IServerConn(int fd, Server *parent, IConnHandler *h);
  virtual ~IServerConn();
  virtual int read(char * buf, size_t sz);
  virtual int write();
  virtual void close();
  virtual void Greet() = 0;
  virtual bool IsTimedOut() = 0;
  virtual bool keepalive() ;
  Server *Parent() { return m_parent; };
  IConnHandler *GetHandler() { return m_handler; };

private:
  Server *m_parent;
  IConnHandler *m_handler;
  std::string m_writeLeftover;
};

class Server : public ev::io
{
public:
  Server(Mainloop & loop);
  virtual ~Server() {};

  virtual bool acceptable() const { return true; };
  virtual void close();
  virtual bool readable() const { return false; };
  virtual int read(char *,size_t) { return -1; };
  virtual bool writeable() const { return false; };
  virtual int write() {return -1; };
  virtual int accept();
  virtual bool keepalive() { return true; };


  /** create connection handler from open stream */
  virtual IServerConn *CreateConn(int fd) = 0;
  /** bind to address */
  bool Bind(const std::string &addr);

  typedef std::function<void(IServerConn *)> ConnVisitor;

  /** visit all open connections */
  void VisitConns(ConnVisitor v);

  /** remove connection from server, called after proper close */
  void RemoveConn(IServerConn *conn);

protected:
  virtual void OnAcceptError(int status) = 0;

private:

  void OnAccept(int fd, int status);
  Mainloop & m_Loop;
  std::deque<IServerConn *> m_conns;
};
}

#endif
