#ifndef NNTPCHAN_LINE_HPP
#define NNTPCHAN_LINE_HPP
#include "server.hpp"
#include <stdint.h>
namespace nntpchan
{

  /** @brief a buffered line reader */
  class LineReader
  {
  public:

    LineReader(size_t lineLimit);

    /** @brief queue inbound data from connection */
    void Data(const char * data, ssize_t s);

    /** implements IConnHandler */
    virtual bool ShouldClose();

  protected:
    /** @brief handle a line from the client */
    virtual void HandleLine(const std::string & line) = 0;


  private:
    void OnLine(const char * d, const size_t l);
    std::string m_leftovers;
    bool m_close;
    const size_t lineLimit;
  };
}

#endif
