#include "buffer.hpp"
#include "server.hpp"
#include "net.hpp"
#include <cassert>
#include <iostream>

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
    std::cout << "Close server" << std::endl;
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
    m_conns.push_back(conn);
    conn->Greet();
  }

  void Server::RemoveConn(IServerConn * conn)
  {
    auto itr = m_conns.begin();
    while(itr != m_conns.end())
      {
        if(*itr == conn)
          itr = m_conns.erase(itr);
        else
          ++itr;
      }
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

  IServerConn::IServerConn(uv_loop_t * l, uv_stream_t * st, Server * parent, IConnHandler * h)
  {
    m_loop = l;
    m_parent = parent;
    m_handler = h;
    uv_tcp_init(l, &m_conn);
    m_conn.data = this;
    uv_accept(st, (uv_stream_t*) &m_conn);
    uv_read_start((uv_stream_t*) &m_conn, [] (uv_handle_t * h, size_t s, uv_buf_t * b) {
        IServerConn * self = (IServerConn*) h->data;
        if(self == nullptr) return;
        b->base = self->m_readbuff;
        if (s > sizeof(self->m_readbuff))
          b->len = sizeof(self->m_readbuff);
        else
          b->len = s;
      }, [] (uv_stream_t * s, ssize_t nread, const uv_buf_t * b) {
        IServerConn * self = (IServerConn*) s->data;
        if(self == nullptr) return;
        if(nread > 0) {
          self->m_handler->OnData(b->base, nread);
          self->SendNextReply();
          if(self->m_handler->ShouldClose())
            self->Close();
        } else {
          if (nread != UV_EOF) {
            std::cerr << "error in nntp server conn alloc: ";
            std::cerr << uv_strerror(nread);
            std::cerr << std::endl;
          }
          // got eof or error
          self->Close();
        }
      });
  }

  IServerConn::~IServerConn()
  {
    delete m_handler;
  }

  void IServerConn::SendString(const std::string & str)
  {
    WriteBuffer * b = new WriteBuffer(str);
    uv_write(&b->w, (uv_stream_t*)&m_conn, &b->b, 1, [](uv_write_t * w, int status) {
        (void) status;
        WriteBuffer * wb = (WriteBuffer *) w->data;
        if(wb)
          delete wb;
      });
  }

  void IServerConn::Close()
  {
    m_parent->RemoveConn(this);
    uv_close((uv_handle_t*)&m_conn, [] (uv_handle_t * s) {
        IServerConn * self = (IServerConn*) s->data;
        if(self)
          delete self;
        s->data = nullptr;
      });
  }
}
