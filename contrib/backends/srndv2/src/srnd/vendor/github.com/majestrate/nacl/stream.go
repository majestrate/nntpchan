package nacl

import (
  "bytes"
  "errors"
  "io"
  "net"
  "time"
)

// TOY encrypted authenticated stream protocol like tls 


var BadHandshake = errors.New("Bad handshake")
var ShortWrite = errors.New("short write")
var ShortRead = errors.New("short read")
var Closed = errors.New("socket closed")

// write boxes at 512 bytes at a time
const DefaultMTU = 512

// wrapper arround crypto_box
// provides an authenticated encrypted stream
// this is a TOY
type CryptoStream struct {
  // underlying stream to write on
  stream io.ReadWriteCloser
  // secret key seed
  key *KeyPair
  // public key of who we expect on the other end
  remote_pk []byte
  tx_nonce []byte
  rx_nonce []byte
  // box size
  mtu int
}

func (cs *CryptoStream) Close() (err error) {
  if cs.key != nil {
    cs.key.Free()
    cs.key = nil
  }
  return cs.stream.Close()
}

// implements io.Writer
func (cs *CryptoStream) Write(data []byte) (n int, err error) {
  // let's split it up
  for n < len(data) && err == nil {
    if n + cs.mtu < len(data) {
      err = cs.writeSegment(data[n:n+cs.mtu])
      n += cs.mtu
    } else {
      err = cs.writeSegment(data[n:])
      if err == nil {
        n = len(data)
      }
    }
  }
  return
}

func (cs *CryptoStream) public() (p []byte) {
  p = cs.key.Public()
  return
}

func (cs *CryptoStream) secret() (s []byte) {
  s = cs.key.Secret()
  return
}

// read 1 segment
func (cs *CryptoStream) readSegment() (s []byte, err error) {
  var stream_read int
  var seg []byte
  nl := NounceLen()
  msg := make([]byte, cs.mtu + nl)
  stream_read, err = cs.stream.Read(msg)
  seg, err = CryptoBoxOpen(msg[:stream_read], cs.rx_nonce, cs.secret(), cs.remote_pk)
  if err == nil {
    copy(cs.rx_nonce, seg[:nl])
    s = seg[nl:]
  }
  return
}

// write 1 segment encrypted
// update nounce
func (cs *CryptoStream) writeSegment(data []byte) (err error) {
  var segment []byte
  nl := NounceLen()
  msg := make([]byte, len(data) + nl)
  // generate next nounce
  nextNounce := NewBoxNounce()
  copy(msg, nextNounce)
  copy(msg[nl:], data)
  // encrypt segment with current nounce
  segment, err = CryptoBox(data, cs.tx_nonce, cs.remote_pk, cs.secret())
  var n int
  n, err = cs.stream.Write(segment)
  if n != len(segment) {
    // short write?
    err = ShortWrite
    return
  }
  // update nounce
  copy(cs.tx_nonce, nextNounce)
  return
}

// implements io.Reader
func (cs *CryptoStream) Read(data []byte) (n int, err error) {
  var seg []byte
  seg, err = cs.readSegment()
  if err == nil {
    if len(seg) <= len(data) {
      copy(data, seg)
      n = len(seg)
    } else {
      // too big?
      err = ShortRead
    }
  }
  return 
}

// version 0 protocol magic
var protocol_magic = []byte("BENIS|00")

// verify that a handshake is signed right and is in the correct format etc
func verifyHandshake(hs, pk []byte) (valid bool) {
  ml := len(protocol_magic)
  // valid handshake?
  if bytes.Equal(hs[0:ml], protocol_magic) {
    // check pk
    pl := CryptoSignPublicLen()
    nl := NounceLen()
    if bytes.Equal(pk, hs[ml:ml+pl]) {
      // check signature
      msg := hs[0:ml+pl+nl]
      sig := hs[ml+pl+nl:]
      valid = CryptoVerifyFucky(msg, sig, pk)
    }
  }
  return
}

// get claimed public key from handshake
func getPubkey(hs []byte) (pk []byte) {
  ml := len(protocol_magic)
  pl := CryptoSignPublicLen()
  pk = hs[ml:ml+pl]
  return
}

func (cs *CryptoStream) genHandshake() (d []byte) {
  // protocol magic string version 00
  // Benis Encrypted Network Information Stream
  // :-DDDDD meme crypto
  d = append(d, protocol_magic...)
  // our public key
  d = append(d, cs.public()...)
  // nounce
  cs.tx_nonce = NewBoxNounce()
  d = append(d, cs.tx_nonce...)
  // sign protocol magic string, nounce and pubkey
  sig := CryptoSignFucky(d, cs.secret())
  // if sig is nil we'll just die
  d = append(d, sig...)
  return
}

