//
// util.go -- various utilities
//

package srnd

import (
	"bufio"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/majestrate/nacl"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func DelFile(fname string) {
	if CheckFile(fname) {
		os.Remove(fname)
	}
}

func CheckFile(fname string) bool {
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		return false
	}
	return true
}

func IsDir(dirname string) bool {
	stat, err := os.Stat(dirname)
	if err != nil {
		log.Fatal(err)
	}
	return stat.IsDir()
}

// ensure a directory exists
func EnsureDir(dirname string) {
	stat, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		os.Mkdir(dirname, 0755)
	} else if !stat.IsDir() {
		os.Remove(dirname)
		os.Mkdir(dirname, 0755)
	}
}

var exp_valid_message_id = regexp.MustCompilePOSIX(`^<[a-zA-Z0-9$.]{2,128}@[a-zA-Z0-9\-.]{2,63}>$`)

func ValidMessageID(id string) bool {
	return exp_valid_message_id.MatchString(id)
}

// message id hash
func HashMessageID(msgid string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(msgid)))
}

// short message id hash
func ShortHashMessageID(msgid string) string {
	return strings.ToLower(HashMessageID(msgid)[:18])
}

// will this message id produce quads?
func MessageIDWillDoQuads(msgid string) bool {
	h := HashMessageID(msgid)
	return h[0] == h[1] && h[1] == h[2] && h[2] == h[3]
}

// will this message id produce trips?
func MessageIDWillDoTrips(msgid string) bool {
	h := HashMessageID(msgid)
	return h[0] == h[1] && h[1] == h[2]
}

// will this message id produce dubs?
func MessageIDWillDoDubs(msgid string) bool {
	h := HashMessageID(msgid)
	return h[0] == h[1]
}

// shorter message id hash
func ShorterHashMessageID(msgid string) string {
	return strings.ToLower(HashMessageID(msgid)[:10])
}

func OpenFileWriter(fname string) (io.WriteCloser, error) {
	return os.Create(fname)
}

// make a random string
func randStr(length int) string {
	return hex.EncodeToString(nacl.RandBytes(length))[length:]
}

// time for right now as int64
func timeNow() int64 {
	return time.Now().UTC().Unix()
}

// sanitize data for nntp
func nntpSanitize(data string) (ret string) {
	parts := strings.Split(data, "\n")
	lines := len(parts)
	for idx, part := range parts {
		part = strings.Replace(part, "\n", "", -1)
		part = strings.Replace(part, "\r", "", -1)
		if part == "." {
			part = "  ."
		}
		ret += part
		if idx+1 < lines {
			ret += "\n"
		}
	}
	return ret
}

type int64Sorter []int64

func (self int64Sorter) Len() int {
	return len(self)
}

func (self int64Sorter) Less(i, j int) bool {
	return self[i] < self[j]
}

func (self int64Sorter) Swap(i, j int) {
	tmp := self[j]
	self[j] = self[i]
	self[i] = tmp
}

// obtain the "real" ip address
func getRealIP(name string) string {
	if len(name) > 0 {
		ip, err := net.ResolveIPAddr("ip", name)
		if err == nil {
			if ip.IP.IsGlobalUnicast() {
				return ip.IP.String()
			}
		}
	}
	return ""
}

// check that we have permission to access this
// fatal on fail
func checkPerms(fname string) {
	fstat, err := os.Stat(fname)
	if err != nil {
		log.Fatalf("Cannot access %s, %s", fname, err)
	}
	// check if we can access this dir
	if fstat.IsDir() {
		tmpfname := filepath.Join(fname, ".test")
		f, err := os.Create(tmpfname)
		if err != nil {
			log.Fatalf("No Write access in %s, %s", fname, err)
		}
		err = f.Close()
		if err != nil {
			log.Fatalf("failed to close test file %s !? %s", tmpfname, err)
		}
		err = os.Remove(tmpfname)
		if err != nil {
			log.Fatalf("failed to remove test file %s, %s", tmpfname, err)
		}
	} else {
		// this isn't a dir, treat it like a regular file
		f, err := os.Open(fname)
		if err != nil {
			log.Fatalf("cannot read file %s, %s", fname, err)
		}
		f.Close()
	}
}

// number of bytes to use in otp
func encAddrBytes() int {
	return 64
}

// length of an encrypted clearnet address
func encAddrLen() int {
	return 88
}

// length of an i2p dest hash
func i2pDestHashLen() int {
	return 44
}

// given an address
// generate a new encryption key for it
// return the encryption key and the encrypted address
func newAddrEnc(addr string) (string, string) {
	key_bytes := nacl.RandBytes(encAddrBytes())
	key := base64.StdEncoding.EncodeToString(key_bytes)
	return key, encAddr(addr, key)
}

