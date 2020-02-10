//
// store.go
//

package srnd

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrOversizedMessage = errors.New("oversized message")

// (cathugger)
// my test showed that 8MiB of attachments split in 5 parts
// plus some text produce something close to typhical big message
// resulted in 11483923 bytes.
// that's consistent with rough size calculation mentioned in
// <https://en.wikipedia.org/wiki/Base64#MIME>
// ((origlen * 1.37) + 814)
// which resulted in 11493206 bytes for 8MiB of data.
// previous default of 10MiB (10485760) was too low in practice.
// use 11MiB (11534336) to leave some space for longer than usual texts.
const DefaultMaxMessageSize = 11 * 1024 * 1024

// HARD max message size
const MaxMessageSize = 1024 * 1024 * 1024

type ArticleStore interface {

	// full filepath to attachment directory
	AttachmentDir() string

	// get the filepath for an attachment
	AttachmentFilepath(fname string) string
	// get the filepath for an attachment's thumbnail
	ThumbnailFilepath(fname string) string
	// do we have this article?
	HasArticle(msgid string) bool
	// create a file for a message
	CreateFile(msgid string) io.WriteCloser
	// get the filename of a message
	GetFilename(msgid string) string
	// open a message in the store for reading given its message-id
	// return io.ReadCloser, error
	OpenMessage(msgid string) (io.ReadCloser, error)
	// get article headers only
	GetHeaders(msgid string) ArticleHeaders
	// get mime header
	GetMIMEHeader(msgid string) textproto.MIMEHeader
	// get our temp directory for articles
	TempDir() string
	// get temp filename for article
	GetFilenameTemp(msgid string) string
	// get a list of all the attachments we have
	GetAllAttachments() ([]string, error)
	// generate a thumbnail
	GenerateThumbnail(fname string) (ThumbInfo, error)
	// generate all thumbanils for this message
	ThumbnailMessage(msgid string) []ThumbInfo
	// did we enable compression?
	Compression() bool
	// process nntp message, register attachments and the article
	// write the body into writer as we go through the message
	// writes mime body and does any spam rewrite
	ProcessMessage(wr io.Writer, msg io.Reader, filter func(string) bool, group string) error
	// register this post with the daemon
	RegisterPost(nntp NNTPMessage) error
	// register signed message
	RegisterSigned(msgid, pk string) error

	GetMessage(msgid string, visit func(NNTPMessage))

	// get size of message on disk
	GetMessageSize(msgid string) (int64, error)

	// get thumbnail info of file by path
	ThumbInfo(fpath string) (ThumbInfo, error)

	// delete message by message-id
	Remove(msgid string) error

	// move message to spam dir
	MarkSpam(msgid string) error

	// move message out of spam dir
	UnmarkSpam(msgid string) error

	// get filepath for spam file via msgid
	SpamFile(msgid string) string

	// iterate over all spam messages
	IterSpam(func(string) error) error

	// iterate over all spam message headers
	IterSpamHeaders(func(map[string][]string) error) error

	// move temp article to article store
	AcceptTempArticle(msgid string) error
}
type articleStore struct {
	directory     string
	temp          string
	attachments   string
	thumbs        string
	database      Database
	convert_path  string
	ffmpeg_path   string
	sox_path      string
	identify_path string
	placeholder   string
	spamdir       string
	hamdir        string
	compression   bool
	compWriter    *gzip.Writer
	spamd         *SpamFilter
	thumbnails    *ThumbnailConfig
}

func createArticleStore(config map[string]string, thumbConfig *ThumbnailConfig, database Database, spamd *SpamFilter) ArticleStore {
	store := &articleStore{
		directory:     config["store_dir"],
		temp:          config["incoming_dir"],
		attachments:   config["attachments_dir"],
		thumbs:        config["thumbs_dir"],
		convert_path:  config["convert_bin"],
		identify_path: config["identify_bin"],
		ffmpeg_path:   config["ffmpegthumbnailer_bin"],
		sox_path:      config["sox_bin"],
		placeholder:   config["placeholder_thumbnail"],
		database:      database,
		compression:   config["compression"] == "1",
		spamd:         spamd,
		spamdir:       filepath.Join(config["store_dir"], "spam"),
		hamdir:        filepath.Join(config["store_dir"], "ham"),

		thumbnails: thumbConfig,
	}
	store.Init()
	return store
}