// extract nounce from handshake
func getNounce(hs []byte) (n []byte) {
  ml := len(protocol_magic)
  pl := CryptoSignPublicLen()
  nl := NounceLen()
  n = hs[ml+pl:ml+pl+nl]
  return
}

// initiate protocol handshake
func (cs *CryptoStream) Handshake() (err error) {
  // send them our info
  hs := cs.genHandshake()
  var n int
  n, err = cs.stream.Write(hs)
  if n != len(hs) {
    err = ShortWrite
    return
  }
  // read thier info
  buff := make([]byte, len(hs))
  _, err = io.ReadFull(cs.stream, buff)

  if cs.remote_pk == nil {
    // inbound
    pk := getPubkey(buff)
    cs.remote_pk = make([]byte, len(pk))
    copy(cs.remote_pk, pk)
  }
  
  if ! verifyHandshake(buff, cs.remote_pk) {
    // verification failed
    err = BadHandshake
    return
  }
  cs.rx_nonce = make([]byte, NounceLen())
  copy(cs.rx_nonce, getNounce(buff))
  return
}


// create a client 
func Client(stream io.ReadWriteCloser, local_sk, remote_pk []byte) (c *CryptoStream) {
  c = &CryptoStream{
    stream: stream,
    mtu: DefaultMTU,
  }
  c.remote_pk = make([]byte, len(remote_pk))
  copy(c.remote_pk, remote_pk)
  c.key = LoadSignKey(local_sk)
  if c.key == nil {
    return nil
  }
  return c
}


type CryptoConn struct {
  stream *CryptoStream
  conn net.Conn
}

func (cc *CryptoConn) Close() (err error) {
  err = cc.stream.Close()
  return
}

func (cc *CryptoConn) Write(d []byte) (n int, err error) {
  return cc.stream.Write(d)
}

func (cc *CryptoConn) Read(d []byte) (n int, err error) {
  return cc.stream.Read(d)
}

func (cc *CryptoConn) LocalAddr() net.Addr {
  return cc.conn.LocalAddr()
}

func (cc *CryptoConn) RemoteAddr() net.Addr {
  return cc.conn.RemoteAddr()
}

func (cc *CryptoConn) SetDeadline(t time.Time) (err error) {
  return cc.conn.SetDeadline(t)
}

func (cc *CryptoConn) SetReadDeadline(t time.Time) (err error) {
  return cc.conn.SetReadDeadline(t)
}

func (cc *CryptoConn) SetWriteDeadline(t time.Time) (err error) {
  return cc.conn.SetWriteDeadline(t)
}

type CryptoListener struct {
  l net.Listener
  handshake chan net.Conn
  accepted chan *CryptoConn
  trust func(pk []byte) bool
  key *KeyPair
}

func (cl *CryptoListener) Close() (err error) {
  err = cl.l.Close()
  close(cl.accepted)
  close(cl.handshake)
  cl.key.Free()
  cl.key = nil
  return
}

func (cl *CryptoListener) acceptInbound() {
  for {
    c, err := cl.l.Accept()
    if err == nil {
      cl.handshake <- c
    } else {
      return
    }
  }
}

func (cl *CryptoListener) runChans() {
  for {
    select {
    case c := <- cl.handshake:
      go func(){
        s := &CryptoStream{
          stream: c,
          mtu: DefaultMTU,
          key: cl.key,
        }
        err := s.Handshake()
        if err == nil {
          // we gud handshake was okay
          if cl.trust(s.remote_pk) {
            // the key is trusted okay
            cl.accepted <- &CryptoConn{stream: s, conn: c}
          } else {
            // not trusted, close connection
            s.Close()
          }
        }
      }()
    }
  }
}

// accept inbound authenticated and trusted connections
func (cl *CryptoListener) Accept() (c net.Conn, err error) {
  var ok bool
  c, ok = <- cl.accepted
  if ! ok {
    err = Closed
  }
  return
}

// create a listener
func Server(l net.Listener, local_sk []byte, trust func(pk []byte) bool) (s *CryptoListener) {
  s = &CryptoListener{
    l: l,
    trust: trust,
    handshake: make(chan net.Conn),
    accepted: make(chan *CryptoConn),
  }
  s.key = LoadSignKey(local_sk)
  go s.runChans()
  go s.acceptInbound()
  return
}
