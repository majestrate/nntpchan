#include <nntpchan/base64.hpp>
#include <nntpchan/crypto.hpp>

#include <cassert>
#include <cstring>
#include <iostream>
#include <sodium.h>
#include <string>

static void print_help(const std::string &exename)
{
  std::cout << "usage: " << exename << " [help|gen|check]" << std::endl;
}

static void gen_passwd(const std::string &username, const std::string &passwd)
{
  std::array<uint8_t, 8> random;
  randombytes_buf(random.data(), random.size());
  std::string salt = nntpchan::B64Encode(random.data(), random.size());
  std::string cred = passwd + salt;
  nntpchan::SHA512Digest d;
  nntpchan::SHA512((const uint8_t *)cred.c_str(), cred.size(), d);
  std::string hash = nntpchan::B64Encode(d.data(), d.size());
  std::cout << username << ":" << hash << ":" << salt << std::endl;
}

static bool check_cred(const std::string &cred, const std::string &passwd)
{
  auto idx = cred.find(":");
  if (idx == std::string::npos || idx == 0)
    return false;
  std::string part = cred.substr(idx + 1);
  idx = part.find(":");
  if (idx == std::string::npos || idx == 0)
    return false;
  std::string salt = part.substr(idx + 1);
  std::string hash = part.substr(0, idx);
  std::vector<uint8_t> h;
  if (!nntpchan::B64Decode(hash, h))
    return false;
  nntpchan::SHA512Digest d;
  std::string l = passwd + salt;
  nntpchan::SHA512((const uint8_t *)l.data(), l.size(), d);
  return std::memcmp(h.data(), d.data(), d.size()) == 0;
}

int main(int argc, char *argv[])
{
  assert(sodium_init() == 0);
  if (argc == 1)
  {
    print_help(argv[0]);
    return 0;
  }
  std::string cmd(argv[1]);
  if (cmd == "help")
  {
    print_help(argv[0]);
    return 0;
  }
  if (cmd == "gen")
  {
    if (argc == 4)
    {
      gen_passwd(argv[2], argv[3]);
      return 0;
    }
    else
    {
      std::cout << "usage: " << argv[0] << " gen username password" << std::endl;
      return 0;
    }
  }
  if (cmd == "check")
  {
    std::string cred;
    std::cout << "credential: ";
    if (!std::getline(std::cin, cred))
    {
      std::cout << "read error" << std::endl;
      return 1;
    }
    std::string passwd;
    std::cout << "password: ";
    if (!std::getline(std::cin, passwd))
    {
      std::cout << "read error" << std::endl;
      return 1;
    }
    if (check_cred(cred, passwd))
    {
      std::cout << "okay" << std::endl;
      return 0;
    }
    std::cout << "bad login" << std::endl;
    return 1;
  }
  print_help(argv[0]);
  return 1;
}
