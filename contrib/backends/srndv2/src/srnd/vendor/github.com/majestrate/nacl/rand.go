package nacl

// #cgo freebsd CFLAGS: -I/usr/local/include
// #cgo freebsd LDFLAGS: -L/usr/local/lib
// #cgo LDFLAGS: -lsodium
// #include <sodium.h>
import "C"

func randbytes(size C.size_t) *Buffer {

  buff := malloc(size)
  C.randombytes_buf(buff.ptr, size)
  return buff

}

func RandBytes(size int) []byte {
  if size > 0 {
    buff := randbytes(C.size_t(size))
    defer buff.Free()
    return buff.Bytes()
  }
  return nil
}
