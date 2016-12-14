package nntp

import (
	"nntpchan/lib/nntp/message"
	"io"
)

// defines interface for filtering an nntp article
// filters can (and does) modify the article it operates on
type ArticleFilter interface {
	// filter the article header
	// returns the modified Header and an error if one occurs
	FilterHeader(hdr message.Header) (message.Header, error)

	// reads the article's body and write the filtered version to an io.Writer
	// returns the number of bytes written to the io.Writer, true if the body was
	// modifed (or false if body is unchanged) and an error if one occurs
	FilterAndWriteBody(body io.Reader, wr io.Writer) (int64, bool, error)
}
