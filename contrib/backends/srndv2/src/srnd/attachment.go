//
// attachment.go -- nntp attachements
//

package srnd

import (
	"bytes"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

type NNTPAttachment interface {
	io.WriterTo
	io.Writer

	// the name of the file
	Filename() string
	// the filepath to the saved file
	Filepath() string
	// the mime type of the attachment
	Mime() string
	// the file extension of the attachment
	Extension() string
	// get the sha512 hash of the attachment
	Hash() []byte
	// do we need to generate a thumbnail?
	NeedsThumbnail() bool
	// mime header
	Header() textproto.MIMEHeader
	// make into a model
	ToModel(prefix string) AttachmentModel
	// base64'd file data
	Filedata() string
	// as raw string
	AsString() string
	// reset contents
	Reset()
	// get bytes
	Bytes() []byte
	// save to directory, filename is decided by the attachment
	Save(dir string) error
	// get body as io.ReadCloser
	OpenBody() (io.ReadCloser, error)
}

type nntpAttachment struct {
	ext      string
	mime     string
	filename string
	filepath string
	hash     []byte
	header   textproto.MIMEHeader
	body     *bytes.Buffer
	rawpath  string
	store    ArticleStore
}

type byteBufferReadCloser struct {
	b *bytes.Buffer
}

func (b *byteBufferReadCloser) Close() error {
	b.b.Reset()
	return nil
}

func (b *byteBufferReadCloser) Read(d []byte) (int, error) {
	return b.b.Read(d)
}

func (self *nntpAttachment) OpenBody() (io.ReadCloser, error) {
	if self.store != nil {
		return os.Open(self.store.AttachmentFilepath(self.filepath))
	} else {
		return &byteBufferReadCloser{
			self.body,
		}, nil
	}
}

func (self *nntpAttachment) Reset() {
	self.body = nil
	self.header = nil
	self.hash = nil
	self.filepath = ""
	self.filename = ""
	self.mime = ""
	self.ext = ""
	self.store = nil
}

func (self *nntpAttachment) ToModel(prefix string) AttachmentModel {
	return &attachment{
		prefix: prefix,
		Path:   self.Filepath(),
		Name:   self.Filename(),
	}
}

func (self *nntpAttachment) Bytes() []byte {
	if self.body == nil {
		return nil
	}
	return self.body.Bytes()
}

func (self *nntpAttachment) Save(dir string) (err error) {
	if self.body == nil {
		// no body wat
		err = errors.New("no attachment body")
	} else {
		fpath := filepath.Join(dir, self.filepath)
		if !CheckFile(fpath) {
			var f io.WriteCloser
			// does not exist so will will write it
			f, err = os.Create(fpath)
			if err == nil {
				_, err = f.Write(self.Bytes())
				f.Close()
			}
		}
	}
	return
}

func (self *nntpAttachment) Write(b []byte) (int, error) {
	if self.body == nil {
		self.body = new(bytes.Buffer)
	}
	return self.body.Write(b)
}

func (self *nntpAttachment) AsString() string {
	if self.body == nil {
		return ""
	}
	return string(self.Bytes())
}

func (self *nntpAttachment) Filedata() string {
	e := base64.StdEncoding
	str := e.EncodeToString(self.Bytes())
	e = nil
	return str
}

func (self *nntpAttachment) Filename() string {
	return self.filename
}

func (self *nntpAttachment) Filepath() string {
	return self.filepath
}

func (self *nntpAttachment) Mime() string {
	return self.mime
}

func (self *nntpAttachment) Extension() string {
	return self.ext
}

func (self *nntpAttachment) WriteTo(wr io.Writer) (int64, error) {
	w, err := wr.Write(self.Bytes())
	return int64(w), err
}

func (self *nntpAttachment) Hash() []byte {
	// hash it if we haven't already
	if self.hash == nil || len(self.hash) == 0 {
		h := sha512.Sum512(self.Bytes())
		self.hash = h[:]
	}
	return self.hash
}

// TODO: detect
func (self *nntpAttachment) NeedsThumbnail() bool {
	for _, ext := range []string{".png", ".jpeg", ".jpg", ".gif", ".bmp", ".webm", ".mp4", ".avi", ".mpeg", ".mpg", ".ogg", ".mp3", ".oga", ".opus", ".flac", ".ico", "m4a"} {
		if ext == strings.ToLower(self.ext) {
			return true
		}
	}
	return false
}

func (self *nntpAttachment) Header() textproto.MIMEHeader {
	return self.header
}

// create a plaintext attachment
func createPlaintextAttachment(msg []byte) NNTPAttachment {
	header := make(textproto.MIMEHeader)
	mime := "text/plain; charset=UTF-8"
	header.Set("Content-Type", mime)
	header.Set("Content-Transfer-Encoding", "base64")
	att := &nntpAttachment{
		mime:   mime,
		ext:    ".txt",
		header: header,
	}
	msg = bytes.Trim(msg, "\r")
	att.Write(msg)
	return att
}

// assumes base64'd
func createAttachment(content_type, fname string, body io.Reader) NNTPAttachment {

	media_type, _, err := mime.ParseMediaType(content_type)
	if err == nil {
		a := new(nntpAttachment)
		dec := base64.NewDecoder(base64.StdEncoding, body)
		_, err = io.Copy(a, dec)
		if err == nil {
			a.header = make(textproto.MIMEHeader)
			a.mime = media_type + "; charset=UTF-8"
			idx := strings.LastIndex(fname, ".")
			a.ext = ".txt"
			if idx > 0 {
				a.ext = fname[idx:]
			}
			a.header.Set("Content-Disposition", `form-data; filename="`+fname+`"; name="attachment"`)
			a.header.Set("Content-Type", a.mime)
			a.header.Set("Content-Transfer-Encoding", "base64")
			h := a.Hash()
			hashstr := base32.StdEncoding.EncodeToString(h[:])
			a.hash = h[:]
			a.filepath = hashstr + a.ext
			a.filename = fname
			return a
		}
	}
	return nil
}

func readAttachmentFromMimePartAndStore(part *multipart.Part, store ArticleStore) NNTPAttachment {
	hdr := part.Header
	att := &nntpAttachment{}
	att.store = store
	att.header = hdr
	content_type := hdr.Get("Content-Type")
	var err error
	att.mime, _, err = mime.ParseMediaType(content_type)
	att.filename = part.FileName()
	idx := strings.LastIndex(att.filename, ".")
	att.ext = ".txt"
	if idx > 0 {
		att.ext = att.filename[idx:]
	}
	h := sha512.New()
	transfer_encoding := hdr.Get("Content-Transfer-Encoding")
	var r io.Reader
	if transfer_encoding == "base64" {
		// decode
		r = base64.NewDecoder(base64.StdEncoding, part)
	} else {
		r = part
	}
	var fpath string
	var mw io.Writer
	if store == nil {
		mw = io.MultiWriter(att, h)
	} else {
		fname := randStr(10) + ".temp"
		fpath = filepath.Join(store.AttachmentDir(), fname)
		f, err := os.Create(fpath)
		if err != nil {
			log.Println("!!! failed to store attachment: ", err, "!!!")
			return nil
		}
		defer f.Close()
		if strings.ToLower(att.mime) == "text/plain" {
			mw = io.MultiWriter(f, h, att)
		} else {
			mw = io.MultiWriter(f, h)
		}
	}
	_, err = io.Copy(mw, r)
	if err != nil {
		log.Println("failed to read attachment from mimepart", err)
		if fpath != "" {
			DelFile(fpath)
		}
		return nil
	}
	hsh := h.Sum(nil)
	att.hash = hsh[:]
	enc := base32.StdEncoding
	hashstr := enc.EncodeToString(att.hash[:])
	att.filepath = hashstr + att.ext
	// we are good just return it
	if store == nil {
		return att
	}
	att_fpath := filepath.Join(store.AttachmentDir(), att.filepath)
	if !CheckFile(att_fpath) {
		// attachment isn't there
		// move it into it
		err = os.Rename(fpath, att_fpath)
	}
	if err == nil {
		// now thumbnail
		if !CheckFile(store.ThumbnailFilepath(att.filepath)) {
			store.GenerateThumbnail(att.filepath)
		}
	} else {
		// wtf?
		log.Println("!!! failed to store attachment", err, "!!!")
		DelFile(fpath)
	}
	return att
}
