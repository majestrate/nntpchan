//
// message.go
//
package srnd

import (
	"bufio"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/majestrate/nacl"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"strings"
	"time"
)

type ArticleHeaders map[string][]string

func (self ArticleHeaders) Has(key string) bool {
	_, ok := self[key]
	return ok
}

func (self ArticleHeaders) Set(key, val string) {
	self[key] = []string{val}
}

func (self ArticleHeaders) Add(key, val string) {
	if self.Has(key) {
		self[key] = append(self[key], val)
	} else {
		self.Set(key, val)
	}
}

func (self ArticleHeaders) Get(key, fallback string) string {
	val, ok := self[key]
	if ok {
		str := ""
		for _, k := range val {
			str += k + ", "
		}
		return str[:len(str)-2]
	} else {
		return fallback
	}
}

type NNTPMessage interface {
	// this message's messsge id
	MessageID() string
	// the parent message's messageid if it's specified
	Reference() string
	// the newsgroup this post is in
	Newsgroup() string
	// the name of the poster
	Name() string
	// any email address associated with the post
	Email() string
	// the subject of the post
	Subject() string
	// when this was posted
	Posted() int64
	// the path header
	Path() string
	// get signed part
	SignedPart() NNTPAttachment
	// append something to path
	// return message with new path
	AppendPath(part string) NNTPMessage
	// the type of this message usually a mimetype
	ContentType() string
	// was this post a sage?
	Sage() bool
	// was this post a root post?
	OP() bool
	// all attachments
	Attachments() []NNTPAttachment
	// all headers
	Headers() ArticleHeaders
	// write out everything
	WriteTo(wr io.Writer) error
	// write out body
	WriteBody(wr io.Writer) error
	// attach a file
	Attach(att NNTPAttachment)
	// get the plaintext message if it exists
	Message() string
	// pack the whole message and prepare for write
	Pack()
	// get the pubkey for this message if it was signed, otherwise empty string
	Pubkey() string
	// get the origin encrypted address, i2p destination or empty string for onion posters
	Addr() string
	// reset contents
	Reset()
}

type nntpArticle struct {
	// mime header
	headers ArticleHeaders
	// multipart boundary
	boundary string
	// the text part of the message
	message string
	// any attachments
	attachments []NNTPAttachment
	// the inner nntp message to be verified
	signedPart *nntpAttachment
}

func (self *nntpArticle) Reset() {
	self.headers = nil
	self.boundary = ""
	self.message = ""
	if self.attachments != nil {
		for idx, _ := range self.attachments {
			self.attachments[idx].Reset()
			self.attachments[idx] = nil
		}
	}
	self.attachments = nil
	if self.signedPart != nil {
		self.signedPart.Reset()
		self.signedPart = nil
	}
}

func (self *nntpArticle) SignedPart() NNTPAttachment {
	return self.signedPart
}

// create a simple plaintext nntp message
func newPlaintextArticle(message, email, subject, name, instance, message_id, newsgroup string) NNTPMessage {
	nntp := &nntpArticle{
		headers: make(ArticleHeaders),
	}
	nntp.headers.Set("From", fmt.Sprintf("%s <%s>", name, email))
	nntp.headers.Set("Subject", subject)
	if isSage(subject) {
		nntp.headers.Set("X-Sage", "1")
	}
	nntp.headers.Set("Path", instance)
	nntp.headers.Set("Message-ID", message_id)
	// posted now
	nntp.headers.Set("Date", timeNowStr())
	nntp.headers.Set("Newsgroups", newsgroup)
	nntp.message = strings.Trim(message, "\r")
	nntp.Pack()
	return nntp
}

