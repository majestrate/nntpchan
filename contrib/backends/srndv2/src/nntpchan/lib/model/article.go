package model

type Article struct {
	Subject     string
	Name        string
	Header      map[string][]string
	Text        string
	Attachments []Attachment
	MessageID   string
	Newsgroup   string
	Reference   string
	Path        string
	Posted      int64
	Addr        string
}
