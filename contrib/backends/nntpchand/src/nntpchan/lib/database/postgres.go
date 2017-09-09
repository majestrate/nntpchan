package database

import (
	"database/sql"
	_ "github.com/lib/pq"
	"nntpchan/lib/model"
)

type PostgresDB struct {
	conn *sql.DB
}

func (db *PostgresDB) ThreadByMessageID(msgid string) (thread *model.Thread, err error) {

	return
}

func (db *PostgresDB) ThreadByHash(hash string) (thread *model.Thread, err error) {
	return
}

func (db *PostgresDB) MessageIDByHash(hash string) (msgid string, err error) {

	return
}

func (db *PostgresDB) BoardPage(newsgroup string, pageno, perpage int) (page *model.BoardPage, err error) {

	return
}

func (db *PostgresDB) StorePost(post model.Post) (err error) {
	return
}

func (db *PostgresDB) Init() (err error) {

	return
}

func createPostgresDatabase(addr, user, passwd string) (p *PostgresDB, err error) {
	p = new(PostgresDB)
	var authstring string
	if len(addr) > 0 {
		authstring += " host=" + addr
	}
	if len(user) > 0 {
		authstring += " username=" + user
	}
	if len(passwd) > 0 {
		authstring += " password=" + passwd
	}
	p.conn, err = sql.Open("postgres", authstring)
	if err != nil {
		p = nil
	}
	return
}
