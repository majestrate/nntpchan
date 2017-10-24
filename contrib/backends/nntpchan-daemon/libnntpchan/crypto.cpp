#include <cassert>
#include <nntpchan/base64.hpp>
#include <nntpchan/crypto.hpp>
#include <sodium.h>

namespace nntpchan
{
void SHA512(const uint8_t *d, const std::size_t l, SHA512Digest &h) { crypto_hash(h.data(), d, l); }

  void Blake2B(const uint8_t *d, std::size_t l, Blake2BDigest & h) { crypto_generichash(h.data(), h.size(), d, l, nullptr, 0); }

  std::string Blake2B_base32(const std::string & str)
  {
    Blake2BDigest d;
    Blake2B(reinterpret_cast<const uint8_t*>(str.c_str()), str.size(), d);
    return B32Encode(d.data(), d.size());
  }

  
Crypto::Crypto() { assert(sodium_init() == 0); }

Crypto::~Crypto() {}
}