// sign an article with a seed
func signArticle(nntp NNTPMessage, seed []byte) (signed *nntpArticle, err error) {
	signed = new(nntpArticle)
	signed.headers = make(ArticleHeaders)
	h := nntp.Headers()
	// copy headers
	// copy into signed part
	for k := range h {
		if k == "X-PubKey-Ed25519" || k == "X-Signature-Ed25519-SHA512" {
			// don't set signature or pubkey header
		} else if k == "Content-Type" {
			signed.headers.Set(k, "message/rfc822; charset=UTF-8")
		} else {
			v := h[k][0]
			signed.headers.Set(k, v)
		}
	}
	sha := sha512.New()
	signed.signedPart = &nntpAttachment{}
	// write body to sign buffer
	mw := io.MultiWriter(sha, signed.signedPart)
	err = nntp.WriteTo(mw)
	mw.Write([]byte{10})
	if err == nil {
		// build keypair
		kp := nacl.LoadSignKey(seed)
		if kp == nil {
			log.Println("failed to load seed for signing article")
			return
		}
		defer kp.Free()
		sk := kp.Secret()
		pk := getSignPubkey(sk)
		// sign it nigguh
		digest := sha.Sum(nil)
		sig := cryptoSign(digest, sk)
		// log that we signed it
		log.Printf("signed %s pubkey=%s sig=%s hash=%s", nntp.MessageID(), pk, sig, hexify(digest))
		signed.headers.Set("X-Signature-Ed25519-SHA512", sig)
		signed.headers.Set("X-PubKey-Ed25519", pk)
	}
	return
}

func (self *nntpArticle) WriteTo(wr io.Writer) (err error) {
	// write headers
	hdrs := self.headers
	for hdr, hdr_vals := range hdrs {
		for _, hdr_val := range hdr_vals {
			wr.Write([]byte(hdr))
			wr.Write([]byte(": "))
			wr.Write([]byte(hdr_val))
			_, err = wr.Write([]byte{10})
			if err != nil {
				log.Println("error while writing headers", err)
				return
			}
		}
	}
	// done headers
	_, err = wr.Write([]byte{10})
	if err != nil {
		log.Println("error while writing body", err)
		return
	}

	// write body
	err = self.WriteBody(wr)
	return
}

func (self *nntpArticle) Pubkey() string {
	return self.headers.Get("X-PubKey-Ed25519", self.headers.Get("X-Pubkey-Ed25519", ""))
}

func (self *nntpArticle) MessageID() (msgid string) {
	for _, h := range []string{"Message-ID", "Messageid", "MessageID", "Message-Id"} {
		mid := self.headers.Get(h, "")
		if mid != "" {
			msgid = string(mid)
			return
		}
	}
	return
}

func (self *nntpArticle) Pack() {
	if len(self.attachments) > 0 {
		if len(self.boundary) == 0 {
			// we have no boundry, set it
			self.boundary = randStr(24)
			// set headers
			self.headers.Set("Mime-Version", "1.0")
			self.headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", self.boundary))
		}
	} else if self.signedPart == nil {
		self.headers.Set("Content-Type", "text/plain; charset=utf-8")
	}
}

func (self *nntpArticle) Reference() string {
	return self.headers.Get("Reference", self.headers.Get("References", ""))
}

func (self *nntpArticle) Newsgroup() string {
	return self.headers.Get("Newsgroups", "")
}

func (self *nntpArticle) Name() string {
	from := self.headers.Get("From", "anonymous <a@no.n>")
	idx := strings.Index(from, "<")
	if idx > 1 {
		return from[:idx]
	}
	return "[Invalid From header]"
}

func (self *nntpArticle) Addr() (addr string) {
	addr = self.headers.Get("X-Encrypted-Ip", "")
	if addr != "" {
		return
	}

	addr = self.headers.Get("X-Encrypted-IP", "")
	if addr != "" {
		return
	}

	addr = self.headers.Get("X-I2P-DestHash", "")
	if addr != "" {
		if addr == "None" {
			return ""
		}
		return
	}

	addr = self.headers.Get("X-I2p-Desthash", "")
	return
}

func (self *nntpArticle) Email() string {
	from := self.headers.Get("From", "anonymous <a@no.n>")
	idx := strings.Index(from, "<")
	if idx > 2 {
		return from[:idx-2]
	}
	return "[Invalid From header]"

}

func (self *nntpArticle) Subject() string {
	return self.headers.Get("Subject", "")
}

func (self *nntpArticle) Posted() int64 {
	posted := self.headers.Get("Date", "")
	t, err := time.Parse(time.RFC1123Z, posted)
	if err == nil {
		return t.Unix()
	}
	return 0
}

