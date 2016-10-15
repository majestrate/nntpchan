#include "crypto.hpp"

namespace nntpchan
{
  void SHA512(const uint8_t * d, const std::size_t l, SHA512Digest & h)
  {
    crypto_hash(h.data(), d, l);
  }
}
