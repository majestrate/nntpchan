//
// model.go
// template model interfaces
//
package srnd

import (
	"time"
)

// base model type
type BaseModel interface {

	// set sfw flag
	MarkSFW(sfw bool)

	// site url prefix
	Prefix() string

	// impelements json.Marshaller
	MarshalJSON() ([]byte, error)

	// to json string
	JSON() string

	// inject I18N
	I18N(i *I18N)
}

type ThumbInfo struct {
	Width  int
	Height int
}

// for attachments
type AttachmentModel interface {
	BaseModel

	Thumbnail() string
	Source() string
	Filename() string
	Hash() string
	ThumbInfo() ThumbInfo
}

// for individual posts
type PostModel interface {
	BaseModel

	Brief() string
	CSSClass() string
	FrontendPubkey() string
	MessageID() string
	PostHash() string
	ShortHash() string
	PostURL() string
	Frontend() string
	Subject() string
	Name() string
	Date() string
	OP() bool
	Attachments() []AttachmentModel
	NumAttachments() int
	Board() string
	Sage() bool
	Pubkey() string
	PubkeyHex() string
	Reference() string
	ReferenceHash() string

	RenderBody() string
	RenderPost() string
	RenderBodyPre() string

	// replaces Truncate().RenderBody()
	RenderTruncatedBody() string

	// replaces Truncate().RenderPost()
	RenderTruncatedPost() string

	// returns true if this post was truncated
	IsTruncated() bool

	// return true if this post is a mod message
	IsCtl() bool

	IsI2P() bool
	IsTor() bool
	IsClearnet() bool

	// deprecated
	// truncate body to a certain size
	// return copy
	Truncate() PostModel

	// what is our position in this thread?
	// 0 for OP, nonzero for reply
	Index() int
	// set post index
	SetIndex(idx int)

	// nntp id number
	NNTPID() int
}

// interface for models that have a navbar
type NavbarModel interface {
	Navbar() string
}

// for threads
type ThreadModel interface {
	BaseModel
	NavbarModel

	SetAllowFiles(allow bool)
	AllowFiles() bool
	OP() PostModel
	Replies() []PostModel
	Board() string
	BoardURL() string
	// return a short version of the thread
	// does not include all replies
	Truncate() ThreadModel

	// number of posts in this thread
	PostCount() int
	// number of images in this thread
	ImageCount() int
	// number of posts excluded during truncation
	// returns 0 if not truncated
	MissingPostCount() int
	// number of images excluded during truncation
	// returns 0 if not truncated
	MissingImageCount() int
	// returns true if this thread has truncated replies
	HasOmittedReplies() bool
	// returns true if this thread has truncated images
	HasOmittedImages() bool

	// update the thread's replies
	Update(db Database)
	// is this thread dirty and needing updating?
	IsDirty() bool
	// mark thread as dirty
	MarkDirty()
	// is the threa bumplocked?
	BumpLock() bool
}

// board interface
// for 1 page on a board
type BoardModel interface {
	BaseModel
	NavbarModel

	Frontend() string
	Name() string
	Threads() []ThreadModel

	AllowFiles() bool
	SetAllowFiles(files bool)

	// JUST update this thread
	// if we don't have it already loaded do nothing
	UpdateThread(message_id string, db Database)

	// get a thread model with this id
	// returns nil if we don't have it
	GetThread(message_id string) ThreadModel

	// put a thread back after updating externally
	PutThread(th ThreadModel)

	// deprecated, use GetThread
	HasThread(message_id string) bool

	// update the board's contents
	Update(db Database)
}

type CatalogModel interface {
	BaseModel
	NavbarModel

	Frontend() string
	Name() string
	Threads() []CatalogItemModel
}

type CatalogItemModel interface {
	OP() PostModel
	ReplyCount() string
	Page() string
	MarkSFW(sfw bool)
}

type LinkModel interface {
	Text() string
	LinkURL() string
}

// newsgroup model
// every page on a newsgroup
type GroupModel []BoardModel

// TODO: optimize using 1 query?
// update every page
func (self *GroupModel) UpdateAll(db Database) {
	m := *self
	for _, page := range m {
		page.Update(db)
	}
}

// update a certain page
// does nothing if out of bounds
func (self *GroupModel) Update(page int, db Database) {
	m := *self
	if len(m) > page {
		m[page].Update(db)
	}
}

type boardPageRow struct {
	Board string
	Hour  int64
	Day   int64
	All   int64
	Hi    int64
	Lo    int64
}

type boardPageRows []boardPageRow

func (self boardPageRows) Len() int {
	return len(self)
}

func (self boardPageRows) Less(i, j int) bool {
	i_val := self[i]
	j_val := self[j]
	return (i_val.Day + i_val.Hour*24) > (j_val.Day + j_val.Hour*24)
}

func (self boardPageRows) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type postsGraphRow struct {
	day time.Time
	Num int64
	mag int64
}

func (p *postsGraphRow) GraphRune(r string) (s string) {
	var num int64
	if p.mag > 0 {
		num = p.Num / p.mag
	} else {
		num = p.Num
	}
	for num > 0 {
		s += r
		num--
	}
	return
}

func (p postsGraphRow) Date() (s string) {
	return p.day.Format(I18nProvider.Format("month_date_format"))
}

func (p postsGraphRow) Day() (s string) {
	return p.day.Format(I18nProvider.Format("day_date_format"))
}

func (p postsGraphRow) RegularGraph() (s string) {
	return p.GraphRune("=")
}

// :0========3 overcock :3 graph of data
func (p postsGraphRow) OvercockGraph() (s string) {
	var num int64
	if p.mag > 0 {
		num = p.Num / p.mag
	} else {
		num = p.Num
	}
	if num > 0 {
		s = ":0"
		num -= 1
		for num > 0 {
			s += "="
			num--
		}
		s += "3"
	} else {
		s = ":3"
	}
	return
}

type postsGraph []postsGraphRow

func (self postsGraph) Len() int {
	return len(self)
}

func (self postsGraph) Less(i, j int) bool {
	i_val := self[i]
	j_val := self[j]
	return i_val.day.Unix() > j_val.day.Unix()
}

func (self postsGraph) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self postsGraph) Scale() (graph postsGraph) {
	// find max
	max := int64(0)
	for _, p := range self {
		if p.Num > max {
			max = p.Num
		}
	}
	mag := max / 25
	for _, p := range self {
		p.mag = mag
		graph = append(graph, p)
	}
	return
}

type overviewModel []PostModel