func (self *articleStore) AttachmentDir() string {
	return self.attachments
}

func (self *articleStore) Compression() bool {
	return self.compression
}

func (self *articleStore) TempDir() string {
	return self.temp
}

// initialize article store
func (self *articleStore) Init() {
	EnsureDir(self.directory)
	EnsureDir(self.temp)
	EnsureDir(self.attachments)
	EnsureDir(self.thumbs)
	EnsureDir(self.spamdir)
	if !CheckFile(self.convert_path) {
		log.Fatal("cannot find executable for convert: ", self.convert_path, " not found")
	}
	if !CheckFile(self.ffmpeg_path) {
		log.Fatal("connt find executable for ffmpegthumbnailer: ", self.ffmpeg_path, " not found")
	}
	if !CheckFile(self.sox_path) {
		log.Fatal("connt find executable for sox: ", self.sox_path, " not found")
	}
	if !CheckFile(self.identify_path) {
		log.Fatal("cannot find executable for identify: ", self.identify_path, "not found")
	}

	if !CheckFile(self.placeholder) {
		log.Println("falling back to use default placeholder image")
		self.placeholder = "contrib/static/placeholder.png"
		if !CheckFile(self.placeholder) {
			log.Fatal("cannot find thumbnail placeholder file: ", self.placeholder, " not found")
		}
	}
}

func (self *articleStore) Remove(msgid string) (err error) {
	if ValidMessageID(msgid) {
		fpath := self.GetFilename(msgid)
		err = os.Remove(fpath)
	} else {
		err = errors.New("invalid message-id: " + msgid)
	}
	return
}

func (self *articleStore) RegisterSigned(msgid, pk string) (err error) {
	err = self.database.RegisterSigned(msgid, pk)
	return
}

func (self *articleStore) isAudio(fname string) bool {
	for _, ext := range []string{".mp3", ".ogg", ".oga", ".opus", ".flac", ".m4a"} {
		if strings.HasSuffix(strings.ToLower(fname), ext) {
			return true
		}
	}
	return false
}

func (self *articleStore) ThumbnailMessage(msgid string) (infos []ThumbInfo) {
	atts := self.database.GetPostAttachments(msgid)
	for _, att := range atts {
		if CheckFile(self.ThumbnailFilepath(att)) {
			continue
		}
		info, err := self.GenerateThumbnail(att)
		if err == nil {
			infos = append(infos, info)
		}
	}
	return
}

// is this an image format we need convert for?
func (self *articleStore) isImage(fname string) bool {
	for _, ext := range []string{".gif", ".ico", ".png", ".jpeg", ".jpg", ".png", ".webp"} {
		if strings.HasSuffix(strings.ToLower(fname), ext) {
			return true
		}
	}
	return false
}

// is this a video file?
func (self *articleStore) isVideo(fname string) bool {
	for _, ext := range []string{".mpeg", ".ogv", ".mkv", ".avi", ".mp4", ".webm"} {
		if strings.HasSuffix(strings.ToLower(fname), ext) {
			return true
		}
	}
	return false
}

func (self *articleStore) ThumbInfo(fpath string) (ThumbInfo, error) {
	var info ThumbInfo
	cmd := exec.Command(self.identify_path, "-format", "%[fx:w] %[fx:h]", fpath)
	output, err := cmd.Output()
	if err == nil {
		parts := strings.Split(string(output), " ")
		if len(parts) == 2 {
			info.Width, err = strconv.Atoi(parts[0])
			if err == nil {
				info.Height, err = strconv.Atoi(parts[1])
			}
		}
	} else {
		log.Println("failed to determine size of thumbnail", err, string(output))
	}
	return info, err
}

