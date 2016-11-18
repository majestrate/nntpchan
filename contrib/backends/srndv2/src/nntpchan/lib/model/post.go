package model

type Post struct {
	Board            string
	PostName         string
	PostSubject      string
	PostMessage      string
	message_rendered string
	Message_id       string
	MessagePath      string
	Addr             string
	OP               bool
	Posted           int64
	Parent           string
	Sage             bool
	Key              string
	Files            []*Attachment
	HashLong         string
	HashShort        string
	URL              string
	Tripcode         string
	BodyMarkup       string
	PostMarkup       string
	PostPrefix       string
	index            int
}

// ( message-id, references, newsgroup )
type PostReference [3]string
