package model

type Attachment struct {
	Path string
	Name string
	Mime string
	Hash string
	// only filled for api
	Body string
}
