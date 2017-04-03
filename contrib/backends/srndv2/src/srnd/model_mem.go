//
// model_mem.go
//
// models held in memory
//
package srnd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type catalogModel struct {
	frontend string
	prefix   string
	board    string
	threads  []CatalogItemModel
}

type catalogItemModel struct {
	page       int
	replycount int
	op         PostModel
}

func (self *catalogModel) Navbar() string {
	param := make(map[string]interface{})
	param["name"] = fmt.Sprintf("Catalog for %s", self.board)
	param["frontend"] = self.frontend
	var links []LinkModel
	links = append(links, linkModel{
		link: fmt.Sprintf("%sb/%s/", self.prefix, self.board),
		text: "Board index",
	})
	param["prefix"] = self.prefix
	param["links"] = links
	return template.renderTemplate("navbar.mustache", param)
}

func (self *catalogModel) MarshalJSON() (b []byte, err error) {
	return nil, nil
}

func (self *catalogModel) JSON() string {
	return "null"
}

func (self *catalogModel) Frontend() string {
	return self.frontend
}

func (self *catalogModel) Prefix() string {
	return self.prefix
}

func (self *catalogModel) Name() string {
	return self.board
}

func (self *catalogModel) Threads() []CatalogItemModel {
	return self.threads
}

func (self *catalogItemModel) OP() PostModel {
	return self.op
}

func (self *catalogItemModel) Page() string {
	return strconv.Itoa(self.page)
}

func (self *catalogItemModel) ReplyCount() string {
	return strconv.Itoa(self.replycount)
}

type boardModel struct {
	allowFiles bool
	frontend   string
	prefix     string
	board      string
	page       int
	pages      int
	threads    []ThreadModel
}

func (self *boardModel) MarshalJSON() (b []byte, err error) {
	j := make(map[string]interface{})
	j["posts"] = self.threads
	j["page"] = self.page
	j["name"] = self.board
	return json.Marshal(j)
}

func (self *boardModel) JSON() string {
	d, err := self.MarshalJSON()
	if err == nil {
		return string(d)
	} else {
		return "null"
	}
}

func (self *boardModel) SetAllowFiles(allow bool) {
	self.allowFiles = allow
}

func (self *boardModel) AllowFiles() bool {
	return self.allowFiles
}

func (self *boardModel) PutThread(th ThreadModel) {
	idx := -1
	for i, t := range self.threads {
		if th.OP().MessageID() == t.OP().MessageID() {
			idx = i
			break
		}
	}
	if idx != -1 {
		self.threads[idx] = th
	}
}

func (self *boardModel) Navbar() string {
	param := make(map[string]interface{})
	param["name"] = fmt.Sprintf("page %d for %s", self.page, self.board)
	param["frontend"] = self.frontend
	param["prefix"] = self.prefix
	param["links"] = self.PageList()
	return template.renderTemplate("navbar.mustache", param)
}

func (self *boardModel) Board() string {
	return self.board
}

func (self *boardModel) PageList() []LinkModel {
	var links []LinkModel
	for i := 0; i < self.pages; i++ {
		board := fmt.Sprintf("%sb/%s/%d/", self.prefix, self.board, i)
		if i == 0 {
			board = fmt.Sprintf("%sb/%s/", self.prefix, self.board)
		}
		links = append(links, linkModel{
			link: board,
			text: fmt.Sprintf("[ %d ]", i),
		})
	}
	return links
}

func (self *boardModel) UpdateThread(messageID string, db Database) {

	for _, th := range self.threads {
		if th.OP().MessageID() == messageID {
			// found it
			th.Update(db)
			break
		}
	}
}

func (self *boardModel) GetThread(messageID string) ThreadModel {
	for _, th := range self.threads {
		if th.OP().MessageID() == messageID {
			return th
		}
	}
	return nil
}

func (self *boardModel) HasThread(messageID string) bool {
	return self.GetThread(messageID) != nil
}

func (self *boardModel) Frontend() string {
	return self.frontend
}

func (self *boardModel) Prefix() string {
	return self.prefix
}

