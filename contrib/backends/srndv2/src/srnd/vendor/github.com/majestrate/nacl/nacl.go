package nacl

// #cgo freebsd CFLAGS: -I/usr/local/include
// #cgo freebsd LDFLAGS: -L/usr/local/lib
// #cgo LDFLAGS: -lsodium
// #include <sodium.h>
import "C"

import (
  "log"
)

// return how many bytes overhead does CryptoBox have
func CryptoBoxOverhead() int {
  return int(C.crypto_box_macbytes())
}

// size of crypto_box public keys
func CryptoBoxPubKeySize() int {
  return int(C.crypto_box_publickeybytes())
}

// size of crypto_box private keys
func CryptoBoxPrivKeySize() int {
  return int(C.crypto_box_secretkeybytes())
}

// size of crypto_sign public keys
func CryptoSignPubKeySize() int {
  return int(C.crypto_sign_publickeybytes())
}

// size of crypto_sign private keys
func CryptoSignPrivKeySize() int {
  return int(C.crypto_sign_secretkeybytes())
}


// initialize sodium
func init() {
  status := C.sodium_init()
  if status == -1 {
    log.Fatalf("failed to initialize libsodium status=%d", status)
  }
}
