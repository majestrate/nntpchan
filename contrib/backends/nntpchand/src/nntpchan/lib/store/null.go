package store

import (
	"io"
	"nntpchan/lib/util"
	"os"
)

type nullStore struct{}

func (n *nullStore) discard(r io.Reader) (s string, err error) {
	_, err = io.Copy(util.Discard, r)
	s = "/dev/null"
	return
}

func (n *nullStore) HasArticle(msgid string) error {
	return ErrNoSuchArticle
}

func (n *nullStore) StoreAttachment(r io.Reader, filename string) (string, error) {
	return n.discard(r)
}

func (n *nullStore) StoreArticle(r io.Reader, msgid, newsgroup string) (string, error) {
	return n.discard(r)
}

func (n *nullStore) DeleteArticle(msgid string) (err error) {
	return
}

func (n *nullStore) Ensure() (err error) {
	return
}

func (n *nullStore) ForEachInGroup(newsgroup string, chnl chan string) {
	return
}

func (n *nullStore) OpenArticle(msgid string) (r *os.File, err error) {
	err = ErrNoSuchArticle
	return
}

func (n *nullStore) HasNewsgroup(newsgroup string) (has bool, err error) {
	has = true
	return
}

func (n *nullStore) GetAllNewsgroups() (list []string, err error) {
	return
}

func (n *nullStore) GetWatermark(newsgroup string) (hi, lo uint64, err error) {
	return
}

// create a storage backend that does nothing
func NewNullStorage() Storage {
	return &nullStore{}
}
