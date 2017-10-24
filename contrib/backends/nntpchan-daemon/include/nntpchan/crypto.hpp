#ifndef NNTPCHAN_CRYPTO_HPP
#define NNTPCHAN_CRYPTO_HPP

#include <array>
#include <sodium/crypto_hash.h>
#include <sodium/crypto_generichash.h>

namespace nntpchan
{
typedef std::array<uint8_t, crypto_hash_BYTES> SHA512Digest;

void SHA512(const uint8_t *d, std::size_t l, SHA512Digest &h);

  typedef std::array<uint8_t, crypto_generichash_BYTES> Blake2BDigest;
  void Blake2B(const uint8_t *d, std::size_t l, Blake2BDigest & h);

  std::string Blake2B_base32(const std::string & str);

  
/** global crypto initializer */
struct Crypto
{
  Crypto();
  ~Crypto();
};
}

#endif