func (self *nntpArticle) Message() string {
	return strings.Trim(self.message, "\x00")
}

func (self *nntpArticle) Path() string {
	return self.headers.Get("Path", "unspecified")
}

func (self *nntpArticle) Headers() ArticleHeaders {
	return self.headers
}

func (self *nntpArticle) AppendPath(part string) NNTPMessage {
	if self.headers.Has("Path") {
		self.headers.Set("Path", part+"!"+self.Path())
	} else {
		self.headers.Set("Path", part)
	}
	return self
}
func (self *nntpArticle) ContentType() string {
	// assumes text/plain if unspecified
	return self.headers.Get("Content-Type", "text/plain; charset=UTF-8")
}

func (self *nntpArticle) Sage() bool {
	return self.headers.Get("X-Sage", "") == "1"
}

func (self *nntpArticle) OP() bool {
	ref := self.Reference()
	return ref == "" || ref == self.MessageID()
}

func (self *nntpArticle) Attachments() []NNTPAttachment {
	return self.attachments
}

func (self *nntpArticle) Attach(att NNTPAttachment) {
	self.attachments = append(self.attachments, att)
}

func (self *nntpArticle) WriteBody(wr io.Writer) (err error) {
	// this is a signed message, don't treat it special
	if self.signedPart != nil {
		_, err = wr.Write(self.signedPart.Bytes())
		return
	}
	self.Pack()
	content_type := self.ContentType()
	_, params, err := mime.ParseMediaType(content_type)
	if err != nil {
		log.Println("failed to parse media type", err)
		return err
	}

	boundary, ok := params["boundary"]
	if ok {
		w := multipart.NewWriter(NewLineWriter(wr))

		err = w.SetBoundary(boundary)
		if err == nil {
			attachments := []NNTPAttachment{createPlaintextAttachment([]byte(self.message))}
			attachments = append(attachments, self.attachments...)
			for _, att := range attachments {
				if att == nil {
					continue
				}
				hdr := att.Header()
				hdr.Add("Content-Transfer-Encoding", "base64")
				part, err := w.CreatePart(hdr)
				if err != nil {
					log.Println("failed to create part?", err)
				}
				var buff [1024]byte
				var b io.ReadCloser
				b, err = att.OpenBody()
				if err == nil {
					enc := base64.NewEncoder(base64.StdEncoding, part)
					_, err = io.CopyBuffer(enc, b, buff[:])
					b.Close()
					enc.Close()
					if err != nil {
						break
					}
				}
				part = nil
			}
		}
		if err != nil {
			log.Println("error writing part", err)
		}
		err = w.Close()
		w = nil
	} else {
		// write out message
		_, err = io.WriteString(wr, self.message)
	}
	return err
}

// verify a signed message's body
// innerHandler must close reader when done
// returns error if one happens while verifying article
func verifyMessage(pk, sig string, body io.Reader, innerHandler func(map[string][]string, io.Reader)) (err error) {
	log.Println("unwrapping signed message from", pk)
	pk_bytes := unhex(pk)
	sig_bytes := unhex(sig)
	h := sha512.New()
	pr, pw := io.Pipe()
	// read header
	// handle inner body
	go func(hdr_reader *io.PipeReader) {
		r := bufio.NewReader(hdr_reader)
		msg, err := readMIMEHeader(r)
		if err == nil {
			innerHandler(msg.Header, msg.Body)
		}
		hdr_reader.Close()
	}(pr)
	body = io.TeeReader(body, pw)
	// copy body 128 bytes at a time
	var buff [128]byte
	_, err = io.CopyBuffer(h, body, buff[:])
	if err == nil {
		hash := h.Sum(nil)
		log.Printf("hash=%s", hexify(hash))
		log.Printf("sig=%s", hexify(sig_bytes))
		if nacl.CryptoVerifyFucky(hash, sig_bytes, pk_bytes) {
			log.Println("signature is valid :^)")
		} else {
			err = errors.New("invalid signature")
		}
	}
	// flush pipe
	pw.Close()
	return
}
