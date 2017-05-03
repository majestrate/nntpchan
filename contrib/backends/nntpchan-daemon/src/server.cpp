#include "server.hpp"
#include "net.hpp"
#include <cassert>

namespace nntpchan
{
  Server::Server(uv_loop_t * loop)
  {
    m_loop = loop;
    uv_tcp_init(m_loop, &m_server);
    m_server.data = this;
  }

  void Server::Close()
  {
    uv_close((uv_handle_t*)&m_server, [](uv_handle_t * s) {
      Server * self = (Server*)s->data;
      if (self) delete self;
      s->data = nullptr;
    });
  }

  void Server::Bind(const std::string & addr)
  {
    auto saddr = ParseAddr(addr);
    assert(uv_tcp_bind(*this, saddr, 0) == 0);
    auto cb = [] (uv_stream_t * s, int status) {
      Server * self = (Server *) s->data;
      self->OnAccept(s, status);
    };
    assert(uv_listen(*this, 5, cb) == 0);
  }

  void Server::OnAccept(uv_stream_t * s, int status)
  {
    if(status < 0) {
      OnAcceptError(status);
      return;
    }
    IServerConn * conn = CreateConn(s);
    assert(conn);
    conn->Greet();
  }

  void IConnHandler::QueueLine(const std::string & line)
  {
    m_sendlines.push_back(line);
  }

  bool IConnHandler::HasNextLine()
  {
    return m_sendlines.size() > 0;
  }

  std::string IConnHandler::GetNextLine()
  {
    std::string line = m_sendlines[0];
    m_sendlines.pop_front();
    return line;
  }
}
