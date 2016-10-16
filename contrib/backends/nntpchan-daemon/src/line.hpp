#ifndef NNTPCHAN_LINE_HPP
#define NNTPCHAN_LINE_HPP
#include <string>
#include <deque>
namespace nntpchan
{

  /** @brief a buffered line reader */
  class LineReader
  {
  public:
    

    /** @brief queue inbound data from connection */
    void OnData(const char * data, ssize_t s);

    /** @brief do we have line to send to the client? */
    bool HasNextLine();
    /** @brief get the next line to send to the client, does not check if it exists */
    std::string GetNextLine();

  protected:
    /** @brief handle a line from the client */
    virtual void HandleLine(const std::string & line) = 0;
    /** @brief queue the next line to send to the client */
    void QueueLine(const std::string & line);
    
  private:
    void OnLine(const char * d, const size_t l);
    // lines to send
    std::deque<std::string> m_sendlines;
  };
}

#endif