func (self *articleStore) GenerateThumbnail(fname string) (info ThumbInfo, err error) {
	outfname := self.ThumbnailFilepath(fname)
	if self.thumbnails == nil {
		err = self.generateThumbnailFallback(fname)
		if err == nil {
			info, err = self.ThumbInfo(outfname)
		}
		return
	}
	infname := self.AttachmentFilepath(fname)
	err = self.thumbnails.GenerateThumbnail(infname, outfname, map[string]string{
		"ffmpeg":      self.ffmpeg_path,
		"convert":     self.convert_path,
		"sox":         self.sox_path,
		"identify":    self.identify_path,
		"placeholder": self.placeholder,
	})
	if err != nil {
		log.Println(err.Error(), "so we'll use fallback thumbnailing")
		err = self.generateThumbnailFallback(fname)
	}
	if err == nil {
		info, err = self.ThumbInfo(outfname)
	}
	return
}

func (self *articleStore) generateThumbnailFallback(fname string) (err error) {
	outfname := self.ThumbnailFilepath(fname)
	infname := self.AttachmentFilepath(fname)
	tmpfname := ""
	var cmd *exec.Cmd
	if self.isImage(fname) {
		if strings.HasSuffix(infname, ".gif") {
			infname += "[0]"
		}
		cmd = exec.Command(self.convert_path, "-thumbnail", "200", infname, outfname)
	} else if self.isAudio(fname) {
		tmpfname = infname + ".wav"
		cmd = exec.Command(self.ffmpeg_path, "-i", infname, tmpfname)
		var out []byte

		out, err = cmd.CombinedOutput()

		if err == nil {
			cmd = exec.Command(self.sox_path, tmpfname, "-n", "spectrogram", "-a", "-d", "0:10", "-r", "-p", "6", "-x", "200", "-y", "150", "-o", outfname)
		} else {
			log.Println("error making thumbnail", string(out))
		}

	} else if self.isVideo(fname) || strings.HasSuffix(fname, ".txt") {
		cmd = exec.Command(self.ffmpeg_path, "-i", infname, "-vf", "scale=300:200", "-vframes", "1", outfname)
	}
	if cmd == nil {
		log.Println("use placeholder for", infname)
		os.Link(self.placeholder, outfname)
	} else {
		exec_out, err := cmd.CombinedOutput()
		if err == nil {
			log.Println("made thumbnail for", infname)
		} else {
			log.Println("error generating thumbnail", string(exec_out))
		}
	}
	if len(tmpfname) > 0 {
		DelFile(tmpfname)
	}
	return
}

func (self *articleStore) GetAllAttachments() (names []string, err error) {
	var f *os.File
	f, err = os.Open(self.attachments)
	if err == nil {
		names, err = f.Readdirnames(0)
	}
	return
}

func (self *articleStore) OpenMessage(msgid string) (rc io.ReadCloser, err error) {
	fname := self.GetFilename(msgid)
	var f *os.File
	f, err = os.Open(fname)
	if err == nil {
		if self.compression {
			// read gzip header
			var hdr [2]byte
			_, err = f.Read(hdr[:])
			// seek back to beginning
			f.Seek(0, 0)
			if err == nil {
				if hdr[0] == 0x1f && hdr[1] == 0x8b {
					// gzip header detected
					rc, err = gzip.NewReader(f)
				} else {
					// fall back to uncompressed
					rc = f
				}
			} else {
				// error reading file
				f.Close()
				rc = nil
			}
			// will fall back to regular file if gzip header not found
		} else {
			// compression disabled
			// assume uncompressed
			rc = f
		}
	}
	return
}

func (self *articleStore) RegisterPost(nntp NNTPMessage) (err error) {
	err = self.database.RegisterArticle(nntp)
	return
}

func (self *articleStore) saveAttachment(att NNTPAttachment) {
	fpath := att.Filepath()
	upload := self.AttachmentFilepath(fpath)
	if !CheckFile(upload) {
		// attachment does not exist on disk
		f, err := os.Create(upload)
		if f != nil {
			_, err = att.WriteTo(f)
			f.Close()
		}
		if err != nil {
			log.Println("failed to save attachemnt", fpath, err)
		}
	}
	att.Reset()
	self.thumbnailAttachment(fpath)
}

