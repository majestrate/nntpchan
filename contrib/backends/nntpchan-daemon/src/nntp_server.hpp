#ifndef NNTPCHAN_NNTP_SERVER_HPP
#define NNTPCHAN_NNTP_SERVER_HPP
#include <uv.h>
#include <string>
#include <deque>
#include "storage.hpp"
#include "frontend.hpp"
#include "nntp_auth.hpp"
#include "nntp_handler.hpp"

namespace nntpchan
{
  
  class NNTPServerConn;
  
  class NNTPServer
  {
  public:
    NNTPServer(uv_loop_t * loop);
    ~NNTPServer();

    void Bind(const std::string & addr);

    void OnAccept(uv_stream_t * s, int status);

    void SetStoragePath(const std::string & path);

    void SetLoginDB(const std::string path);

    void SetFrontend(Frontend * f);

    void Close();
    
  private:
    
    operator uv_handle_t * () { return (uv_handle_t*) &m_server; }
    operator uv_tcp_t * () { return &m_server; }
    operator uv_stream_t * () { return (uv_stream_t *) &m_server; }
    
    uv_tcp_t m_server;
    uv_loop_t * m_loop;
    
    std::deque<NNTPServerConn *> m_conns;

    std::string m_logindbpath;
    std::string m_storagePath;

    Frontend * m_frontend;
    
  };
  
  class NNTPServerConn
  {
  public:
    NNTPServerConn(uv_loop_t * l, uv_stream_t * s, const std::string & storage, NNTPCredentialDB * creds);
    /** @brief close connection, this connection cannot be used after calling this */
    void Close();
    
    /** @brief send next queued reply */
    void SendNextReply();

    void Greet();
    
  private:

    void SendString(const std::string & line);
    
    
    operator uv_stream_t * () { return (uv_stream_t *) &m_conn; }
    
    uv_tcp_t m_conn;

    NNTPServerHandler m_handler;

    char m_readbuff[1028];
    
  };
}

#endif
