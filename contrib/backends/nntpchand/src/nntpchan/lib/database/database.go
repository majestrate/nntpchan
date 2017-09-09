package database

import (
	"errors"
	"nntpchan/lib/config"
	"nntpchan/lib/model"
	"strings"
)

//
type Database interface {
	ThreadByMessageID(msgid string) (*model.Thread, error)
	ThreadByHash(hash string) (*model.Thread, error)
	MessageIDByHash(hash string) (string, error)
	BoardPage(newsgroup string, pageno, perpage int) (*model.BoardPage, error)
	StorePost(post model.Post) error
	Init() error
}

// get new database connector from configuration
func NewDBFromConfig(c *config.DatabaseConfig) (db Database, err error) {
	dbtype := strings.ToLower(c.Type)
	if dbtype == "postgres" {
		db, err = createPostgresDatabase(c.Addr, c.Username, c.Password)
	} else {
		err = errors.New("no such database driver: " + c.Type)
	}
	return
}