// generate attachment thumbnail
func (self *articleStore) thumbnailAttachment(fpath string) {
	thumb := self.ThumbnailFilepath(fpath)
	if !CheckFile(thumb) {
		_, err := self.GenerateThumbnail(fpath)
		if err != nil {
			log.Println("failed to generate thumbnail for", fpath, err)
		}
	}
}

func (self *articleStore) GetMessageSize(msgid string) (sz int64, err error) {
	var info os.FileInfo
	info, err = os.Stat(self.GetFilename(msgid))
	if err == nil {
		sz = info.Size()
	}
	return
}

// get the filepath for an attachment
func (self *articleStore) AttachmentFilepath(fname string) string {
	return filepath.Join(self.attachments, fname)
}

// get the filepath for a thumbanil
func (self *articleStore) ThumbnailFilepath(fname string) string {
	// all thumbnails are jpegs now
	//if strings.HasSuffix(fname, ".gif") {
	//	return filepath.Join(self.thumbs, fname)
	//}
	return filepath.Join(self.thumbs, fname+".jpg")
}

func (self *articleStore) GetFilenameTemp(msgid string) (fpath string) {
	if ValidMessageID(msgid) {
		fpath = filepath.Join(self.TempDir(), msgid)
	}
	return
}

// create a file for this article
func (self *articleStore) CreateFile(messageID string) io.WriteCloser {
	fname := self.GetFilenameTemp(messageID)
	if CheckFile(fname) {
		// already exists
		log.Println("article with message-id", messageID, "already exists, not saving")
		return nil
	}
	file, err := os.Create(fname)
	if err != nil {
		log.Println("cannot open file", fname)
		return nil
	}
	return file
}

// return true if we have an article
func (self *articleStore) HasArticle(messageID string) bool {
	return CheckFile(self.GetFilename(messageID)) || CheckFile(self.SpamFile(messageID)) || CheckFile(self.GetFilenameTemp(messageID))
}

// get the filename for this article
func (self *articleStore) GetFilename(messageID string) string {
	if !ValidMessageID(messageID) {
		log.Println("!!! bug: tried to open invalid message", messageID, "!!!")
		return ""
	}
	return filepath.Join(self.directory, messageID)
}

func (self *articleStore) GetHeaders(messageID string) (hdr ArticleHeaders) {
	txthdr := self.getMIMEHeader(messageID)
	if txthdr != nil {
		hdr = make(ArticleHeaders)
		for k, val := range txthdr {
			for _, v := range val {
				hdr.Add(k, v)
			}
		}
	}
	return
}

func (self *articleStore) GetMIMEHeader(messageID string) textproto.MIMEHeader {
	return self.getMIMEHeader(messageID)
}

func (self *articleStore) getMIMEHeader(msgid string) (hdr textproto.MIMEHeader) {
	if ValidMessageID(msgid) {
		hdr = self.getMIMEHeaderByFile(self.GetFilename(msgid))
	}
	return
}

func (self *articleStore) getMIMEHeaderByFile(fname string) (hdr map[string][]string) {
	f, err := os.Open(fname)
	if f != nil {
		r := bufio.NewReader(f)
		var msg *mail.Message
		msg, err = readMIMEHeader(r)
		f.Close()
		if msg != nil {
			hdr = msg.Header
		}
	}
	if err != nil {
		log.Println("failed to load article headers from", fname, err)
	}
	return
}

