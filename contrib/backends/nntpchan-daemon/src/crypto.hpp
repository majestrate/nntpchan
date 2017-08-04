#ifndef NNTPCHAN_CRYPTO_HPP
#define NNTPCHAN_CRYPTO_HPP

#include <sodium/crypto_hash.h>
#include <array>

namespace nntpchan
{
  typedef std::array<uint8_t, crypto_hash_BYTES> SHA512Digest;

  void SHA512(const uint8_t * d, std::size_t l, SHA512Digest & h);

  enum HashType
  {
    eHashTypeSHA1,
    eHashTypeSHA256,
    eHashTypeSHA512,
    eHashTypeBLAKE2B
  };

  template<HashType t>
  struct Hasher
  {
    typedef std::function<void(uint8_t *, size_t)> DigestVisitor;

    Hasher();
    ~Hasher();
    void Update(const char * data, size_t sz);
    void Final(DigestVisitor v);
  };

  /** global crypto initializer */
  struct Crypto
  {
    Crypto();
    ~Crypto();
  };
}


#endif
