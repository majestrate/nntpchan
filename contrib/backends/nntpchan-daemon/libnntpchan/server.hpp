#ifndef NNTPCHAN_SERVER_HPP
#define NNTPCHAN_SERVER_HPP
#include <uv.h>
#include <deque>
#include <functional>
#include <string>

namespace nntpchan
{

  class Server;


  struct IConnHandler
  {

    virtual ~IConnHandler() {};

    /** got inbound data */
    virtual void OnData(const char * data, ssize_t s) = 0;

    /** get next line of data to send */
    std::string GetNextLine();

    /** return true if we have a line to send */
    bool HasNextLine();

    /** return true if we should close this connection otherwise return false */
    virtual bool ShouldClose() = 0;

    /** queue a data send */
    void QueueLine(const std::string & line);

    virtual void Greet() = 0;


  private:
    std::deque<std::string> m_sendlines;
  };

  /** server connection handler interface */
  struct IServerConn
  {
    IServerConn(uv_loop_t * l, uv_stream_t * s, Server * parent, IConnHandler * h);
    virtual ~IServerConn();
    virtual void Close();
    virtual void Greet() = 0;
    virtual void SendNextReply() = 0;
    virtual bool IsTimedOut() = 0;
    void SendString(const std::string & str);
    Server * Parent() { return m_parent; };
    IConnHandler * GetHandler() { return m_handler; };
    uv_loop_t * GetLoop() { return m_loop; };
  private:
    uv_tcp_t m_conn;
    uv_loop_t * m_loop;
    Server * m_parent;
    IConnHandler * m_handler;
    char m_readbuff[65536];
  };

  class Server
  {
  public:
    Server(uv_loop_t * loop);
    /** called after socket close, NEVER call directly */
    virtual ~Server() {}
    /** create connection handler from open stream */
    virtual IServerConn * CreateConn(uv_stream_t * s) = 0;
    /** close all sockets and stop */
    void Close();
    /** bind to address */
    void Bind(const std::string & addr);

    typedef std::function<void(IServerConn *)> ConnVisitor;

    /** visit all open connections */
    void VisitConns(ConnVisitor v);

    /** remove connection from server, called after proper close */
    void RemoveConn(IServerConn * conn);

  protected:
    uv_loop_t * GetLoop() { return m_loop; }
    virtual void OnAcceptError(int status) = 0;
  private:
    operator uv_handle_t * () { return (uv_handle_t*) &m_server; }
    operator uv_tcp_t * () { return &m_server; }
    operator uv_stream_t * () { return (uv_stream_t *) &m_server; }

    void OnAccept(uv_stream_t * s, int status);
    std::deque<IServerConn *> m_conns;
    uv_tcp_t m_server;
    uv_loop_t * m_loop;
  };
}


#endif