func (self *articleStore) ProcessMessage(wr io.Writer, msg io.Reader, spamfilter func(string) bool, group string) (err error) {
	process := func(nntp NNTPMessage) {
		if !spamfilter(nntp.Message()) {
			err = errors.New("spam message")
			return
		}
		hdr := nntp.MIMEHeader()
		err = self.RegisterPost(nntp)
		if err == nil {
			pk := hdr.Get("X-PubKey-Ed25519")
			if len(pk) > 0 {
				// signed and valid
				err = self.RegisterSigned(getMessageID(hdr), pk)
				if err != nil {
					log.Println("register signed failed", err)
				}
			}
		} else {
			log.Println("error procesing message body", err)
		}
	}
	if self.spamd.Enabled(group) {
		pr_in, pw_in := io.Pipe()
		pr_out, pw_out := io.Pipe()
		resc := make(chan SpamResult)
		go func() {
			res := self.spamd.Rewrite(pr_in, pw_out, group)
			resc <- res
		}()
		go func() {
			var buff [65536]byte

			_, e := io.CopyBuffer(pw_in, msg, buff[:])
			if e != nil {
				log.Println("failed to read entire message", e)
			}
			pw_in.CloseWithError(e)
			pr_in.CloseWithError(e)
		}()
		r := bufio.NewReader(pr_out)
		m, e := readMIMEHeader(r)
		err = e
		defer func() {
			pr_out.Close()
		}()
		if err != nil {
			return
		}
		msgid := getMessageID(m.Header)
		writeMIMEHeader(wr, m.Header)
		err = read_message_body(m.Body, m.Header, self, wr, false, process)
		spamRes := <-resc
		if spamRes.Err != nil {
			return spamRes.Err
		}

		if spamRes.IsSpam {
			err = self.MarkSpam(msgid)
		}

	} else {
		r := bufio.NewReader(msg)
		m, e := readMIMEHeader(r)
		err = e
		if err != nil {
			return
		}
		writeMIMEHeader(wr, m.Header)
		err = read_message_body(m.Body, m.Header, self, wr, false, process)
	}
	return
}

func (self *articleStore) GetMessage(msgid string, visit func(NNTPMessage)) {
	r, err := self.OpenMessage(msgid)
	if err == nil {
		defer r.Close()
		br := bufio.NewReader(r)
		msg, err := readMIMEHeader(br)
		if err == nil {
			hdr := textproto.MIMEHeader(msg.Header)
			err = read_message_body(msg.Body, hdr, nil, nil, true, func(n NNTPMessage) {
				if n != nil {
					// inject pubkey for mod
					n.Headers().Set("X-PubKey-Ed25519", hdr.Get("X-PubKey-Ed25519"))
				}
				visit(n)
			})
		}
	}
	return
}

