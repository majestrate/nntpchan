/*
The MIT License (MIT)

Copyright (c) 2016 ZiRo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cuckoo

import "crypto/sha256"

const (
	SIZESHIFT uint64 = 20
	PROOFSIZE uint64 = 42
	SIZE      uint64 = 1 << SIZESHIFT
	HALFSIZE  uint64 = SIZE / 2
	NODEMASK  uint64 = HALFSIZE - 1
)

type Cuckoo struct {
	v [4]uint64
}

func u8(b byte) uint64 {
	return (uint64)(b) & 0xff
}

func u8to64(p [sha256.Size]byte, i int) uint64 {
	return u8(p[i]) | u8(p[i+1])<<8 |
		u8(p[i+2])<<16 | u8(p[i+3])<<24 |
		u8(p[i+4])<<32 | u8(p[i+5])<<40 |
		u8(p[i+6])<<48 | u8(p[i+7])<<56
}

func NewCuckoo(header []byte) *Cuckoo {
	hdrkey := sha256.Sum256(header)

	return NewCuckooSHA(hdrkey)
}

func NewCuckooSHA(hdrkey [sha256.Size]byte) *Cuckoo {
	self := new(Cuckoo)
	k0 := u8to64(hdrkey, 0)
	k1 := u8to64(hdrkey, 8)
	self.v[0] = k0 ^ 0x736f6d6570736575
	self.v[1] = k1 ^ 0x646f72616e646f6d
	self.v[2] = k0 ^ 0x6c7967656e657261
	self.v[3] = k1 ^ 0x7465646279746573

	return self
}

type Edge struct {
	U uint64
	V uint64
}

func (self *Edge) HashCode() int {
	return int(self.U) ^ int(self.V)
}

func (self *Cuckoo) Sipedge(nonce uint64) *Edge {
	return &Edge{self.Sipnode(nonce, 0), self.Sipnode(nonce, 1)}
}

func (self *Cuckoo) siphash24(nonce uint64) uint64 {
	v0 := self.v[0]
	v1 := self.v[1]
	v2 := self.v[2]
	v3 := self.v[3] ^ nonce

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | v1>>51
	v3 = (v3 << 16) | v3>>48
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | v0>>32
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | v1>>47
	v3 = (v3 << 21) | v3>>43
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | v2>>32

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | v1>>51
	v3 = (v3 << 16) | v3>>48
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | v0>>32
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | v1>>47
	v3 = (v3 << 21) | v3>>43
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | v2>>32

	v0 ^= nonce
	v2 ^= 0xff

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | v1>>51
	v3 = (v3 << 16) | v3>>48
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | v0>>32
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | v1>>47
	v3 = (v3 << 21) | v3>>43
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | v2>>32

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | v1>>51
	v3 = (v3 << 16) | v3>>48
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | v0>>32
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | v1>>47
	v3 = (v3 << 21) | v3>>43
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | v2>>32

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | v1>>51
	v3 = (v3 << 16) | v3>>48
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | v0>>32
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | v1>>47
	v3 = (v3 << 21) | v3>>43
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | v2>>32

	v0 += v1
	v2 += v3
	v1 = (v1 << 13) | v1>>51
	v3 = (v3 << 16) | v3>>48
	v1 ^= v0
	v3 ^= v2
	v0 = (v0 << 32) | v0>>32
	v2 += v1
	v0 += v3
	v1 = (v1 << 17) | v1>>47
	v3 = (v3 << 21) | v3>>43
	v1 ^= v2
	v3 ^= v0
	v2 = (v2 << 32) | v2>>32
	return v0 ^ v1 ^ v2 ^ v3
}

// generate edge in cuckoo graph
func (self *Cuckoo) Sipnode(nonce uint64, uorv uint32) uint64 {
	return self.siphash24(2*nonce+uint64(uorv)) & NODEMASK
}

// verify that (ascending) nonces, all less than easiness, form a cycle in graph
func (self *Cuckoo) Verify(nonces []uint64, easiness uint64) bool {
	us := make([]uint64, PROOFSIZE)
	vs := make([]uint64, PROOFSIZE)
	i := 0
	var n uint64

	for n = 0; n < PROOFSIZE; n++ {
		if nonces[n] >= easiness || (n != 0 && nonces[n] <= nonces[n-1]) {
			return false
		}
		us[n] = self.Sipnode(nonces[n], 0)
		vs[n] = self.Sipnode(nonces[n], 1)
	}

	loop := true
	for loop { // follow cycle until we return to i==0; n edges left to visit
		j := i
		for k := 0; uint64(k) < PROOFSIZE; k++ { // find unique other j with same vs[j]
			if k != i && vs[k] == vs[i] {
				if j != i {
					return false
				}
				j = k
			}
		}
		if j == i {
			return false
		}
		i = j
		for k := 0; uint64(k) < PROOFSIZE; k++ { // find unique other i with same us[i]
			if k != j && us[k] == us[j] {
				if i != j {
					return false
				}
				i = k
			}
		}
		if i == j {
			return false
		}
		n -= 2
		loop = (i != 0)
	}
	return n == 0
}