func (self *boardModel) Name() string {
	return self.board
}

func (self *boardModel) Threads() []ThreadModel {
	return self.threads
}

// refetch all threads on this page
func (self *boardModel) Update(db Database) {
	// ignore error
	perpage, _ := db.GetThreadsPerPage(self.board)
	// refetch all on this page
	model := db.GetGroupForPage(self.prefix, self.frontend, self.board, self.page, int(perpage))
	for _, th := range model.Threads() {
		// XXX: do we really need to update it again?
		th.Update(db)
	}
	self.threads = model.Threads()
}

type post struct {
	truncated        bool
	prefix           string
	board            string
	PostName         string
	PostSubject      string
	PostMessage      string
	message_rendered string
	Message_id       string
	MessagePath      string
	addr             string
	Newsgroup        string
	op               bool
	Posted           int64
	Parent           string
	sage             bool
	Key              string
	Files            []AttachmentModel
	HashLong         string
	HashShort        string
	URL              string
	Tripcode         string
	BodyMarkup       string
	PostMarkup       string
	PostPrefix       string
	index            int
	Type             string
	nntp_id          int
}

func (self *post) NNTPID() int {
	return self.nntp_id
}

func (self *post) Index() int {
	return self.index + 1
}

func (self *post) NumImages() int {
	return len(self.Files)
}

func (self *post) RepresentativeThumb() string {
	if len(self.Attachments()) > 0 {
		return self.Attachments()[0].Thumbnail()
	}
	//TODO don't hard-code this
	return self.prefix + "static/placeholder.png"
}

func (self *post) MarshalJSON() (b []byte, err error) {
	// compute on fly
	// TODO: don't do this
	self.HashLong = self.PostHash()
	self.HashShort = self.ShortHash()
	if len(self.Key) > 0 {
		self.Tripcode = makeTripcode(self.Key)
	}
	if len(self.PostMarkup) > 0 {
		self.PostMarkup = self.RenderPost()
	}
	self.PostPrefix = self.Prefix()
	// for liveui
	self.Type = "Post"
	self.Newsgroup = self.board
	self.URL = self.PostURL()
	return json.Marshal(*self)
}

func (self *post) JSON() string {
	d, err := self.MarshalJSON()
	if err == nil {
		return string(d)
	} else {
		return "null"
	}
}

type attachment struct {
	prefix      string
	Path        string
	Name        string
	ThumbWidth  int
	ThumbHeight int
}

func (self *attachment) MarshalJSON() (b []byte, err error) {
	return json.Marshal(*self)
}

func (self *attachment) JSON() string {
	d, err := self.MarshalJSON()
	if err == nil {
		return string(d)
	} else {
		return "null"
	}
}

func (self *attachment) Hash() string {
	return strings.Split(self.Path, ".")[0]
}

func (self *attachment) ThumbInfo() ThumbInfo {
	return ThumbInfo{
		Width:  self.ThumbWidth,
		Height: self.ThumbHeight,
	}
}

func (self *attachment) Prefix() string {
	return self.prefix
}

func (self *attachment) Thumbnail() string {
	return self.prefix + "thm/" + self.Path + ".jpg"
}

func (self *attachment) Source() string {
	return self.prefix + "img/" + self.Path
}

func (self *attachment) Filename() string {
	return self.Name
}

func PostModelFromMessage(parent, prefix string, nntp NNTPMessage) PostModel {
	p := new(post)
	p.PostName = nntp.Name()
	p.PostSubject = nntp.Subject()
	p.PostMessage = nntp.Message()
	p.MessagePath = nntp.Path()
	p.Message_id = nntp.MessageID()
	p.board = nntp.Newsgroup()
	p.Posted = nntp.Posted()
	p.op = nntp.OP()
	p.prefix = prefix
	p.Parent = parent
	p.addr = nntp.Addr()
	p.sage = nntp.Sage()
	p.Key = nntp.Pubkey()
	for _, att := range nntp.Attachments() {
		p.Files = append(p.Files, att.ToModel(prefix))
	}
	return p
}

