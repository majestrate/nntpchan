package model

import (
	"time"
)

type ArticleHeader map[string][]string

// a ( MessageID , newsgroup ) tuple
type ArticleEntry [2]string

func (self ArticleEntry) Newsgroup() string {
	return self[1]
}

func (self ArticleEntry) MessageID() string {
	return self[0]
}

// a ( time point, post count ) tuple
type PostEntry [2]int64

func (self PostEntry) Time() time.Time {
	return time.Unix(self[0], 0)
}

func (self PostEntry) Count() int64 {
	return self[1]
}
