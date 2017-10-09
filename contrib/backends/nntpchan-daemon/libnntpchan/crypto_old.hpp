

#ifndef NNTPCHAN_CRYPTO_OLD_HPP
#define NNTPCHAN_CRYPTO_OLD_HPP

#include <cstdint>
#include <cstdlib>
extern "C" {

typedef struct
{
    uint32_t state[5];
    uint32_t count[2];
    unsigned char buffer[64];
} SHA1_CTX;

void SHA1Transform(
  uint32_t state[5],
  const unsigned char buffer[64]
);
  
void SHA1Init(
  SHA1_CTX * context
);

void SHA1Update(
    SHA1_CTX * context,
    const unsigned char *data,
    uint32_t len
    );

void SHA1Final(
    unsigned char digest[20],
    SHA1_CTX * context
    );

void sha1(
    uint8_t *hash_out,
    const uint8_t *str,
    size_t len);

}

#endif
