#include "buffer.hpp"
#include "nntp_server.hpp"
#include "net.hpp"
#include <cassert>
#include <iostream>
#include <sstream>

namespace nntpchan
{
  NNTPServer::NNTPServer(uv_loop_t * loop)
  {
    uv_tcp_init(loop, &m_server);
    m_loop = loop;
  }

  NNTPServer::~NNTPServer()
  {
    uv_close((uv_handle_t*)&m_server, [](uv_handle_t *) {});
  }

  void NNTPServer::Bind(const std::string & addr)
  {
    auto saddr = ParseAddr(addr);
    assert(uv_tcp_bind(*this, saddr, 0) == 0);
    std::cerr << "nntp server bound to " << saddr.to_string() << std::endl;
    m_server.data = this;
    auto cb = [] (uv_stream_t * s, int status) {
      NNTPServer * self = (NNTPServer *) s->data;
      self->OnAccept(s, status);
    };
    
    assert(uv_listen(*this, 5, cb) == 0);
  }

  void NNTPServer::OnAccept(uv_stream_t * s, int status)
  {
    if(status < 0) {
      std::cerr << "nntp server OnAccept fail: " << uv_strerror(status) << std::endl;
      return;
    }
    NNTPServerConn * conn = new NNTPServerConn(m_loop, s, m_storagePath);
    conn->SendCode(200, "Posting Allowed");
  }


  void NNTPServer::SetStoragePath(const std::string & path)
  {
    m_storagePath = path;
  }
  
  NNTPServerConn::NNTPServerConn(uv_loop_t * l, uv_stream_t * s, const std::string & storage) :
    m_handler(storage)
  {
    uv_tcp_init(l, &m_conn);
    m_conn.data = this;
    uv_accept(s, (uv_stream_t*) &m_conn);
    uv_read_start((uv_stream_t*) &m_conn, [] (uv_handle_t * h, size_t s, uv_buf_t * b) {
        NNTPServerConn * self = (NNTPServerConn*) h->data;
        b->base = self->m_readbuff;
        if (s > sizeof(self->m_readbuff))
          b->len = sizeof(self->m_readbuff);
        else
          b->len = s;
      }, [] (uv_stream_t * s, ssize_t nread, const uv_buf_t * b) {
        NNTPServerConn * self = (NNTPServerConn*) s->data;
        if(nread > 0) {
          self->ProcessData(b->base, nread);
          self->SendNextReply();
        } else {
          if (nread != UV_EOF) {
            std::cerr << "error in nntp server conn alloc: ";
            std::cerr << uv_strerror(nread);
            std::cerr << std::endl;
          }
           
          delete self;
          s->data = nullptr;
          
        }
      });
  }

  NNTPServerConn::~NNTPServerConn()
  {
    uv_close((uv_handle_t*)&m_conn, [] (uv_handle_t *) {});
  }

  void NNTPServerConn::ProcessData(const char *d, ssize_t l)
  {
    m_handler.OnData(d, l);
  }

  void NNTPServerConn::SendNextReply()
  {
    if(m_handler.HasNextLine()) {
      auto line = m_handler.GetNextLine();
      SendLine(line);
    }
  }
  
  void NNTPServerConn::SendCode(const int code, const std::string & msg)
  {
    std::stringstream ss;
    ss << code << " " << msg << std::endl;
    SendString(ss.str());
  }

  void NNTPServerConn::SendString(const std::string & line)
  {
    WriteBuffer * b = new WriteBuffer(line);
    uv_write(&b->w, *this, &b->b, 1, [](uv_write_t * w, int status) {
        WriteBuffer * wb = (WriteBuffer *) w->data;
        delete wb;
    });
  }

  NNTPServerHandler::NNTPServerHandler(const std::string & storagepath) :
    m_state(eNNTPStateGreet)
  {
    m_storage.SetPath(storagepath);
  }

  NNTPServerHandler::~NNTPServerHandler()
  {
  }

  void NNTPServerHandler::OnData(const char * d, ssize_t l)
  {
    
  }

}
