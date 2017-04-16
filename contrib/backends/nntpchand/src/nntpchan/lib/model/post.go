package model

import (
	"time"
)

type Tripcode string

type Post struct {
	MessageID   string
	Newsgroup   string
	Attachments []Attachment
	Subject     string
	Posted      time.Time
	PostedAt    uint64
	Name        string
	Tripcode    Tripcode
}

// ( message-id, references, newsgroup )
type PostReference [3]string

func (r PostReference) MessageID() string {
	return r[0]
}

func (r PostReference) References() string {
	return r[1]
}
func (r PostReference) Newsgroup() string {
	return r[2]
}
