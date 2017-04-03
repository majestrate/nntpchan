package nacl

// #cgo freebsd CFLAGS: -I/usr/local/include
// #cgo freebsd LDFLAGS: -L/usr/local/lib
// #cgo LDFLAGS: -lsodium
// #include <sodium.h>
import "C"

import (
  "errors"
)

// encrypts a message to a user given their public key is known
// returns an encrypted box
func CryptoBox(msg, nounce, pk, sk []byte) ([]byte, error) {
  msgbuff := NewBuffer(msg)
  defer msgbuff.Free()

  // check sizes
  if len(pk) != int(C.crypto_box_publickeybytes()) {
    err := errors.New("len(pk) != crypto_box_publickey_bytes")
    return nil, err
  }
  if len(sk) != int(C.crypto_box_secretkeybytes()) {
    err := errors.New("len(sk) != crypto_box_secretkey_bytes")
    return nil, err
  }
  if len(nounce) != int(C.crypto_box_macbytes()) {
    err := errors.New ("len(nounce) != crypto_box_macbytes()")
    return nil, err
  }
  
  pkbuff := NewBuffer(pk)
  defer pkbuff.Free()
  skbuff := NewBuffer(sk)
  defer skbuff.Free()
  nouncebuff := NewBuffer(nounce)
  defer nouncebuff.Free()
  
  resultbuff := malloc(msgbuff.size + nouncebuff.size)
  defer resultbuff.Free()
  res := C.crypto_box_easy(resultbuff.uchar(), msgbuff.uchar(), C.ulonglong(msgbuff.size), nouncebuff.uchar(), pkbuff.uchar(), skbuff.uchar())
  if res != 0 {
    err := errors.New("crypto_box_easy failed")
    return nil, err
  }
  return resultbuff.Bytes(), nil
}

// open an encrypted box
func CryptoBoxOpen(box, nounce, sk, pk []byte) ([]byte, error) {
  boxbuff := NewBuffer(box)
  defer boxbuff.Free()

  // check sizes
  if len(pk) != int(C.crypto_box_publickeybytes()) {
    err := errors.New("len(pk) != crypto_box_publickey_bytes")
    return nil, err
  }
  if len(sk) != int(C.crypto_box_secretkeybytes()) {
    err := errors.New("len(sk) != crypto_box_secretkey_bytes")
    return nil, err
  }
  if len(nounce) != int(C.crypto_box_macbytes()) {
    err := errors.New("len(nounce) != crypto_box_macbytes()")
    return nil, err
  }
    
  pkbuff := NewBuffer(pk)
  defer pkbuff.Free()
  skbuff := NewBuffer(sk)
  defer skbuff.Free()
  nouncebuff := NewBuffer(nounce)
  defer nouncebuff.Free()
  resultbuff := malloc(boxbuff.size - nouncebuff.size)
  defer resultbuff.Free()
  
  // decrypt
  res := C.crypto_box_open_easy(resultbuff.uchar(), boxbuff.uchar(), C.ulonglong(boxbuff.size), nouncebuff.uchar(), pkbuff.uchar(), skbuff.uchar())
  if res != 0 {
    return nil, errors.New("crypto_box_open_easy failed")
  }
  // return result
  return resultbuff.Bytes(), nil
}

// generate a new nounce
func NewBoxNounce() []byte {
  return RandBytes(NounceLen())
}

// length of a nounce
func NounceLen() int {
  return int(C.crypto_box_macbytes())
}
