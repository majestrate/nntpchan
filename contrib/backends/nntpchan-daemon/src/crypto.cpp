#include "crypto.hpp"
#include <sodium.h>
#include <cassert>

namespace nntpchan
{
  void SHA512(const uint8_t * d, const std::size_t l, SHA512Digest & h)
  {
    crypto_hash(h.data(), d, l);
  }

  Crypto::Crypto()
  {
    assert(sodium_init() == 0);
  }

  Crypto::~Crypto()
  {
  }
}
