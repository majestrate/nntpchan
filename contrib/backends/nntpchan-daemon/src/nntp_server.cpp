#include "buffer.hpp"
#include "nntp_server.hpp"
#include "nntp_auth.hpp"
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
    m_server.data = this;
  }

  NNTPServer::~NNTPServer()
  {
    if (m_frontend) delete m_frontend;
  }

  void NNTPServer::Close()
  {
    uv_close((uv_handle_t*)&m_server, [](uv_handle_t * s) {
        NNTPServer * self = (NNTPServer*)s->data;
        if (self) delete self;
        s->data = nullptr;
    });
  }

  void NNTPServer::Bind(const std::string & addr)
  {
    auto saddr = ParseAddr(addr);
    assert(uv_tcp_bind(*this, saddr, 0) == 0);
    std::cerr << "nntp server bound to " << saddr.to_string() << std::endl;
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
    NNTPCredentialDB * creds = nullptr;

    std::ifstream i;
    i.open(m_logindbpath);
    if(i.is_open()) creds = new HashedFileDB(m_logindbpath);
    
    NNTPServerConn * conn = new NNTPServerConn(m_loop, s, m_storagePath, creds);
    conn->Greet();
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
  
  NNTPServerConn::NNTPServerConn(uv_loop_t * l, uv_stream_t * s, const std::string & storage, NNTPCredentialDB * creds) :
    m_handler(storage)
  {
    m_handler.SetAuth(creds);
    uv_tcp_init(l, &m_conn);
    m_conn.data = this;
    uv_accept(s, (uv_stream_t*) &m_conn);
    uv_read_start((uv_stream_t*) &m_conn, [] (uv_handle_t * h, size_t s, uv_buf_t * b) {
        NNTPServerConn * self = (NNTPServerConn*) h->data;
        if(self == nullptr) return;
        b->base = self->m_readbuff;
        if (s > sizeof(self->m_readbuff))
          b->len = sizeof(self->m_readbuff);
        else
          b->len = s;
      }, [] (uv_stream_t * s, ssize_t nread, const uv_buf_t * b) {
        NNTPServerConn * self = (NNTPServerConn*) s->data;
        if(self == nullptr) return;
        if(nread > 0) {
          self->m_handler.OnData(b->base, nread);
          self->SendNextReply();
          if(self->m_handler.Done())
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
  
  void NNTPServerConn::SendNextReply()
  {
    if(m_handler.HasNextLine()) {
      auto line = m_handler.GetNextLine();
      SendString(line+"\n");
    }
  }


  void NNTPServerConn::Greet()
  {
    m_handler.Greet();
    SendNextReply();
  }
  
  void NNTPServerConn::SendString(const std::string & str)
  {
    WriteBuffer * b = new WriteBuffer(str);
    uv_write(&b->w, *this, &b->b, 1, [](uv_write_t * w, int status) {
        (void) status;
        WriteBuffer * wb = (WriteBuffer *) w->data;
        if(wb)
          delete wb;
    });
  }

  void NNTPServerConn::Close()
  {
    uv_close((uv_handle_t*)&m_conn, [] (uv_handle_t * s) {
        NNTPServerConn * self = (NNTPServerConn*) s->data;
        if(self)
          delete self;
        s->data = nullptr;
    });
  }  
}
