#ifndef NNTPCHAN_NNTP_AUTH_HPP
#define NNTPCHAN_NNTP_AUTH_HPP
#include <string>
#include <iostream>
#include <fstream>
#include <mutex>
#include "line.hpp"

namespace nntpchan
{
  /** @brief nntp credential db interface */
  class NNTPCredentialDB
  {
  public:
    /** @brief open connection to database, return false on error otherwise return true */
    virtual bool Open() = 0;
    /** @brief close connection to database */
    virtual void Close() = 0;
    /** @brief return true if username password combo is correct */
    virtual bool CheckLogin(const std::string & user, const std::string & passwd) = 0;
    virtual ~NNTPCredentialDB() {}
  };

  /** @brief nntp credential db using hashed+salted passwords */
  class HashedCredDB : public NNTPCredentialDB, public LineReader
  {
  public:
    HashedCredDB();
    bool CheckLogin(const std::string & user, const std::string & passwd);
  protected:
    void SetStream(std::istream * i);

    std::string Hash(const std::string & data, const std::string & salt);
    void HandleLine(const std::string & line);
  private:
    bool ProcessLine(const std::string & line);

    std::mutex m_access;
    std::string m_user, m_passwd;
    bool m_found;
    /** return true if we have a line that matches this username / password combo */
    std::istream * m_instream;
  };

  class HashedFileDB : public HashedCredDB
  {
  public:
    HashedFileDB(const std::string & fname);
    ~HashedFileDB();
    bool Open();
    void Close();
  private:
    std::string m_fname;
    std::ifstream f;
  };
}

#endif
