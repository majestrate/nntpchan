package database

import (
	"nntpchan/lib/model"
)

type PostgresDB struct {
}

func (db *PostgresDB) ThreadByMessageID(msgid string) (thread *model.Thread, err error) {

	return
}

func (db *PostgresDB) ThreadByHash(hash string) (thread *model.Thread, err error) {

	return
}

func (db *PostgresDB) BoardPage(newsgroup string, pageno, perpage int) (page *model.BoardPage, err error) {

	return
}

func createPostgresDatabase(addr, user, passwd string) (p *PostgresDB, err error) {

	return
}
