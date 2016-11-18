package model

type Thread struct {
	Root    *Post
	Replies []*Post
}
