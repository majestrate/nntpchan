#include "nntp_auth.hpp"
#include "crypto.hpp"
#include "base64.hpp"
#include <array>
#include <iostream>
#include <fstream>

namespace nntpchan
{
  HashedCredDB::HashedCredDB(std::istream & i) :
    m_instream(i) {}
  
  bool HashedCredDB::CheckLogin(const std::string & user, const std::string & passwd)
  {
    std::unique_lock<std::mutex> lock(m_access);
    m_found = false;
    m_user = user;
    m_passwd = passwd;
    m_instream.seekg(0, std::ios::end);
    const auto l = m_instream.tellg();
    m_instream.seekg(0, std::ios::beg);
    char * buff = new char[l];
    // read file
    m_instream.read(buff, l);
    OnData(buff, l);
    delete [] buff;
    return m_found;
  }

  bool HashedCredDB::ProcessLine(const std::string & line)
  {
    // strip comments
    auto comment = line.find("#");
    std::string part = line;
    for (; comment != std::string::npos; comment = part.find("#")) {
      if(comment)
        part = part.substr(0, comment);
      else break;
    }
    if(!part.size()) return false; // empty line after comments
    auto idx = part.find(":");
    if (idx == std::string::npos) return false; // bad format
    if (m_user != part.substr(0, idx)) return false; // username mismatch
    part = part.substr(idx+1);

    idx = part.find(":");
    if (idx == std::string::npos) return false; // bad format
    std::string cred = part.substr(0, idx);
    std::string salt = part.substr(idx+1);
    return Hash(m_passwd, salt) == cred;
  }

  std::string HashedCredDB::Hash(const std::string & data, const std::string & salt)
  {
    SHA512Digest h;
    std::string d = data + salt;
    SHA512((const uint8_t*)d.c_str(), d.size(), h);
    return B64Encode(h.data(), h.size());
  }
}

