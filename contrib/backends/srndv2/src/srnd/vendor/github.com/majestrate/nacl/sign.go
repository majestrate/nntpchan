package nacl

// #cgo freebsd CFLAGS: -I/usr/local/include
// #cgo freebsd LDFLAGS: -L/usr/local/lib
// #cgo LDFLAGS: -lsodium
// #include <sodium.h>
import "C"


// sign data detached with secret key sk 
func CryptoSignDetached(msg, sk []byte) []byte {
  msgbuff := NewBuffer(msg)
  defer msgbuff.Free()
  skbuff := NewBuffer(sk)
  defer skbuff.Free()
  if skbuff.size != C.crypto_sign_bytes() {
    return nil
  }
  
  // allocate the signature buffer
  sig := malloc(C.crypto_sign_bytes())
  defer sig.Free()
  // compute signature
  siglen := C.ulonglong(0)
  res := C.crypto_sign_detached(sig.uchar(), &siglen, msgbuff.uchar(), C.ulonglong(msgbuff.size), skbuff.uchar())
  if res == 0 && siglen == C.ulonglong(C.crypto_sign_bytes()) {
    // return copy of signature buffer
    return sig.Bytes()
  }
  // failure to sign
  return nil
}


// sign data with secret key sk
// return detached sig
// this uses crypto_sign instead pf crypto_sign_detached
func CryptoSignFucky(msg, sk []byte) []byte {
  msgbuff := NewBuffer(msg)
  defer msgbuff.Free()
  skbuff := NewBuffer(sk)
  defer skbuff.Free()
  if skbuff.size != C.crypto_sign_bytes() {
    return nil
  }
  
  // allocate the signed message buffer
  sig := malloc(C.crypto_sign_bytes()+msgbuff.size)
  defer sig.Free()
  // compute signature
  siglen := C.ulonglong(0)
  res := C.crypto_sign(sig.uchar(), &siglen, msgbuff.uchar(), C.ulonglong(msgbuff.size), skbuff.uchar())
  if res == 0 {
    // return copy of signature inside the signed message
    offset := int(C.crypto_sign_bytes())
    return sig.Bytes()[:offset]
  }
  // failure to sign
  return nil
}
