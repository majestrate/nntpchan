package nacl

// #cgo freebsd CFLAGS: -I/usr/local/include
// #cgo freebsd LDFLAGS: -L/usr/local/lib
// #cgo LDFLAGS: -lsodium
// #include <sodium.h>
import "C"


// verify a fucky detached sig
func CryptoVerifyFucky(msg, sig, pk []byte) bool {
  var smsg []byte
  smsg = append(smsg, sig...)
  smsg = append(smsg, msg...)
  return CryptoVerify(smsg, pk)
}

// verify a signed message
func CryptoVerify(smsg, pk []byte) bool {
  smsg_buff := NewBuffer(smsg)
  defer smsg_buff.Free()
  pk_buff := NewBuffer(pk)
  defer pk_buff.Free()

  if pk_buff.size != C.crypto_sign_publickeybytes() {
    return false
  }
  mlen := C.ulonglong(0)
  msg := malloc(C.size_t(len(smsg)))
  defer msg.Free()
  smlen := C.ulonglong(smsg_buff.size)
  return C.crypto_sign_open(msg.uchar(), &mlen, smsg_buff.uchar(), smlen, pk_buff.uchar()) != -1
}

// verfiy a detached signature
// return true on valid otherwise false
func CryptoVerifyDetached(msg, sig, pk []byte) bool {
  msg_buff := NewBuffer(msg)
  defer msg_buff.Free()
  sig_buff := NewBuffer(sig)
  defer sig_buff.Free()
  pk_buff := NewBuffer(pk)
  defer pk_buff.Free()

  if pk_buff.size != C.crypto_sign_publickeybytes() {
    return false
  }
  
  // invalid sig size
  if sig_buff.size != C.crypto_sign_bytes() {
    return false
  }
  return C.crypto_sign_verify_detached(sig_buff.uchar(), msg_buff.uchar(), C.ulonglong(len(msg)), pk_buff.uchar()) == 0
}
