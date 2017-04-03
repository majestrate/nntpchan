//
// store.go
//

package srnd

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
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
	// get a list of all the attachments we have
	GetAllAttachments() ([]string, error)
	// generate a thumbnail
	GenerateThumbnail(fname string) (ThumbInfo, error)
	// generate all thumbanils for this message
	ThumbnailMessage(msgid string) []ThumbInfo
	// did we enable compression?
	Compression() bool
	// process body of nntp message, register attachments and the article
	// write the body into writer as we go through the body
	// does NOT write mime header
	ProcessMessageBody(wr io.Writer, hdr textproto.MIMEHeader, body io.Reader) error
	// register this post with the daemon
	RegisterPost(nntp NNTPMessage) error
	// register signed message
	RegisterSigned(msgid, pk string) error

	GetMessage(msgid string) NNTPMessage

	// get size of message on disk
	GetMessageSize(msgid string) (int64, error)

	// get thumbnail info of file by path
	ThumbInfo(fpath string) (ThumbInfo, error)
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
	compression   bool
	compWriter    *gzip.Writer
}

func createArticleStore(config map[string]string, database Database) ArticleStore {
	store := &articleStore{
		directory:     config["store_dir"],
		temp:          config["incoming_dir"],
		attachments:   config["attachments_dir"],
		thumbs:        config["thumbs_dir"],
		convert_path:  config["convert_bin"],
		identify_path: config["identify_path"],
		ffmpeg_path:   config["ffmpegthumbnailer_bin"],
		sox_path:      config["sox_bin"],
		placeholder:   config["placeholder_thumbnail"],
		database:      database,
		compression:   config["compression"] == "1",
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
	if !CheckFile(self.convert_path) {
		log.Fatal("cannot find executable for convert: ", self.convert_path, " not found")
	}
	if !CheckFile(self.ffmpeg_path) {
		log.Fatal("connt find executable for ffmpegthumbnailer: ", self.ffmpeg_path, " not found")
	}
	if !CheckFile(self.sox_path) {
		log.Fatal("connt find executable for sox: ", self.sox_path, " not found")
	}
	if !CheckFile(self.placeholder) {
		log.Println("falling back to use default placeholder image")
		self.placeholder = "contrib/static/placeholder.png"
		if !CheckFile(self.placeholder) {
			log.Fatal("cannot find thumbnail placeholder file: ", self.placeholder, " not found")
		}
	}
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
	log.Println("made thumbnail for", fpath)
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
		log.Println("failed to determine size of thumbnail")
	}
	return info, err
}

func (self *articleStore) GenerateThumbnail(fname string) (info ThumbInfo, err error) {
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
	return info, err
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

// create a file for this article
func (self *articleStore) CreateFile(messageID string) io.WriteCloser {
	fname := self.GetFilename(messageID)
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
	return CheckFile(self.GetFilename(messageID))
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

// get article with headers only
func (self *articleStore) getMIMEHeader(messageID string) (hdr textproto.MIMEHeader) {
	if ValidMessageID(messageID) {
		fname := self.GetFilename(messageID)
		f, err := os.Open(fname)
		if f != nil {
			r := bufio.NewReader(f)
			var msg *mail.Message
			msg, err = readMIMEHeader(r)
			f.Close()
			hdr = textproto.MIMEHeader(msg.Header)
		}
		if err != nil {
			log.Println("failed to load article headers for", messageID, err)
		}
	}
	return hdr
}

func (self *articleStore) ProcessMessageBody(wr io.Writer, hdr textproto.MIMEHeader, body io.Reader) (err error) {
	err = read_message_body(body, hdr, self, wr, false, func(nntp NNTPMessage) {
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
	})
	return
}

func (self *articleStore) GetMessage(msgid string) (nntp NNTPMessage) {
	r, err := self.OpenMessage(msgid)
	if err == nil {
		defer r.Close()
		br := bufio.NewReader(r)
		msg, err := readMIMEHeader(br)
		if err == nil {
			chnl := make(chan NNTPMessage)
			hdr := textproto.MIMEHeader(msg.Header)
			err = read_message_body(msg.Body, hdr, nil, nil, true, func(nntp NNTPMessage) {
				c := chnl
				// inject pubkey for mod
				nntp.Headers().Set("X-PubKey-Ed25519", hdr.Get("X-PubKey-Ed25519"))
				c <- nntp
				close(c)
			})
			nntp = <-chnl
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
	if ok || content_type == "multipart/mixed" {
		partReader := multipart.NewReader(body, boundary)
		for {
			part, err := partReader.NextPart()
			if err == io.EOF {
				callback(nntp)
				return nil
			} else if err == nil {
				hdr := part.Header
				// get content type of part
				part_type := hdr.Get("Content-Type")
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
		pk := nntp.Pubkey()
		if pk == "" || sig == "" {
			log.Println("invalid sig or pubkey", sig, pk)
			nntp.Reset()
			return errors.New("invalid headers")
		}
		// process inner body
		// verify message
		err = verifyMessage(pk, sig, body, func(h map[string][]string, innerBody io.Reader) {
			// handle inner message
			err := read_message_body(innerBody, h, store, nil, true, callback)
			if err != nil {
				log.Println("error reading inner signed message", err)
			}
		})
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