// read message body with mimeheader pre-read
// calls callback for each read nntp message
// if writer is not nil and discardAttachmentBody is false the message body will be written to the writer and the nntp message will not be filled
// if writer is not nil and discardAttachmentBody is true the message body will be discarded and writer ignored
// if writer is nil and discardAttachmentBody is true the body is discarded entirely
// if writer is nil and discardAttachmentBody is false the body is loaded into the nntp message
// if the body contains a signed message it unrwarps 1 layer of signing
func read_message_body(body io.Reader, hdr map[string][]string, store ArticleStore, wr io.Writer, discardAttachmentBody bool, callback func(NNTPMessage)) error {
	nntp := new(nntpArticle)
	nntp.headers = ArticleHeaders(hdr)
	content_type := nntp.ContentType()
	media_type, params, err := mime.ParseMediaType(content_type)
	if err != nil {
		log.Println("failed to parse media type", err, "for mime", content_type)
		nntp.Reset()
		return err
	}
	if wr != nil && !discardAttachmentBody {
		body = io.TeeReader(body, wr)
	}
	boundary, ok := params["boundary"]
	if strings.HasPrefix(media_type, "multipart/") && ok {
		partReader := multipart.NewReader(body, boundary)
		for {
			part, err := partReader.NextPart()
			if part == nil && err == io.EOF {
				callback(nntp)
				return nil
			} else if err == nil {
				hdr := part.Header
				// get content type of part
				part_type := strings.TrimSpace(hdr.Get("Content-Type"))
				if part_type == "" {
					// default if unspecified
					part_type = "text/plain"
				}
				// parse content type
				media_type, _, err = mime.ParseMediaType(part_type)
				if err == nil {
					if media_type == "text/plain" {
						att := readAttachmentFromMimePartAndStore(part, store)
						if att == nil {
							log.Println("failed to load plaintext attachment")
						} else {
							if att.Filename() == "" {
								// message part
								nntp.message = att.AsString()
							} else {
								// plaintext attachment
								nntp.Attach(att)
							}
						}
					} else {
						// non plaintext gets added to attachments
						att := readAttachmentFromMimePartAndStore(part, store)
						if att == nil {
							// failed to read attachment
							log.Println("failed to read attachment of type", media_type)
						} else {
							nntp.Attach(att)
						}
					}
				} else {
					log.Println("part has no content type", err)
				}
				part.Close()
				part = nil
			} else {
				log.Println("failed to load part! ", err)
				nntp.Reset()
				return err
			}
		}
	} else if media_type == "message/rfc822" {
		// tripcoded message
		sig := nntp.headers.Get("X-Signature-Ed25519-Sha512", "")
		var blake bool
		if sig == "" {
			sig = nntp.headers.Get("X-Signature-Ed25519-Blake2b", "")
			blake = sig != ""
		}
		pk := nntp.Pubkey()
		if (pk == "" || sig == "") && !blake {
			log.Println("invalid sig or pubkey", sig, pk)
			nntp.Reset()
			return errors.New("invalid headers")
		}
		// process inner body
		// verify message
		f := func(h ArticleHeaders, innerBody io.Reader) {
			// override some of headers of inner message
			msgid := nntp.MessageID()
			if msgid != "" {
				h.Set("Message-Id", msgid)
			}
			h.Set("Path", nntp.headers.Get("Path", ""))
			h.Set("X-Pubkey-Ed25519", pk)
			// handle inner message
			e := read_message_body(innerBody, h, store, nil, true, callback)
			if e != nil {
				log.Println("error reading inner signed message", e)
			}
		}
		if blake {
			err = verifyMessageBLAKE2B(pk, sig, body, f)
		} else {
			err = verifyMessageSHA512(pk, sig, body, f)
		}
		if err != nil {
			log.Println("error reading inner message", err)
		}
	} else {
		// plaintext attachment
		b := new(bytes.Buffer)
		_, err = io.Copy(b, body)
		if err == nil {
			nntp.message = b.String()
			callback(nntp)
		}
	}
	return err
}

func (self *articleStore) SpamFile(msgid string) string {
	return filepath.Join(self.spamdir, msgid)
}

func (self *articleStore) MarkSpam(msgid string) (err error) {
	if ValidMessageID(msgid) {
		err = os.Rename(self.GetFilename(msgid), self.SpamFile(msgid))
	}
	return
}

func (self *articleStore) UnmarkSpam(msgid string) (err error) {
	if ValidMessageID(msgid) {
		err = os.Rename(self.SpamFile(msgid), self.GetFilename(msgid))
	}
	return
}

func (self *articleStore) iterSpamFiles(v func(os.FileInfo) error) error {
	infos, err := ioutil.ReadDir(self.spamdir)
	if err == nil {
		for idx := range infos {
			err = v(infos[idx])
			if err != nil {
				break
			}
		}
	}
	return err
}

func (self *articleStore) IterSpam(v func(string) error) error {
	return self.iterSpamFiles(func(i os.FileInfo) error {
		fname := i.Name()
		if ValidMessageID(fname) {
			return v(fname)
		}
		return nil
	})
}

func (self *articleStore) IterSpamHeaders(v func(map[string][]string) error) error {
	return self.IterSpam(func(msgid string) error {
		hdr := self.getMIMEHeaderByFile(self.SpamFile(msgid))
		if hdr != nil {
			return v(hdr)
		}
		return nil
	})
}

func (self *articleStore) AcceptTempArticle(msgid string) (err error) {
	if ValidMessageID(msgid) {
		temp := self.GetFilenameTemp(msgid)
		store := self.GetFilename(msgid)
		if CheckFile(temp) {
			if CheckFile(store) {
				// already in store
				err = os.Remove(temp)
			} else {
				err = os.Rename(temp, store)
			}
		} else {
			err = fmt.Errorf("no such inbound article %s", temp)
		}
	} else {
		err = fmt.Errorf("invalid message id %s", msgid)
	}
	return
}
