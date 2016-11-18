package message

import (
	"io"
)

// attachment in an nntp article
type Attachment struct {
	// mimetype
	Mime string
	// the filename
	FileName string
	// the fully decoded attachment body
	// must close when done
	Body io.ReadCloser
}
