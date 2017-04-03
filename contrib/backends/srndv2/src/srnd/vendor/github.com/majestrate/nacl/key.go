package nacl

// #cgo freebsd CFLAGS: -I/usr/local/include
// #cgo freebsd LDFLAGS: -L/usr/local/lib
// #cgo LDFLAGS: -lsodium
// #include <sodium.h>
import "C"

import (
  "encoding/hex"
  "errors"
  "fmt"
)

type KeyPair struct {
  pk *Buffer
  sk *Buffer
}

// free this keypair from memory
func (self *KeyPair) Free() {
  self.pk.Free()
  self.sk.Free()
}

func (self *KeyPair) Secret() []byte {
  return self.sk.Bytes()
}

func (self *KeyPair) Public() []byte {
  return self.pk.Bytes()
}

func (self *KeyPair) Seed() []byte {
  seed_len := C.crypto_sign_seedbytes()
  return self.sk.Bytes()[:seed_len]
}

// generate a keypair
func GenSignKeypair() *KeyPair {
  sk_len := C.crypto_sign_secretkeybytes()
  sk := malloc(sk_len)
  pk_len := C.crypto_sign_publickeybytes()
  pk := malloc(pk_len)
  res := C.crypto_sign_keypair(pk.uchar(), sk.uchar())
  if res == 0 {
    return &KeyPair{pk,sk}
  }
  pk.Free()
  sk.Free()
  return nil
}

// get public key from secret key
func GetSignPubkey(sk []byte) ([]byte, error) {
  sk_len := C.crypto_sign_secretkeybytes()
  if C.size_t(len(sk)) != sk_len {
    return nil, errors.New(fmt.Sprintf("nacl.GetSignPubkey() invalid secret key size %d != %d", len(sk), sk_len))
  }
  
  pk_len := C.crypto_sign_publickeybytes()
  pkbuff := malloc(pk_len)
  defer pkbuff.Free()

  skbuff := NewBuffer(sk)
  defer skbuff.Free()
  //XXX: hack
  res := C.crypto_sign_seed_keypair(pkbuff.uchar(), skbuff.uchar(), skbuff.uchar())
  
  if res != 0 {
    return nil, errors.New(fmt.Sprintf("nacl.GetSignPubkey() failed to get public key from secret key: %d", res))
  }
  
  return pkbuff.Bytes(), nil
}

// make keypair from seed
func LoadSignKey(seed []byte) *KeyPair {
  seed_len := C.crypto_sign_seedbytes()
  if C.size_t(len(seed)) != seed_len {
    return nil
  }
  seedbuff := NewBuffer(seed)
  defer seedbuff.Free()
  pk_len := C.crypto_sign_publickeybytes()
  sk_len := C.crypto_sign_secretkeybytes()
  pkbuff := malloc(pk_len)
  skbuff := malloc(sk_len)
  res := C.crypto_sign_seed_keypair(pkbuff.uchar(), skbuff.uchar(), seedbuff.uchar())
  if res != 0 {
    pkbuff.Free()
    skbuff.Free()
    return nil
  }
  return &KeyPair{pkbuff, skbuff}
}

func GenBoxKeypair() *KeyPair {
  sk_len := C.crypto_box_secretkeybytes()
  sk := malloc(sk_len)
  pk_len := C.crypto_box_publickeybytes()
  pk := malloc(pk_len)
  res := C.crypto_box_keypair(pk.uchar(), sk.uchar())
  if res == 0 {
    return &KeyPair{pk,sk}
  }
  pk.Free()
  sk.Free()
  return nil  
}


// get public key from secret key
func GetBoxPubkey(sk []byte) []byte {
  sk_len := C.crypto_box_seedbytes()
  if C.size_t(len(sk)) != sk_len {
    return nil
  }
  
  pk_len := C.crypto_box_publickeybytes()
  pkbuff := malloc(pk_len)
  defer pkbuff.Free()

  skbuff := NewBuffer(sk)
  defer skbuff.Free()

  // compute the public key
  C.crypto_scalarmult_base(pkbuff.uchar(), skbuff.uchar())
  
  return pkbuff.Bytes()
}

// load keypair from secret key
func LoadBoxKey(sk []byte) *KeyPair {
  pk := GetBoxPubkey(sk)
  if pk == nil {
    return nil
  }
  pkbuff := NewBuffer(pk)
  skbuff := NewBuffer(sk)
  return &KeyPair{pkbuff, skbuff}
}

// make keypair from seed
func SeedBoxKey(seed []byte) *KeyPair {
  seed_len := C.crypto_box_seedbytes()
  if C.size_t(len(seed)) != seed_len {
    return nil
  }
  seedbuff := NewBuffer(seed)
  defer seedbuff.Free()
  pk_len := C.crypto_box_publickeybytes()
  sk_len := C.crypto_box_secretkeybytes()
  pkbuff := malloc(pk_len)
  skbuff := malloc(sk_len)
  res := C.crypto_box_seed_keypair(pkbuff.uchar(), skbuff.uchar(), seedbuff.uchar())
  if res != 0 {
    pkbuff.Free()
    skbuff.Free()
    return nil
  }
  return &KeyPair{pkbuff, skbuff}
}

func (self *KeyPair) String() string {
  return fmt.Sprintf("pk=%s sk=%s", hex.EncodeToString(self.pk.Data()), hex.EncodeToString(self.sk.Data()))
}

func CryptoSignPublicLen() int {
  return int(C.crypto_sign_publickeybytes())
}


func CryptoSignSecretLen() int {
  return int(C.crypto_sign_secretkeybytes())
}

func CryptoSignSeedLen() int {
  return int(C.crypto_sign_seedbytes())
}