// xor address with a one time pad
// if the address isn't long enough it's padded with spaces
func encAddr(addr, key string) string {
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

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// decrypt an address
// strips any whitespaces
func decAddr(encaddr, key string) string {
	encaddr_bytes, err := base64.StdEncoding.DecodeString(encaddr)
	if err != nil {
		log.Println("decAddr() encaddr base64 decode", err)
		return ""
	}
	if len(encaddr) != len(key) {
		log.Println("decAddr() len(encaddr_bytes) != len(key)")
		return ""
	}
	key_bytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		log.Println("decAddr() key base64 decode", err)
	}
	res_bytes := make([]byte, len(key))
	for idx, b := range key_bytes {
		res_bytes[idx] = encaddr_bytes[idx] ^ b
	}
	res := string(res_bytes)
	return strings.Trim(res, " ")
}

var exp_valid_newsgroup = regexp.MustCompilePOSIX(`^[a-zA-Z0-9.]{1,128}$`)

func newsgroupValidFormat(newsgroup string) bool {
	return exp_valid_newsgroup.MatchString(newsgroup)
}

func ValidNewsgroup(newsgroup string) bool {
	return newsgroupValidFormat(newsgroup)
}

// generate a new signing keypair
// public, secret
func newSignKeypair() (string, string) {
	kp := nacl.GenSignKeypair()
	defer kp.Free()
	pk := kp.Public()
	sk := kp.Seed()
	return hex.EncodeToString(pk), hex.EncodeToString(sk)
}

// make a utf-8 tripcode
func makeTripcode(pk string) string {
	data, err := hex.DecodeString(pk)
	if err == nil {
		tripcode := ""
		//  here is the python code this is based off of
		//  i do something slightly different but this is the base
		//
		//  for x in range(0, length / 2):
		//    pub_short += '&#%i;' % (9600 + int(full_pubkey_hex[x*2:x*2+2], 16))
		//  length -= length / 2
		//  for x in range(0, length):
		//    pub_short += '&#%i;' % (9600 + int(full_pubkey_hex[-(length*2):][x*2:x*2+2], 16))
		//
		for _, c := range data {
			ch := 9600
			ch += int(c)
			tripcode += fmt.Sprintf("&#%04d;", ch)
		}
		return tripcode
	}
	return "[invalid]"
}

// generate a new message id with base name
func genMessageID(name string) string {
	return fmt.Sprintf("<%s%d@%s>", randStr(5), timeNow(), name)
}

// time now as a string timestamp
func timeNowStr() string {
	return time.Unix(timeNow(), 0).UTC().Format(time.RFC1123Z)
}

func queryGetInt64(q url.Values, key string, fallback int64) int64 {
	val := q.Get(key)
	if val != "" {
		i, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			return i
		}
	}
	return fallback
}

// get from a map an int given a key or fall back to a default value
func mapGetInt(m map[string]string, key string, fallback int) int {
	val, ok := m[key]
	if ok {
		i, err := strconv.ParseInt(val, 10, 32)
		if err == nil {
			return int(i)
		}
	}
	return fallback
}

func isSage(str string) bool {
	str = strings.ToLower(str)
	return str == "sage" || strings.HasPrefix(str, "sage ")
}

func unhex(str string) []byte {
	buff, _ := hex.DecodeString(str)
	return buff
}

func hexify(data []byte) string {
	return hex.EncodeToString(data)
}

// extract pubkey from secret key
// return as hex
func getSignPubkey(sk []byte) string {
	k, _ := nacl.GetSignPubkey(sk)
	return hexify(k)
}

// sign data with secret key the fucky srnd way
// return signature as hex
func cryptoSign(h, sk []byte) string {
	// sign
	sig := nacl.CryptoSignFucky(h, sk)
	if sig == nil {
		return "[failed to sign]"
	}
	return hexify(sig)
}

// given a tripcode after the #
// make a seed byteslice
func parseTripcodeSecret(str string) []byte {
	// try decoding hex
	raw := unhex(str)
	keylen := nacl.CryptoSignSeedLen()
	if raw == nil || len(raw) != keylen {
		// treat this as a "regular" chan tripcode
		// decode as bytes then pad the rest with 0s if it doesn't fit
		raw = make([]byte, keylen)
		str_bytes := []byte(str)
		if len(str_bytes) > keylen {
			copy(raw, str_bytes[:keylen])
		} else {
			copy(raw, str_bytes)
		}
	}
	return raw
}

// generate a login salt for nntp users
func genLoginCredSalt() (salt string) {
	salt = randStr(128)
	return
}