func (self *post) ReferenceHash() string {
	ref := self.Reference()
	if len(ref) > 0 {
		return HashMessageID(self.Reference())
	}
	return self.PostHash()
}
func (self *post) NumAttachments() int {
	return len(self.Files)
}

func (self *post) RenderTruncatedBody() string {
	return self.Truncate().RenderBody()
}

func (self *post) Reference() string {
	return self.Parent
}

func (self *post) ShortHash() string {
	return ShortHashMessageID(self.MessageID())
}

func (self *post) Pubkey() string {
	if len(self.Key) > 0 {
		return fmt.Sprintf("<label title=\"%s\">%s</label>", self.Key, makeTripcode(self.Key))
	}
	return ""
}

func (self *post) Sage() bool {
	return self.sage
}

func (self *post) CSSClass() string {
	if self.OP() {
		return "post op"
	} else {
		return "post reply"
	}
}

func (self *post) OP() bool {
	return self.Parent == self.Message_id || len(self.Parent) == 0
}

func (self *post) Date() string {
	return time.Unix(self.Posted, 0).Format(i18nProvider.Format("full_date_format"))
}

func (self *post) DateRFC() string {
	return time.Unix(self.Posted, 0).Format(time.RFC3339)
}

func (self *post) TemplateDir() string {
	return filepath.Join("contrib", "templates", "default")
}

func (self *post) MessageID() string {
	return self.Message_id
}

func (self *post) Frontend() string {
	idx := strings.LastIndex(self.MessagePath, "!")
	if idx == -1 {
		return self.MessagePath
	}
	return self.MessagePath[idx+1:]
}

func (self *post) Board() string {
	return self.board
}

func (self *post) PostHash() string {
	return HashMessageID(self.Message_id)
}

func (self *post) Name() string {
	return self.PostName
}

func (self *post) Subject() string {
	return self.PostSubject
}

func (self *post) Attachments() []AttachmentModel {
	return self.Files
}

func (self *post) PostURL() string {
	return fmt.Sprintf("%st/%s/#%s", self.Prefix(), HashMessageID(self.Parent), self.PostHash())
}

func (self *post) Prefix() string {
	if len(self.prefix) == 0 {
		// fall back if not set
		return "/"
	}
	return self.prefix
}

func (self *post) IsClearnet() bool {
	return len(self.addr) == encAddrLen()
}

func (self *post) IsI2P() bool {
	return len(self.addr) == i2pDestHashLen()
}

func (self *post) IsTor() bool {
	return len(self.addr) == 0
}

func (self *post) SetIndex(idx int) {
	self.index = idx
}

func (self *post) RenderPost() string {
	param := make(map[string]interface{})
	param["post"] = self
	return template.renderTemplate("post.mustache", param)
}

func (self *post) RenderTruncatedPost() string {
	return self.Truncate().RenderPost()
}

func (self *post) IsTruncated() bool {
	return self.truncated
}

func (self *post) Truncate() PostModel {
	message := self.PostMessage
	subject := self.PostSubject
	name := self.PostName
	if len(message) > 500 {
		message = message[:500] + "\n...\n[Post Truncated]\n"
	}
	if len(subject) > 100 {
		subject = subject[:100] + "..."
	}
	if len(name) > 100 {
		name = name[:100] + "..."
	}

	return &post{
		truncated:   true,
		prefix:      self.prefix,
		board:       self.board,
		PostName:    name,
		PostSubject: subject,
		PostMessage: message,
		Message_id:  self.Message_id,
		MessagePath: self.MessagePath,
		addr:        self.addr,
		op:          self.op,
		Posted:      self.Posted,
		Parent:      self.Parent,
		sage:        self.sage,
		Key:         self.Key,
		// TODO: copy?
		Files: self.Files,
	}
}

func (self *post) RenderShortBody() string {
	return MEMEPosting(self.PostMessage, self.Prefix())
}

func (self *post) RenderBodyPre() string {
	return self.PostMessage
}

func (self *post) RenderBody() string {
	// :^)
	if len(self.message_rendered) == 0 {
		self.message_rendered = MEMEPosting(self.PostMessage, self.Prefix())
	}
	return self.message_rendered
}

