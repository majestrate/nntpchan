//
// util.go -- various utilities
//

package srnd

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/ed25519"
	"io"
	"log"
	"mime"
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
	"unicode"
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

// printableASCII tells whether string is made of US-ASCII printable characters
// except of specified one.
func printableASCII(s string, e byte) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		// NOTE: doesn't include space, which is neither printable nor control
		if c <= 32 || c >= 127 || c == e {
			return false
		}
	}
	return true
}

func ValidMessageID(id string) bool {
	/*
	   {RFC 3977}
	   o  A message-id MUST begin with "<", end with ">", and MUST NOT
	      contain the latter except at the end.
	   o  A message-id MUST be between 3 and 250 octets in length.
	   o  A message-id MUST NOT contain octets other than printable US-ASCII
	      characters.

	   additionally, we check path characters, they may be dangerous
	*/
	return len(id) >= 3 && len(id) <= 250 &&
		id[0] == '<' && id[len(id)-1] == '>' &&
		printableASCII(id[1:len(id)-1], '>') &&
		strings.IndexAny(id[1:len(id)-1], "/\\") < 0
}

func ReservedMessageID(id string) bool {
	return id == "<0>" || id == "<keepalive@dummy.tld>"
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

func randbytes(l int) []byte {
	b := make([]byte, l)
	io.ReadFull(rand.Reader, b)
	return b
}

// make a random string
func randStr(length int) string {
	return hex.EncodeToString(randbytes(length))[length:]
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

var safeHeaderReplacer = strings.NewReplacer(
	"\t", " ",
	"\n", string(unicode.ReplacementChar),
	"\r", string(unicode.ReplacementChar),
	"\000", string(unicode.ReplacementChar))

// safeHeader replaces dangerous stuff from header,
// also replaces space with tab for XOVER/OVER output
func safeHeader(s string) string {
	return strings.TrimSpace(safeHeaderReplacer.Replace(s))
}

func isVchar(r rune) bool {
	// RFC 5234 B.1: VCHAR =  %x21-7E ; visible (printing) characters
	// RFC 6532 3.2: VCHAR =/ UTF8-non-ascii
	return (r >= 0x21 && r <= 0x7E) || r >= 0x80
}

func isAtext(r rune) bool {
	// RFC 5322: Printable US-ASCII characters not including specials.  Used for atoms.
	switch r {
	case '(', ')', '<', '>', '[', ']', ':', ';', '@', '\\', ',', '.', '"':
		return false
	}
	return isVchar(r)
}

func isWSP(r rune) bool { return r == ' ' || r == '\t' }

func isQtext(r rune) bool {
	if r == '\\' || r == '"' {
		return false
	}
	return isVchar(r)
}

func writeQuoted(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		if isQtext(r) || isWSP(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('\\')
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
}

func formatAddress(name, email string) string {
	// somewhat based on stdlib' mail.Address.String()

	b := &strings.Builder{}

	if name != "" {
		needsEncoding := false
		needsQuoting := false
		for i, r := range name {
			if r >= 0x80 || (!isWSP(r) && !isVchar(r)) {
				needsEncoding = true
				break
			}
			if isAtext(r) {
				continue
			}
			if r == ' ' && i > 0 && name[i-1] != ' ' && i < len(name)-1 {
				// allow spaces but only surrounded by non-spaces
				// otherwise they will be removed by receiver
				continue
			}
			needsQuoting = true
		}

		if needsEncoding {
			// Text in an encoded-word in a display-name must not contain certain
			// characters like quotes or parentheses (see RFC 2047 section 5.3).
			// When this is the case encode the name using base64 encoding.
			if strings.ContainsAny(name, "\"#$%&'(),.:;<>@[]^`{|}~") {
				b.WriteString(mime.BEncoding.Encode("utf-8", name))
			} else {
				b.WriteString(mime.QEncoding.Encode("utf-8", name))
			}
		} else if needsQuoting {
			writeQuoted(b, name)
		} else {
			b.WriteString(name)
		}

		b.WriteByte(' ')
	}

	at := strings.LastIndex(email, "@")
	var local, domain string
	if at >= 0 {
		local, domain = email[:at], email[at+1:]
	} else {
		local = email
	}

	quoteLocal := false
	for i, r := range local {
		if isAtext(r) {
			// if atom then okay
			continue
		}
		if r == '.' && r > 0 && local[i-1] != '.' && i < len(local)-1 {
			// dots are okay but only if surrounded by non-dots
			continue
		}
		quoteLocal = true
		break
	}

	b.WriteByte('<')
	if !quoteLocal {
		b.WriteString(local)
	} else {
		writeQuoted(b, local)
	}
	b.WriteByte('@')
	b.WriteString(domain)
	b.WriteByte('>')

	return b.String()
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
	key_bytes := randbytes(encAddrBytes())
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
	res = strings.TrimFunc(res, func(r rune) bool {
		if r <= 57 && r >= 48 {
			return false
		}
		if r == '.' || r == '/' {
			return false
		}
		return true
	})
	if strings.Index(res, "/") == -1 {
		// TODO: ipv6
		res += "/32"
	}
	return res
}

var exp_valid_newsgroup = regexp.MustCompilePOSIX(`^[a-zA-Z0-9.]{1,128}$`)

func newsgroupValidFormat(newsgroup string) bool {
	newsgroup = strings.TrimFunc(newsgroup, func(r rune) bool {
		return r == ' '
	})
	return exp_valid_newsgroup.MatchString(newsgroup) && len(newsgroup) > 0
}

func ValidNewsgroup(newsgroup string) bool {
	return newsgroupValidFormat(newsgroup)
}

func genNaclKeypair() (pk, sk []byte) {
	sk = randbytes(32)
	pk, _ = naclSeedToKeyPair(sk)
	return
}

// generate a new signing keypair
// public, secret
func newNaclSignKeypair() (string, string) {
	pk, sk := genNaclKeypair()
	return hex.EncodeToString(pk), hex.EncodeToString(sk)
}

func makeTripcodeLen(pubkey string, length int) string {
	var b strings.Builder

	data, err := hex.DecodeString(pubkey)
	if err != nil {
		return "[invalid]"
	}

	if length <= 0 || length > len(data) {
		length = len(data)
	}

	// originally srnd (and srndv2) used 9600==0x2580
	// however, range shifted by 0x10 looks better to me (cathugger)
	// (instead of `▀▁▂▃▄▅▆▇█▉▊▋▌▍▎▏` it'll use `⚀⚁⚂⚃⚄⚅⚆⚇⚈⚉⚊⚋⚌⚍⚎⚏`)
	// and display equaly good both in torbrowser+DejaVuSans and phone
	// since jeff ack'd it (he doesn't care probably), I'll just use it
	const rstart = 0x2590
	// 0x2500 can display with TBB font whitelist, but looks too cryptic.
	// startin from 0x2600 needs more than DejaVuSans so I'll avoid it

	// logic (same as in srnd):
	// it first writes length/2 chars of begining
	// and then length/2 chars of ending
	// if length==len(data), that essentially means just using whole
	i := 0
	for ; i < length/2; i++ {
		b.WriteRune(rstart + rune(data[i]))
		b.WriteRune(0xFE0E) // text style variant
	}
	for ; i < length; i++ {
		b.WriteRune(rstart + rune(data[len(data)-length+i]))
		b.WriteRune(0xFE0E) // text style variant
	}

	return b.String()
}

// make a utf-8 tripcode
func makeTripcode(pk string) string {
	return makeTripcodeLen(pk, 0)
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

// get from a map an uint64 given a key or fall back to a default value
func mapGetInt64(m map[string]string, key string, fallback int64) int64 {
	val, ok := m[key]
	if ok {
		i, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			return i
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
	pk, _ := naclSeedToKeyPair(sk)
	return hexify(pk)
}

// sign data with secret key the fucky srnd way
// return signature as hex
// XXX: DEPRECATED
func cryptoSignFucky(h, sk []byte) string {
	// sign
	sig := naclCryptoSignFucky(h, sk)
	if sig == nil {
		return "[failed to sign]"
	}
	return hexify(sig)
}

func cryptoSignProper(h, sk []byte) string {
	key := make(ed25519.PrivateKey, ed25519.PrivateKeySize)
	copy(key, sk)
	// sign
	sig := ed25519.Sign(key, h)
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
	keylen := 32
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

	for i := range maxb {
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
func getMessageID(h map[string][]string) string {
	for k := range h {
		kl := strings.ToLower(k)
		if kl == "message-id" || kl == "messageid" {
			return strings.TrimSpace(h[k][0])
		}
	}
	return ""
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
	return b != nil && len(b) == 32
}

func verifyFrontendSig(pubkey, sig, msgid string) bool {
	s := unhex(sig)
	k := unhex(pubkey)
	h := sha512.Sum512([]byte(msgid))
	return naclCryptoVerifyFucky(h[:], s, k)
}

func msgidFrontendSign(sk []byte, msgid string) string {
	h := sha512.Sum512([]byte(msgid))
	return cryptoSignFucky(h[:], sk)
}

func patMatch(v, pat string) (found bool) {
	parts := strings.Split(pat, ",")
	for _, part := range parts {
		var invert bool
		if part[0] == '!' {
			invert = true
			if len(parts) == 0 {
				return
			}
			part = part[1:]
		}
		found, _ = regexp.MatchString(v, part)
		log.Println(v, part, found)
		if invert {
			found = !found
		}
		if found {
			return
		}
	}
	return
}

func headerFindPats(header string, hdr ArticleHeaders, patterns []string) (found ArticleHeaders) {
	found = make(ArticleHeaders)
	if hdr.Has(header) && len(patterns) > 0 {
		for _, v := range hdr[header] {
			for _, pat := range patterns {
				if patMatch(v, pat) {
					found.Add(header, v)
				}
			}
		}
	}
	return
}

func parseRange(str string) (lo, hi int64) {
	parts := strings.Split(str, "-")
	if len(parts) == 1 {
		i, _ := strconv.ParseInt(parts[0], 10, 64)
		lo, hi = i, i
	} else if len(parts) == 2 {
		lo, _ = strconv.ParseInt(parts[0], 10, 64)
		hi, _ = strconv.ParseInt(parts[1], 10, 64)
	}
	return
}

// store message, unpack attachments, register with daemon, send to daemon for federation
// in that order
func storeMessage(daemon *NNTPDaemon, hdr textproto.MIMEHeader, body io.Reader) (err error) {
	var f io.WriteCloser
	msgid := getMessageID(hdr)
	log.Println("store", msgid)
	if msgid == "" {
		// drop, invalid header
		log.Println("dropping message with invalid mime header, no message-id")
		_, err = io.Copy(Discard, body)
		return
	} else if ValidMessageID(msgid) && !ReservedMessageID(msgid) {
		f = daemon.store.CreateFile(msgid)
	} else {
		// invalid message-id
		log.Println("dropping message with invalid message-id", msgid)
		_, err = io.Copy(Discard, body)
		return
	}
	if f == nil {
		// could not open file, probably already storing it from another connection
		log.Println("discarding duplicate message")
		_, err = io.Copy(Discard, body)
		return
	}

	// ask for replies
	replyTos := strings.Split(hdr.Get("In-Reply-To"), " ")
	for _, reply := range replyTos {
		if ValidMessageID(reply) && !ReservedMessageID(reply) {
			if !daemon.store.HasArticle(reply) {
				go daemon.askForArticle(reply)
			}
		}
	}

	path := hdr.Get("Path")
	hdr.Set("Path", daemon.instance_name+"!"+path)
	// do the magick
	pr, pw := io.Pipe()
	go func() {
		var buff [65536]byte
		writeMIMEHeader(pw, hdr)
		_, e := io.CopyBuffer(pw, body, buff[:])
		pw.CloseWithError(e)
	}()
	err = daemon.store.ProcessMessage(f, pr, daemon.CheckText, hdr.Get("Newsgroups"))
	pr.Close()
	f.Close()
	if err == nil {
		// tell daemon
		daemon.loadFromInfeed(msgid)
	} else {
		log.Println("error processing message body", err)
	}

	// clean up
	if ValidMessageID(msgid) {
		fname := daemon.store.GetFilenameTemp(msgid)
		DelFile(fname)
	}
	return
}

func hasAtLeastNWords(str string, n int) bool {
	parts := strings.Split(str, " ")
	return len(parts) > n
}