// do nntp login credential hash given password and salt
func nntpLoginCredHash(passwd, salt string) (str string) {
	var b []byte
	b = append(b, []byte(passwd)...)
	b = append(b, []byte(salt)...)
	h := sha512.Sum512(b)
	str = base64.StdEncoding.EncodeToString(h[:])
	return
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

func getThreadHash(file string) (thread string) {
	exp := regexp.MustCompilePOSIX(`thread-([0-9a-f]+)\.*`)
	matches := exp.FindStringSubmatch(file)
	if len(matches) != 2 {
		return ""
	}
	thread = matches[1]
	return
}

func getUkkoPage(file string) (page int) {
	exp := regexp.MustCompilePOSIX(`ukko-([0-9]+)\.*`)
	matches := exp.FindStringSubmatch(file)
	if len(matches) != 2 {
		return
	}
	var err error
	page, err = strconv.Atoi(matches[1])
	if err != nil {
		page = 0
	}
	return
}

func getGroupAndPage(file string) (board string, page int) {
	exp := regexp.MustCompilePOSIX(`(.*)-([0-9]+)\.*`)
	matches := exp.FindStringSubmatch(file)
	if len(matches) != 3 {
		return "", -1
	}
	var err error
	board = matches[1]
	tmp := matches[2]
	page, err = strconv.Atoi(tmp)
	if err != nil {
		page = -1
	}
	return
}

func getGroupForCatalog(file string) (group string) {
	exp := regexp.MustCompilePOSIX(`catalog-(.+)\.html`)
	matches := exp.FindStringSubmatch(file)
	if len(matches) != 2 {
		return ""
	}
	group = matches[1]
	return
}

// get a message id from a mime header
// checks many values
func getMessageID(hdr textproto.MIMEHeader) (msgid string) {
	msgid = hdr.Get("Message-Id")
	if msgid == "" {
		msgid = hdr.Get("Message-ID")
	}
	if msgid == "" {
		msgid = hdr.Get("message-id")
	}
	if msgid == "" {
		msgid = hdr.Get("MESSAGE-ID")
	}
	return
}

func getMessageIDFromArticleHeaders(hdr ArticleHeaders) (msgid string) {
	msgid = hdr.Get("Message-Id", hdr.Get("Message-ID", hdr.Get("message-id", hdr.Get("MESSAGE-ID", ""))))
	return
}

func readMIMEHeader(r *bufio.Reader) (msg *mail.Message, err error) {
	msg, err = mail.ReadMessage(r)
	/*
		hdr = make(textproto.MIMEHeader)
		for {
			var str string
			str, err = r.ReadString(10)
			if err != nil {
				hdr = nil
				return
			}
			str = strings.Trim(str, "\r")
			str = strings.Trim(str, "\n")
			if str == "" {
				break
			}
			idx := strings.Index(str, ": ")
			if idx > 0 {
				hdrname := strings.Trim(str[:idx], " ")
				hdrval := strings.Trim(str[idx+2:], "\r\n")
				hdr.Add(hdrname, hdrval)
			} else {
				log.Println("invalid header", str)
			}
		}
	*/
	return
}

// write out a mime header to a writer
func writeMIMEHeader(wr io.Writer, hdr map[string][]string) (err error) {
	// write headers
	for k, vals := range hdr {
		for _, val := range vals {
			wr.Write([]byte(k))
			wr.Write([]byte(": "))
			wr.Write([]byte(val))
			_, err = wr.Write([]byte{10})
		}
	}
	// end of headers
	_, err = wr.Write([]byte{10})
	return
}

// like ioutil.Discard but an io.WriteCloser
type discardCloser struct {
}

func (*discardCloser) Write(data []byte) (n int, err error) {
	n = len(data)
	return
}

func (*discardCloser) Close() (err error) {
	return
}

// like ioutil.Discard but an io.WriteCloser
var Discard = new(discardCloser)

func extractParamFallback(param map[string]interface{}, k, fallback string) string {
	v, ok := param[k]
	if ok {
		return v.(string)
	}
	return fallback
}

func extractParam(param map[string]interface{}, k string) string {
	return extractParamFallback(param, k, "")
}

// get real ip addresss from an http request
func extractRealIP(r *http.Request) (ip string, err error) {
	ip, _, err = net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Println("extract real ip: ", err)
	}
	// TODO: have in config upstream proxy ip and check for that
	if strings.HasPrefix(ip, "127.") {
		// if it's loopback check headers for reverse proxy headers
		// TODO: make sure this isn't a tor user being sneaky
		ip = getRealIP(r.Header.Get("X-Real-IP"))
		if ip == "" {
			// try X-Forwarded-For if X-Real-IP not set
			_ip := r.Header.Get("X-Forwarded-For")
			parts := strings.Split(_ip, ",")
			_ip = parts[0]
			ip = getRealIP(_ip)
		}
		if ip == "" {
			ip = "127.0.0.1"
		}
	}
	return
}

func serverPubkeyIsValid(pubkey string) bool {
	b := unhex(pubkey)
	return b != nil && len(b) == nacl.CryptoSignPubKeySize()
}

func verifyFrontendSig(pubkey, sig, msgid string) bool {
	s := unhex(sig)
	k := unhex(pubkey)
	h := sha512.Sum512([]byte(msgid))
	return nacl.CryptoVerifyFucky(h[:], s, k)
}

func msgidFrontendSign(sk []byte, msgid string) string {
	h := sha512.Sum512([]byte(msgid))
	return cryptoSign(h[:], sk)
}
