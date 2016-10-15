#ifndef NNTPCHAN_NNTP_AUTH_HPP
#define NNTPCHAN_NNTP_AUTH_HPP
#include <string>
#include <iostream>
#include <mutex>
#include "line.hpp"

namespace nntpchan
{
  /** @brief nntp credential db interface */
  class NNTPCredentialDB
  {
  public:
    /** @brief return true if username password combo is correct */
    virtual bool CheckLogin(const std::string & user, const std::string & passwd) = 0;
  };
  
  /** @brief nntp credential db using hashed+salted passwords */
  class HashedCredDB : public NNTPCredentialDB, public LineReader
  {
  public:
    HashedCredDB(std::istream & i);
    bool CheckLogin(const std::string & user, const std::string & passwd);
  protected:
    std::string Hash(const std::string & data, const std::string & salt);
  private:

    bool ProcessLine(const std::string & line);
    
    std::mutex m_access;
    std::string m_user, m_passwd;
    bool m_found;
    /** return true if we have a line that matches this username / password combo */
    std::istream & m_instream;
  };
}

#endif