type thread struct {
	allowFiles          bool
	prefix              string
	links               []LinkModel
	Posts               []PostModel
	dirty               bool
	truncatedPostCount  int
	truncatedImageCount int
}

func (self *thread) MarshalJSON() (b []byte, err error) {
	posts := []PostModel{self.OP()}
	posts = append(posts, self.Replies()...)
	return json.Marshal(posts)
}

func (self *thread) JSON() string {
	d, err := self.MarshalJSON()
	if err == nil {
		return string(d)
	} else {
		return "null"
	}
}

func (self *thread) IsDirty() bool {
	return self.dirty
}

func (self *thread) MarkDirty() {
	self.dirty = true
}

func (self *thread) Prefix() string {
	return self.prefix
}

func (self *thread) Navbar() string {
	param := make(map[string]interface{})
	param["name"] = fmt.Sprintf("Thread %s", self.Posts[0].ShortHash())
	param["frontend"] = self.Board()
	param["links"] = self.links
	param["prefix"] = self.prefix
	return template.renderTemplate("navbar.mustache", param)
}

func (self *thread) Board() string {
	return self.Posts[0].Board()
}

func (self *thread) BoardURL() string {
	return fmt.Sprintf("%sb/%s/", self.Prefix(), self.Board())
}

func (self *thread) PostCount() int {
	return len(self.Posts)
}

func (self *thread) ImageCount() (count int) {
	for _, p := range self.Posts {
		count += p.NumAttachments()
	}
	return
}

// get our default template dir
func defaultTemplateDir() string {
	return filepath.Join("contrib", "templates", "default")
}

func createThreadModel(posts ...PostModel) ThreadModel {
	op := posts[0]
	group := op.Board()
	prefix := op.Prefix()
	return &thread{
		dirty:  true,
		prefix: prefix,
		Posts:  posts,
		links: []LinkModel{
			linkModel{
				text: group,
				link: fmt.Sprintf("%sb/%s/", prefix, group),
			},
		},
	}
}

func (self *thread) OP() PostModel {
	return self.Posts[0]
}

func (self *thread) Replies() []PostModel {
	if len(self.Posts) > 1 {
		var replies []PostModel
		// inject post index
		for idx, post := range self.Posts[1:] {
			if post != nil {
				post.SetIndex(idx + 1)
				replies = append(replies, post)
			}
		}
		return replies
	}
	return []PostModel{}
}

func (self *thread) AllowFiles() bool {
	return self.allowFiles
}

func (self *thread) SetAllowFiles(allow bool) {
	self.allowFiles = allow
}

func (self *thread) Truncate() ThreadModel {
	trunc := 5
	if len(self.Posts) > trunc {
		t := &thread{
			allowFiles: self.allowFiles,
			links:      self.links,
			Posts:      append([]PostModel{self.Posts[0]}, self.Posts[len(self.Posts)-trunc:]...),
			prefix:     self.prefix,
			dirty:      false,
		}
		imgs := 0
		for _, p := range t.Posts {
			imgs += p.NumAttachments()
		}
		t.truncatedPostCount = len(self.Posts) - trunc
		t.truncatedImageCount = self.ImageCount() - imgs
		return t
	}
	return self
}

func (self *thread) MissingPostCount() int {
	return self.truncatedPostCount
}

func (self *thread) MissingImageCount() int {
	return self.truncatedImageCount
}

func (self *thread) HasOmittedReplies() bool {
	return self.truncatedPostCount > 0
}

func (self *thread) HasOmittedImages() bool {
	return self.truncatedImageCount > 0
}

func (self *thread) Update(db Database) {
	root := self.Posts[0].MessageID()
	self.Posts = append([]PostModel{self.Posts[0]}, db.GetThreadReplyPostModels(self.prefix, root, 0, 0)...)
	self.dirty = false
}

type linkModel struct {
	text string
	link string
}

func (self linkModel) LinkURL() string {
	return self.link
}

func (self linkModel) Text() string {
	return self.text
}
