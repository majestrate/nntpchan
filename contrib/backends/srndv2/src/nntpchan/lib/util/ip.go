package util

import (
	"encoding/base64"
	"fmt"
	"nntpchan/lib/crypto/nacl"
	"log"
	"net"
)

// given an address
// generate a new encryption key for it
// return the encryption key and the encrypted address
func NewAddrEnc(addr string) (string, string) {
	key_bytes := nacl.RandBytes(encAddrBytes())
	key := base64.StdEncoding.EncodeToString(key_bytes)
	return key, EncAddr(addr, key)
}

// xor address with a one time pad
// if the address isn't long enough it's padded with spaces
func EncAddr(addr, key string) string {
	key_bytes, err := base64.StdEncoding.DecodeString(key)

	if err != nil {
		log.Println("encAddr() key base64 decode", err)
		return ""
	}

	if len(addr) > len(key_bytes) {
		log.Println("encAddr() len(addr) > len(key_bytes)")
		return ""
	}

	// pad with spaces
	for len(addr) < len(key_bytes) {
		addr += " "
	}

	addr_bytes := []byte(addr)
	res_bytes := make([]byte, len(addr_bytes))
	for idx, b := range key_bytes {
		res_bytes[idx] = addr_bytes[idx] ^ b
	}

	return base64.StdEncoding.EncodeToString(res_bytes)
}

// number of bytes to use in otp
func encAddrBytes() int {
	return 64
}

func IsSubnet(cidr string) (bool, *net.IPNet) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err == nil {
		return true, ipnet
	}
	return false, nil
}

func IPNet2MinMax(inet *net.IPNet) (min, max net.IP) {
	netb := []byte(inet.IP)
	maskb := []byte(inet.Mask)
	maxb := make([]byte, len(netb))

	for i, _ := range maxb {
		maxb[i] = netb[i] | (^maskb[i])
	}
	min = net.IP(netb)
	max = net.IP(maxb)
	return
}

func ZeroIPString(ip net.IP) string {
	p := ip

	if len(ip) == 0 {
		return "<nil>"
	}

	if p4 := p.To4(); len(p4) == net.IPv4len {
		return fmt.Sprintf("%03d.%03d.%03d.%03d", p4[0], p4[1], p4[2], p4[3])
	}
	if len(p) == net.IPv6len {
		//>IPv6
		//ishygddt
		return fmt.Sprintf("[%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x]", p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7], p[8], p[9], p[10], p[11], p[12], p[13], p[14], p[15])
	}
	return "?"
}
