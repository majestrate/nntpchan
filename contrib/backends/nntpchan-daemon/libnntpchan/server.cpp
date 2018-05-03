#include <cassert>
#include <iostream>
#include <nntpchan/buffer.hpp>
#include <nntpchan/net.hpp>
#include <nntpchan/server.hpp>

namespace nntpchan
{

Server::Server(Mainloop & loop) : ev::io(-1), m_Loop(loop)
{
}

void Server::close()
{
  auto itr = m_conns.begin();
  while(itr != m_conns.end())
  {
    itr = m_conns.erase(itr);
  }
  m_Loop.UntrackConn(this);
  ev::io::close();
}
bool Server::Bind(const std::string &addr)
{
  auto saddr = ParseAddr(addr);
  return m_Loop.BindTCP(saddr, this);
}

void Server::OnAccept(int f, int status)
{
  if (status)
  {
    OnAcceptError(status);
    return;
  }
  IServerConn *conn = CreateConn(f);
  
  if(m_Loop.TrackConn(conn))
  { 
    m_conns.push_back(conn);  
    conn->Greet();
    conn->write();
  }
  else 
  {
    std::cout << "accept track conn failed" << std::endl;
    conn->close();
    delete conn;
  }
}

int Server::accept()
{
  int res = ::accept4(fd, nullptr, nullptr, SOCK_NONBLOCK);
  if(res == -1) return res;
  OnAccept(res, errno);
  return res;
}

void Server::RemoveConn(IServerConn *conn)
{
  auto itr = m_conns.begin();
  while (itr != m_conns.end())
  {
    if (*itr == conn)
      itr = m_conns.erase(itr);
    else
      ++itr;
  }
  m_Loop.UntrackConn(conn);
}

void IConnHandler::QueueLine(const std::string &line) { m_sendlines.push_back(line+"\r\n"); }

bool IConnHandler::HasNextLine() { return m_sendlines.size() > 0; }

std::string IConnHandler::GetNextLine()
{
  std::string line = m_sendlines[0];
  m_sendlines.pop_front();
  return line;
}

IServerConn::IServerConn(int fd, Server *parent, IConnHandler *h) : ev::io(fd), m_parent(parent), m_handler(h)
{
}

IServerConn::~IServerConn() { delete m_handler; }

int IServerConn::read(char * buf, size_t sz)
{
  ssize_t readsz = ::read(fd, buf, sz);
  if(readsz > 0)
  {
    m_handler->OnData(buf, readsz);
  }
  return readsz;
}

bool IServerConn::keepalive()
{
  return !m_handler->ShouldClose();
}

int IServerConn::write()
{
  auto leftovers = m_writeLeftover.size();
  ssize_t written;
  if(leftovers)
  {
    if(leftovers > 1024)
    {
      leftovers = 1024;
    }
    written = ::write(fd, m_writeLeftover.c_str(), leftovers);
    if(written > 0)
    {
      m_writeLeftover = m_writeLeftover.substr(written);
    }
    else 
    {
      // too much leftovers
      return -1;
    }
  }
  do
  {
    if(!m_handler->HasNextLine())
    {
      return 0;
    }
    auto line = m_handler->GetNextLine();  
    written = ::write(fd, line.c_str(), line.size());
    if(written > 0)
    {
      m_writeLeftover = line.substr(written);
    }
    else
    {
      m_writeLeftover = line;
      return -1;
    }
  }
  while(written > 0);
  return 0;
}

void IServerConn::close()
{
  m_parent->RemoveConn(this);
  ev::io::close();
}
}
