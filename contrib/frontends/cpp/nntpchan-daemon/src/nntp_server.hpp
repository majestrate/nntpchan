#ifndef NNTPCHAN_NNTP_SERVER_HPP
#define NNTPCHAN_NNTP_SERVER_HPP
#include <uv.h>
#include <string>
#include <vector>
#include "storage.hpp"

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
    
  private:
    
    operator uv_handle_t * () { return (uv_handle_t*) &m_server; }
    operator uv_tcp_t * () { return &m_server; }
    operator uv_stream_t * () { return (uv_stream_t *) &m_server; }
    
    uv_tcp_t m_server;
    uv_loop_t * m_loop;
    
    std::vector<NNTPServerConn *> m_conns;

    std::string m_storagePath;
    
  };


    class NNTPServerHandler
  {
  public:

    enum State {
      eNNTPStateGreet,
      eNNTPStateHandshake,
      eNNTPStateReader,
      eNNTPStateStream,
      eNNTPStateTAKETHIS,
      eNNTPStateIHAVE,
      eNNTPStateARTICLE,
      eNNTPStatePOST
    };
    
    NNTPServerHandler(const std::string & storagepath);
    ~NNTPServerHandler();
    
    void OnData(const char * data, ssize_t s);

    bool HasNextLine();
    std::string GetNextLine();
    
  private:
    State m_state;
    ArticleStorage m_storage;
  };

  
  class NNTPServerConn
  {
  public:
    NNTPServerConn(uv_loop_t * l, uv_stream_t * s, const std::string & storage);
    virtual ~NNTPServerConn();

    void Close();

    void Quit();

    void SendLine(const std::string & line);
    void SendCode(const int code, const std::string & message);

    void ProcessData(const char * d, ssize_t l);
    
    void SendNextReply();
    
  private:

    void SendString(const std::string & line);
    
    
    operator uv_stream_t * () { return (uv_stream_t *) &m_conn; }
    
    uv_tcp_t m_conn;

    NNTPServerHandler m_handler;
    
    char m_readbuff[1024];
    
  };
}

#endif
