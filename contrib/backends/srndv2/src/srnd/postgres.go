//
// postgres db backend
//
package srnd

/**
 * TODO:
 *  ~ caching of board settings
 *  ~ caching of encrypted address info
 *  ~ multithreading check
 *  ~ checking for duplicate articles
 */

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// postgres database driver implementation
type PostgresDatabase struct {
	conn   *sql.DB
	db_str string
	stmt   map[string]string
}

// create postgres database driver
func NewPostgresDatabase(host, port, user, password string) Database {
	db := new(PostgresDatabase)
	var err error
	if len(user) > 0 {
		if len(password) > 0 {
			db.db_str = fmt.Sprintf("user=%s password='%s' host=%s port=%s client_encoding='UTF8'", user, password, host, port)
		} else {
			db.db_str = fmt.Sprintf("user=%s host=%s port=%s client_encoding='UTF8'", user, host, port)
		}
	} else {
		if len(port) > 0 {
			db.db_str = fmt.Sprintf("host=%s port=%s client_encoding='UTF8'", host, port)
		} else {
			db.db_str = fmt.Sprintf("host=%s client_encoding='UTF8'", host)
		}
	}
	log.Println("Connecting to postgres...")
	db.conn, err = sql.Open("postgres", db.db_str)
	if err != nil {
		log.Fatalf("can`not open connection to db: %s", err)
	}
	db.SetConnectionLifetime(30)
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(10)
	return db
}

func (db *PostgresDatabase) SetConnectionLifetime(seconds int) {
	db.conn.SetConnMaxLifetime(time.Second * time.Duration(seconds))
}

func (db *PostgresDatabase) SetMaxOpenConns(n int) {
	db.conn.SetMaxOpenConns(n)
}

func (db *PostgresDatabase) SetMaxIdleConns(n int) {
	db.conn.SetMaxIdleConns(n)
}

// finalize all transactions
// close database connections
func (self *PostgresDatabase) Close() {

	if self.conn != nil {
		self.conn.Close()
		self.conn = nil
	}
}

const NewsgroupBanned = "NewsgroupBanned"
const ArticleBanned = "ArticleBanned"
const GetAllNewsgroups = "GetAllNewsgroups"
const GetPostsInGroup = "GetPostsInGroup"
const GetPostModel = "GetPostModel"
const GetArticlePubkey = "GetArticlePubkey"
const GetThreadModel = "GetThreadModel"
const GetThreadModelPubkeys = "GetThreadModelPubkeys"
const GetThreadModelAttachments = "GetThreadModelAttachments"
const DeleteArticle_1 = "DeleteArticle_1"
const DeleteArticle_2 = "DeleteArticle_2"
const DeleteArticle_3 = "DeleteArticle_3"
const DeleteArticle_4 = "DeleteArticle_4"
const DeleteArticle_5 = "DeleteArticle_5"
const DeleteArticleV8 = "DeleteArticleV8"
const DeleteThread = "DeleteThread"
const DeleteThreadV8 = "DeleteThreadV8"
const GetThreadReplyPostModels_1 = "GetThreadReplyPostModels_1"
const GetThreadReplyPostModels_2 = "GetThreadReplyPostModels_2"
const GetThreadReplies_1 = "GetThreadReplies_1"
const GetThreadReplies_2 = "GetThreadReplies_2"
const GetGroupThreads = "GetGroupThreads"
const GetLastBumpedThreadsPaginated_1 = "GetLastBumpedThreadsPaginated_1"
const GetLastBumpedThreadsPaginated_2 = "GetLastBumpedThreadsPaginated_2"
const HasNewsgroup = "HasNewsgroup"
const HasArticle = "HasArticle"
const HasArticleLocal = "HasArticleLocal"
const GetPostAttachments = "GetPostAttachments"
const GetThreadAttachments = "GetThreadAttachments"
const GetPostAttachmentModels = "GetPostAttachmentModels"
const RegisterArticle_GetLastBump = "RegisterArticle_GetLastBump"
const RegisterArticle_1 = "RegisterArticle_1"
const RegisterArticle_2 = "RegisterArticle_2"
const RegisterArticle_3 = "RegisterArticle_3"
const RegisterArticle_4 = "RegisterArticle_4"
const RegisterArticle_5 = "RegisterArticle_5"
const RegisterArticle_6 = "RegisterArticle_6"
const RegisterArticle_7 = "RegisterArticle_7"
const RegisterArticle_8 = "RegisterArticle_8"
const GetMessageIDByHeader = "GetMessageIDByHeader"
const RegisterSigned = "RegisterSigned"
const GetAllArticlesInGroup = "GetAllArticlesInGroup"
const GetAllArticles = "GetAllArticles"
const GetMessageIDByHash = "GetMessageIDByHash"
const CheckEncIPBanned = "CheckEncIPBanned"
const GetFirstAndLastForGroup = "GetFirstAndLastForGroup"
const GetMessageIDForNNTPID = "GetMessageIDForNNTPID"
const GetNNTPIDForMessageID = "GetNNTPIDForMessageID"
const IsExpired = "IsExpired"
const GetLastDaysPostsForGroup = "GetLastDaysPostsForGroup"
const GetLastDaysPosts = "GetLastDaysPosts"
const GetLastPostedPostModels = "GetLastPostedPostModels"
const GetMonthlyPostHistory = "GetMonthlyPostHistory"
const CheckNNTPLogin = "CheckNNTPLogin"
const CheckNNTPUserExists = "CheckNNTPUserExists"
const GetHeadersForMessage = "GetHeadersForMessage"
const CountAllArticlesInGroup = "CountAllArticlesInGroup"
const GetMessageIDByCIDR = "GetMessageIDByCIDR"
const GetMessageIDByEncryptedIP = "GetMessageIDByEncryptedIP"
const GetPostsBefore = "GetPostsBefore"
const SearchQuery_1 = "SearchQuery_1"
const SearchQuery_2 = "SearchQuery_2"
const SearchByHash_1 = "SearchByHash_1"
const SearchByHash_2 = "SearchByHash_2"
const GetNNTPPostsInGroup = "GetNNTPPostsInGroup"
const GetCitesByPostHashLike = "GetCitesByPostHashLike"
const GetYearlyPostHistory = "GetYearlyPostHistory"
const GetNewsgroupList = "GetNewsgroupList"
const CountUkko = "CountUkko"
const GetNewsgroupStats = "GetNewsgroupStats"
const RemoveArticle = "RemoveArticle"

func (self *PostgresDatabase) prepareStatements() {
	self.stmt = map[string]string{
		GetNewsgroupStats:               "SELECT COUNT(message_id), newsgroup FROM articleposts WHERE time_posted > (EXTRACT(epoch FROM NOW()) - (24*3600)) GROUP BY newsgroup",
		NewsgroupBanned:                 "SELECT 1 FROM BannedGroups WHERE newsgroup = $1",
		ArticleBanned:                   "SELECT 1 FROM BannedArticles WHERE message_id = $1",
		GetAllNewsgroups:                "SELECT name FROM Newsgroups WHERE name NOT IN ( SELECT newsgroup FROM BannedGroups )",
		GetPostsInGroup:                 "SELECT newsgroup, message_id, ref_id, name, subject, path, time_posted, message, addr, frontendpubkey FROM ArticlePosts WHERE newsgroup = $1 ORDER BY time_posted",
		GetPostModel:                    "SELECT newsgroup, message_id, ref_id, name, subject, path, time_posted, message, addr, frontendpubkey FROM ArticlePosts WHERE message_id = $1 LIMIT 1",
		GetArticlePubkey:                "SELECT pubkey FROM ArticleKeys WHERE message_id = $1",
		GetThreadModel:                  "SELECT ArticlePosts.newsgroup, ArticlePosts.message_id, ArticlePosts.name, ArticlePosts.subject, ArticlePosts.time_posted, ArticlePosts.message, ArticlePosts.addr, ArticlePosts.frontendpubkey FROM ArticlePosts WHERE ArticlePosts.message_id = $1 OR ArticlePosts.ref_id = $1 ORDER BY ArticlePosts.time_posted",
		GetThreadModelPubkeys:           "SELECT pubkey, message_id from ArticleKeys WHERE message_id IN ( SELECT message_id FROM ArticlePosts WHERE ref_id = $1 OR message_id = $1 )",
		GetThreadModelAttachments:       "SELECT filename, filepath, message_id from ArticleAttachments WHERE message_id IN ( SELECT message_id FROM ArticlePosts WHERE ref_id = $1 OR message_id = $1 )",
		DeleteArticle_1:                 "DELETE FROM NNTPHeaders WHERE header_article_message_id = $1",
		DeleteArticle_2:                 "DELETE FROM ArticleNumbers WHERE message_id = $1",
		DeleteArticle_3:                 "DELETE FROM ArticlePosts WHERE message_id = $1",
		DeleteArticle_4:                 "DELETE FROM ArticleKeys WHERE message_id = $1",
		DeleteArticle_5:                 "DELETE FROM ArticleAttachments WHERE message_id = $1",
		DeleteThread:                    "DELETE FROM ArticleThreads WHERE root_message_id = $1",
		DeleteArticleV8:                 "DELETE FROM ArticlePosts WHERE message_id = $1",
		DeleteThreadV8:                  "DELETE FROM ArticlePosts WHERE ref_id = $1 OR message_id = $1",
		GetThreadReplyPostModels_1:      "SELECT newsgroup, message_id, ref_id, name, subject, path, time_posted, message, addr, frontendpubkey FROM ArticlePosts WHERE message_id IN ( SELECT message_id FROM ArticlePosts WHERE ref_id = $1 ORDER BY time_posted DESC LIMIT $2 ) ORDER BY time_posted ASC",
		GetThreadReplyPostModels_2:      "SELECT newsgroup, message_id, ref_id, name, subject, path, time_posted, message, addr, frontendpubkey FROM ArticlePosts WHERE message_id IN ( SELECT message_id FROM ArticlePosts WHERE ref_id = $1 ) ORDER BY time_posted ASC",
		GetThreadReplies_1:              "SELECT message_id FROM ArticlePosts WHERE message_id IN ( SELECT message_id FROM ArticlePosts WHERE ref_id = $1 ORDER BY time_posted DESC LIMIT $2 ) ORDER BY time_posted ASC",
		GetThreadReplies_2:              "SELECT message_id FROM ArticlePosts WHERE message_id IN ( SELECT message_id FROM ArticlePosts WHERE ref_id = $1 ) ORDER BY time_posted ASC",
		GetGroupThreads:                 "SELECT message_id FROM ArticlePosts WHERE newsgroup = $1 AND ref_id = '' ",
		GetLastBumpedThreadsPaginated_1: "SELECT root_message_id, newsgroup FROM ArticleThreads WHERE newsgroup = $1 ORDER BY last_bump DESC LIMIT $2",
		GetLastBumpedThreadsPaginated_2: "SELECT root_message_id, newsgroup FROM ArticleThreads WHERE newsgroup != 'ctl' ORDER BY last_bump DESC LIMIT $1",
		HasNewsgroup:                    "SELECT 1 FROM Newsgroups WHERE name = $1",
		HasArticle:                      "SELECT 1 FROM Articles WHERE message_id = $1",
		HasArticleLocal:                 "SELECT 1 FROM ArticlePosts WHERE message_id = $1",
		GetPostAttachments:              "SELECT filepath FROM ArticleAttachments WHERE message_id = $1",
		GetThreadAttachments:            "SELECT filepath FROM ArticleAttachments WHERE message_id IN ( SELECT message_id FROM ArticlePosts WHERE ref_id = $1 OR message_id = $1)",
		GetPostAttachmentModels:         "SELECT filepath, filename FROM ArticleAttachments WHERE message_id = $1",
		RegisterArticle_GetLastBump:     "SELECT last_bump FROM ArticleThreads WHERE root_message_id = $1",
		RegisterArticle_1:               "INSERT INTO Articles (message_id, message_id_hash, message_newsgroup, time_obtained, message_ref_id) VALUES($1, $2, $3, $4, $5)",
		RegisterArticle_2:               "UPDATE Newsgroups SET last_post = $1 WHERE name = $2",
		RegisterArticle_3:               "INSERT INTO ArticlePosts(newsgroup, message_id, ref_id, name, subject, path, time_posted, message, addr, frontendpubkey) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		RegisterArticle_4:               "INSERT INTO ArticleThreads(root_message_id, last_bump, last_post, newsgroup) VALUES($1, $2, $2, $3)",
		RegisterArticle_5:               "SELECT COUNT(*) FROM ArticlePosts WHERE ref_id = $1",
		RegisterArticle_6:               "UPDATE ArticleThreads SET last_bump = $2 WHERE root_message_id = $1",
		RegisterArticle_7:               "UPDATE ArticleThreads SET last_post = $2 WHERE root_message_id = $1",
		RegisterArticle_8:               "INSERT INTO ArticleAttachments(message_id, sha_hash, filename, filepath) VALUES($1, $2, $3, $4)",
		GetMessageIDByHeader:            "SELECT header_article_message_id FROM NNTPHeaders WHERE header_name = $1 AND header_value = $2",
		RegisterSigned:                  "INSERT INTO ArticleKeys(message_id, pubkey) VALUES ($1, $2)",
		GetAllArticlesInGroup:           "SELECT message_id FROM ArticlePosts WHERE newsgroup = $1",
		GetAllArticles:                  "SELECT message_id, newsgroup FROM ArticlePosts",
		GetMessageIDByHash:              "SELECT message_id, message_newsgroup FROM Articles WHERE message_id_hash = $1 LIMIT 1",
		CheckEncIPBanned:                "SELECT 1 FROM EncIPBans WHERE encaddr = $1",
		GetFirstAndLastForGroup:         "WITH x(min_no, max_no) AS ( SELECT MIN(message_no) AS min_no, MAX(message_no) AS max_no FROM ArticleNumbers WHERE newsgroup = $1) SELECT CASE WHEN min_no IS NULL THEN 0 ELSE min_no END AS min_no FROM x UNION SELECT CASE WHEN max_no IS NULL THEN 1 ELSE max_no END AS max_no FROM x",
		GetNewsgroupList:                "SELECT newsgroup, min(message_no), max(message_no) FROM ArticleNumbers WHERE newsgroup NOT IN ( SELECT newsgroup FROM bannedgroups ) GROUP BY newsgroup ORDER BY newsgroup",
		GetMessageIDForNNTPID:           "SELECT message_id FROM ArticleNumbers WHERE newsgroup = $1 AND message_no = $2 LIMIT 1",
		GetNNTPIDForMessageID:           "SELECT message_no FROM ArticleNumbers WHERE newsgroup = $1 AND message_id = $2 LIMIT 1",
		IsExpired:                       "WITH x(msgid) AS ( SELECT message_id FROM Articles WHERE message_id = $1 INTERSECT ( SELECT message_id FROM ArticlePosts WHERE message_id = $1 ) ) SELECT COUNT(*) FROM x",
		GetLastDaysPostsForGroup:        "SELECT COUNT(*) FROM ArticlePosts WHERE time_posted < $1 AND time_posted > $2 AND newsgroup = $3",
		GetLastDaysPosts:                "SELECT COUNT(*) FROM ArticlePosts WHERE time_posted < $1 AND time_posted > $2",
		GetLastPostedPostModels:         "SELECT newsgroup, message_id, ref_id, name, subject, path, time_posted, message, addr FROM ArticlePosts WHERE newsgroup != 'ctl' ORDER BY time_posted DESC LIMIT $1",
		GetMonthlyPostHistory:           "SELECT time_posted FROM ArticlePosts WHERE time_posted > 0 ORDER BY time_posted ASC LIMIT 1",
		CheckNNTPLogin:                  "SELECT login_hash, login_salt FROM NNTPUsers WHERE username = $1",
		CheckNNTPUserExists:             "SELECT 1 FROM NNTPUsers WHERE username = $1",
		GetHeadersForMessage:            "SELECT header_name, header_value FROM NNTPHeaders WHERE header_article_message_id = $1",
		CountAllArticlesInGroup:         "SELECT COUNT(message_id) FROM ArticlePosts WHERE newsgroup = $1",
		GetMessageIDByCIDR:              "SELECT message_id FROM ArticlePosts WHERE addr IN ( SELECT encaddr FROM EncryptedAddrs WHERE addr_cidr <<= cidr($1) )",
		GetMessageIDByEncryptedIP:       "SELECT message_id FROM ArticlePosts WHERE addr = $1",
		GetPostsBefore:                  "SELECT message_id FROM ArticlePosts WHERE time_posted < $1",
		SearchQuery_1:                   "SELECT newsgroup, message_id, ref_id FROM ArticlePosts WHERE message LIKE $1 ORDER BY time_posted DESC LIMIT $2",
		SearchQuery_2:                   "SELECT newsgroup, message_id, ref_id FROM ArticlePosts WHERE newsgroup = $1 AND message LIKE $2 ORDER BY time_posted DESC LIMIT $3",
		SearchByHash_1:                  "SELECT message_newsgroup, message_id, message_ref_id FROM Articles WHERE message_id_hash LIKE $1 ORDER BY time_obtained DESC LIMIT $2",
		SearchByHash_2:                  "SELECT message_newsgroup, message_id, message_ref_id FROM Articles WHERE message_newsgroup = $2 AND message_id_hash LIKE $1 ORDER BY time_obtained DESC LIMIT $3",
		GetNNTPPostsInGroup:             "SELECT message_no, ArticlePosts.message_id, subject, time_posted, ref_id, name, path FROM ArticleNumbers INNER JOIN ArticlePosts ON ArticleNumbers.message_id = ArticlePosts.message_id WHERE ArticlePosts.newsgroup = $1 ORDER BY message_no",
		GetCitesByPostHashLike:          "SELECT message_id, message_ref_id FROM Articles WHERE message_id_hash LIKE $1",
		GetYearlyPostHistory:            "WITH times(endtime, begintime) AS ( SELECT CAST(EXTRACT(epoch from i) AS BIGINT) AS endtime, CAST(EXTRACT(epoch from i - interval '1 month') AS BIGINT) AS begintime FROM generate_series(now() - interval '10 year', now(), '1 month'::interval) i ) SELECT begintime, endtime, ( SELECT count(*) FROM ArticlePosts WHERE time_posted > begintime AND time_posted < endtime) FROM times",
		CountUkko:                       "SELECT COUNT(message_id) FROM ArticlePosts WHERE newsgroup != 'ctl' AND ref_id = '' OR ref_id = message_id",
		RemoveArticle:                   "DELETE FROM Articles WHERE message_id = $1",
	}

}

func (self *PostgresDatabase) CreateTables() {
	for {
		version := self.getDBVersion()
		if version == -1 {
			// no tables
			self.createTablesV0()
			self.upgrade0to1()
		} else if version == 0 {
			// upgrade to version 1
			self.upgrade0to1()
		} else if version == 1 {
			// upgrade to version 2
			self.upgrade1to2()
		} else if version == 2 {
			// upgrade to version 3
			self.upgrade2to3()
		} else if version == 3 {
			// update to version 4
			self.upgrade3to4()
		} else if version == 4 {
			// update to version 5
			self.upgrade4to5()
		} else if version == 5 {
			// upgrade to version 6
			self.upgrade5to6()
		} else if version == 6 {
			// upgrade to version 7
			self.upgrade6to7()
		} else if version == 7 {
			self.upgrade7to8()
		} else if version == 8 {
			self.upgrade8to9()
		} else if version == 9 {
			// we are up to date
			log.Println("we are up to date at version", version)
			break
		}
	}
	self.prepareStatements()
}

func (self *PostgresDatabase) upgrade1to2() {
	log.Println("migrating... 1 -> 2")

	var err error

	tables := make(map[string]string)

	tables["NNTPUsers"] = `(
                           username VARCHAR(255) PRIMARY KEY,
                           login_hash VARCHAR(255) NOT NULL,
                           login_salt VARCHAR(255) NOT NULL
                         )`

	table_order := []string{"NNTPUsers"}

	for _, table := range table_order {
		q := tables[table]
		// create table
		_, err = self.conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s%s", table, q))
		if err != nil {
			log.Fatalf("cannot create table %s, %s, login was '%s'", table, err, self.db_str)
		}
	}
	self.setDBVersion(2)
}

func (self *PostgresDatabase) upgrade2to3() {
	log.Println("migrating... 2 -> 3")

	var err error

	tables := make(map[string]string)

	tables["NNTPHeaders"] = `(
                             header_name VARCHAR(255) NOT NULL,
                             header_value TEXT NOT NULL,
                             header_article_message_id VARCHAR(255) NOT NULL,
                             FOREIGN KEY(header_article_message_id) REFERENCES ArticlePosts(message_id)
                           )`

	table_order := []string{"NNTPHeaders"}

	for _, table := range table_order {
		q := tables[table]
		// create table
		_, err = self.conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s%s", table, q))
		if err != nil {
			log.Fatalf("cannot create table %s, %s, login was '%s'", table, err, self.db_str)
		}
	}
	cmds := []string{
		"CREATE INDEX ON NNTPHeaders(header_name)",
	}
	for _, cmd := range cmds {
		_, err = self.conn.Exec(cmd)
		checkError(err)
	}
	self.setDBVersion(3)
}

func (self *PostgresDatabase) upgrade5to6() {
	log.Println("migrating... 5 -> 6")
	tables := make(map[string]string)

	// public key properties, key value pair: pubkey -> status
	tables["PubkeyProperties"] = `(
                                  pubkey VARCHAR(255) PRIMARY KEY,
                                  status VARCHAR(255) NOT NULL
                                )`

	// ledger of public key property modification events
	tables["PubkeyModifyEvents"] = `(
                                     source_pubkey VARCHAR(255) NOT NULL,
                                     target_pubkey VARCHAR(255) NOT NULL,
                                     event_time BIGINT NOT NULL,
                                     status VARCHAR(255) NOT NULL,
                                     id BIGSERIAL PRIMARY KEY
                                   )`

	table_order := []string{"PubkeyProperties", "PubkeyModifyEvents"}
	for _, t := range table_order {
		q := tables[t]
		// create table
		_, err := self.conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s%s", t, q))
		if err != nil {
			log.Fatalf("cannot create table %s, %s", t, err)
		}
	}

	self.setDBVersion(6)
}

func (self *PostgresDatabase) upgrade4to5() {
	log.Println("migrating... 4 -> 5")
	cmds := []string{
		"ALTER TABLE EncryptedAddrs DROP COLUMN IF EXISTS addr_cidr",
		"ALTER TABLE EncryptedAddrs ADD COLUMN addr_cidr cidr",
		"UPDATE EncryptedAddrs AS a SET addr_cidr = e.cidr FROM ( SELECT cidr(addr), addr FROM EncryptedAddrs) AS e WHERE e.addr = a.addr",
	}
	for _, cmd := range cmds {
		_, err := self.conn.Exec(cmd)
		if err != nil {
			log.Fatalf("failed to execute query `%s`, %s", cmd, err.Error())
		}
	}
	self.setDBVersion(5)
}

func (self *PostgresDatabase) upgrade3to4() {
	log.Println("migrating... 3 -> 4")
	tables := make(map[string]string)
	tables["ArticleNumbers"] = `(
                                newsgroup VARCHAR(255) NOT NULL,
                                message_id VARCHAR(255) NOT NULL,
                                message_no BIGINT NOT NULL,
                                FOREIGN KEY (newsgroup) REFERENCES Newsgroups(name),
                                FOREIGN KEY (message_id) REFERENCES ArticlePosts(message_id)
                              )`
	table_order := []string{"ArticleNumbers"}
	cmds := []string{"CREATE INDEX ON ArticleNumbers(message_no)"}
	for _, table := range table_order {
		_, err := self.conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s%s", table, tables[table]))
		if err != nil {
			log.Fatalf("cannot create table %s: %s", table, err.Error())
		}
	}

	for _, cmd := range cmds {
		_, err := self.conn.Exec(cmd)
		if err != nil {
			log.Fatalf("failed to execute query: %s, %s", cmd, err.Error())
		}
	}

	log.Println("migrating post numbers, this can take a bit DO NOT INTERRUPT")
	rows, err := self.conn.Query("SELECT message_id, newsgroup FROM ArticlePosts ORDER BY time_posted DESC")
	if err != nil {
		log.Fatalf("could not query ArticlePosts table: %s", err.Error())
	}
	counter := int64(0)
	for rows.Next() {
		counter++
		var msgid, group string
		err = rows.Scan(&msgid, &group)
		if err != nil {
			log.Fatalf("could not scan row: %s", err.Error())
		}
		err = self.registerNNTPNumber(group, msgid)
		if err != nil {
			log.Fatalf("could not migrate article %s in %s, %s", msgid, group, err.Error())
		}
		if counter%100 == 0 {
			log.Println("migrated ", counter)
		}
	}
	log.Println("total migrated posts: ", counter)
	self.setDBVersion(4)
}

func (self *PostgresDatabase) upgrade0to1() {

	// begin >:D
	log.Println("migrating... 0 -> 1")

	var err error

	cmds := []string{
		// newsgroups table
		"CREATE INDEX ON Newsgroups(name)",
		// article posts table
		"ALTER TABLE ArticlePosts DROP COLUMN IF EXISTS addr",
		"ALTER TABLE ArticlePosts ADD COLUMN addr VARCHAR(255)",
		"ALTER TABLE ArticlePosts DROP CONSTRAINT IF EXISTS group_depend",
		"ALTER TABLE ArticlePosts ADD CONSTRAINT group_depend FOREIGN KEY(newsgroup) REFERENCES Newsgroups(name) ON DELETE CASCADE",
		"ALTER TABLE ArticlePosts DROP CONSTRAINT IF EXISTS msgid_pk",
		"ALTER TABLE ArticlePosts ADD CONSTRAINT msgid_pk PRIMARY KEY(message_id)",
		"CREATE INDEX ON ArticlePosts(ref_id)",
		// article keys table
		"DELETE FROM ArticleKeys WHERE message_id NOT IN ( SELECT message_id FROM ArticlePosts )",
		"ALTER TABLE ArticleKeys DROP CONSTRAINT IF EXISTS msgid_depend",
		"ALTER TABLE ArticleKeys ADD CONSTRAINT msgid_depend FOREIGN KEY(message_id) REFERENCES ArticlePosts(message_id) ON DELETE CASCADE",
		// article threads table
		"ALTER TABLE ArticleThreads DROP CONSTRAINT IF EXISTS msgid_depend",
		"ALTER TABLE ArticleThreads DROP CONSTRAINT IF EXISTS group_depend",
		"DELETE FROM ArticleThreads WHERE root_message_id NOT IN ( SELECT message_id FROM ArticlePosts )",
		"ALTER TABLE ArticleThreads ADD CONSTRAINT msgid_depend FOREIGN KEY(root_message_id) REFERENCES ArticlePosts(message_id) ON DELETE CASCADE",
		"ALTER TABLE ArticleThreads ADD CONSTRAINT group_depend FOREIGN KEY(newsgroup) REFERENCES Newsgroups(name) ON DELETE CASCADE",
		// article attachments table
		"ALTER TABLE ArticleAttachments DROP CONSTRAINT IF EXISTS msgid_depend",
		"DELETE FROM ArticleAttachments WHERE message_id NOT IN ( SELECT message_id FROM ArticlePosts )",
		"ALTER TABLE ArticleAttachments ADD CONSTRAINT msgid_depend FOREIGN KEY(message_id) REFERENCES ArticlePosts(message_id) ON DELETE CASCADE",
	}

	for _, cmd := range cmds {
		_, err = self.conn.Exec(cmd)
		checkError(err)
	}
	self.setDBVersion(1)
}

func (self *PostgresDatabase) upgrade6to7() {
	tables := make(map[string]string)
	log.Println("migrating... 6 -> 7")
	// table for thumbnail info
	tables["Thumbnails"] = `(
                            sha_hash VARCHAR(128) PRIMARY KEY,
                            width INTEGER NOT NULL,
                            height INTEGER NOT NULL
                          )`

	tables["Cites"] = `(
                            post_msgid VARCHAR(255) NOT NULL,
                            cite_msgid VARCHAR(255) NOT NULL
                     )`

	var err error

	table_order := []string{"Thumbnails", "Cites"}
	for _, table := range table_order {
		q := tables[table]
		// create table
		_, err = self.conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s%s", table, q))
		if err != nil {
			log.Fatalf("cannot create table %s, %s, login was '%s'", table, err, self.db_str)
		}
	}

	// make indexes
	cmds := []string{
		"CREATE INDEX ON Thumbnails(sha_hash)",
		"CREATE INDEX ON Cites(cite_msgid)",
	}

	for _, cmd := range cmds {
		_, err = self.conn.Exec(cmd)
		checkError(err)
	}

	/*
		// rebuild ALL cites
		log.Println("!!! Building Cites table, this will take a long time. Do NOT interrupt !!!")

		post_counter := 0
		cite_counter := 0
		var rows *sql.Rows
		rows, err = self.conn.Query("SELECT message, message_id FROM ArticlePosts")
		if err != nil {
			log.Fatalf("error migrating: %s", err)
		}

		cites := make(map[string][]string)

		for rows.Next() {
			var msg, msgid string
			rows.Scan(&msg, &msgid)
			c := findBacklinks(msg)
			cite_counter += len(c)
			cites[msgid] = c
			post_counter++
			if post_counter%100 == 0 {
				log.Printf("selected %d messages %d cites", post_counter, cite_counter)
			}
		}

		rows.Close()

		log.Printf("calculating %d cites ...", cite_counter)

		cites_insert := make(map[string][2]string)

		citemap_counter := 0

		for msgid, citelist := range cites {
			for _, cite := range citelist {
				cite = cite[2:]
				citeLike := cite + "%"
				var cite_msgid string
				err = self.conn.QueryRow("SELECT message_id FROM Articles WHERE message_id_hash LIKE $1 LIMIT 1", citeLike).Scan(&cite_msgid)
				if err != nil {
					continue
					//log.Fatalf("failed to select cite like %s: %s", citeLike, err)
				}
				cites_insert[msgid+cite_msgid] = [2]string{msgid, cite_msgid}
				citemap_counter++
				if cite_counter%100 == 0 {
					log.Printf("calculated %d cites", cite_counter)
				}
			}
		}

		log.Printf("inserting %d cites ...", cite_counter)

		txn, err := self.conn.Begin()
		if err != nil {
			log.Fatalf("failed to begin insert: %s", err)
		}

		st, err := txn.Prepare(pq.CopyIn("Cites", "post_msgid", "cite_msgid"))

		if err != nil {
			log.Fatalf("failed to prepare statement: %s", err)
		}

		for _, ct := range cites_insert {
			_, err = st.Exec(ct[0], ct[1])
			if err != nil {
				log.Fatalf("failed to insert with prepared statement: %s", err)
			}
		}

		_, err = st.Exec()

		if err != nil {
			log.Fatalf("failed to excute statement: %s", err)
		}

		err = st.Close()

		if err != nil {
			log.Fatalf("failed to close statement: %s", err)
		}

		log.Println("committing...")

		err = txn.Commit()
		if err != nil {
			log.Fatalf("failed to commit transaction: %s", err)
		}

		log.Println("insertion done")
	*/
	self.setDBVersion(7)
}

func (self *PostgresDatabase) upgrade7to8() {
	log.Println("migrating 7 -> 8")
	cmds := []string{
		"ALTER TABLE ArticleNumbers DROP CONSTRAINT IF EXISTS articlenumbers_message_id_fkey",
		"ALTER TABLE ArticleNumbers ADD CONSTRAINT msgid_depends FOREIGN KEY (message_id) REFERENCES ArticlePosts(message_id) ON DELETE CASCADE",
		"ALTER TABLE NNTPHeaders DROP CONSTRAINT IF EXISTS nntpheaders_header_article_message_id_fkey",
		"ALTER TABLE NNTPHeaders ADD CONSTRAINT msgid_depends FOREIGN KEY (header_article_message_id) REFERENCES ArticlePosts(message_id) ON DELETE CASCADE",
	}
	for _, cmd := range cmds {
		log.Println("exec", cmd)
		_, err := self.conn.Exec(cmd)
		if err != nil {
			log.Fatalf("%s: %s", cmd, err.Error())
		}
	}
	self.setDBVersion(8)
}

func (self *PostgresDatabase) upgrade8to9() {
	cmds := []string{
		"ALTER TABLE ArticlePosts ADD COLUMN frontendpubkey TEXT",
		"CREATE TABLE IF NOT EXISTS nntpchan_pubkeys(status VARCHAR(16) NOT NULL, pubkey VARCHAR(64) PRIMARY KEY)",
	}
	for _, cmd := range cmds {
		_, err := self.conn.Exec(cmd)
		if err != nil {
			log.Fatalf("%s: %s", cmd, err.Error())
		}
	}
	self.setDBVersion(9)
}

// create all tables for database version 0
func (self *PostgresDatabase) createTablesV0() {
	tables := make(map[string]string)

	// table of active newsgroups
	tables["Newsgroups"] = `(
                            name VARCHAR(255) PRIMARY KEY,
                            last_post INTEGER NOT NULL,
                            restricted BOOLEAN
                          )`

	// table for ip and their encryption key
	tables["EncryptedAddrs"] = `(
                                enckey VARCHAR(255) NOT NULL,
                                addr VARCHAR(255) NOT NULL,
                                encaddr VARCHAR(255) NOT NULL
                              )`

	// table for articles that have been banned
	tables["BannedArticles"] = `(
                                message_id VARCHAR(255) PRIMARY KEY,
                                time_banned INTEGER NOT NULL,
                                ban_reason TEXT NOT NULL
                              )`

	// table for banned newsgroups
	tables["BannedGroups"] = `(
                             newsgroup VARCHAR(255) PRIMARY KEY,
                             time_banned INTEGER NOT NULL
                           )`

	// table for storing nntp article meta data
	tables["Articles"] = `(
                          message_id VARCHAR(255) PRIMARY KEY,
                          message_id_hash VARCHAR(40) UNIQUE NOT NULL,
                          message_newsgroup VARCHAR(255),
                          message_ref_id VARCHAR(255),
                          time_obtained INTEGER NOT NULL,
                          FOREIGN KEY(message_newsgroup) REFERENCES Newsgroups(name)
                        )`

	// table for storing nntp article post content
	tables["ArticlePosts"] = `(
                              newsgroup VARCHAR(255),
                              message_id VARCHAR(255),
                              ref_id VARCHAR(255),
                              name TEXT NOT NULL,
                              subject TEXT NOT NULL,
                              path TEXT NOT NULL,
                              time_posted INTEGER NOT NULL,
                              message TEXT NOT NULL
                            )`

	// table for storing nntp article posts to pubkey mapping
	tables["ArticleKeys"] = `(
                             message_id VARCHAR(255) NOT NULL,
                             pubkey VARCHAR(255) NOT NULL
                           )`

	// table for thread state
	tables["ArticleThreads"] = `(
                                newsgroup VARCHAR(255) NOT NULL,
                                root_message_id VARCHAR(255) NOT NULL,
                                last_bump INTEGER NOT NULL,
                                last_post INTEGER NOT NULL
                              )`

	// table for storing nntp article attachment info
	tables["ArticleAttachments"] = `(
                                    message_id VARCHAR(255),
                                    sha_hash VARCHAR(128) NOT NULL,
                                    filename TEXT NOT NULL,
                                    filepath TEXT NOT NULL
                                  )`

	// table for storing current permissions of mod pubkeys
	tables["ModPrivs"] = `(
                          pubkey VARCHAR(255),
                          newsgroup VARCHAR(255),
                          permission VARCHAR(255)
                        )`

	// table for storing moderation events
	tables["ModLogs"] = `(
                         pubkey VARCHAR(255),
                         action VARCHAR(255),
                         target VARCHAR(255),
                         time INTEGER
                       )`

	// ip range bans
	tables["IPBans"] = `(
                        addr cidr NOT NULL,
                        made INTEGER NOT NULL,
                        expires INTEGER NOT NULL
                      )`
	// bans for encrypted addresses that we don't have the ip for
	tables["EncIPBans"] = `(
                           encaddr VARCHAR(255) NOT NULL,
                           made INTEGER NOT NULL,
                           expires INTEGER NOT NULL
                         )`

	tables["Settings"] = `(
                           name VARCHAR(255) NOT NULL,
                           value VARCHAR(255) NOT NULL
                        )`
	var err error

	table_order := []string{"Newsgroups", "BannedGroups", "BannedArticles", "IPBans", "EncIPBans", "Settings", "Articles", "ArticlePosts", "ArticleKeys", "ArticleThreads", "ArticleAttachments", "ModPrivs", "ModLogs", "EncryptedAddrs"}
	for _, table := range table_order {
		q := tables[table]
		// create table
		_, err = self.conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s%s", table, q))
		if err != nil {
			log.Fatalf("cannot create table %s, %s, login was '%s'", table, err, self.db_str)
		}
	}
	// create indexes
	_, err = self.conn.Exec("CREATE INDEX IF NOT EXISTS ON ArticleThreads(root_message_id)")
	_, err = self.conn.Exec("CREATE INDEX IF NOT EXISTS ON ArticleAttachments(message_id)")
	_, err = self.conn.Exec("CREATE INDEX IF NOT EXISTS ON ArticlePosts(message_id)")
	_, err = self.conn.Exec("CREATE INDEX IF NOT EXISTS ON Articles(message_id)")
	_, err = self.conn.Exec("CREATE INDEX IF NOT EXISTS ON Newsgroups(name)")

	self.setDBVersion(0)
}

// set what the current database version is
func (self *PostgresDatabase) setDBVersion(version int) (err error) {
	log.Println("set db version to", version)
	_, err = self.conn.Exec("DELETE FROM Settings WHERE name = $1", "version")
	_, err = self.conn.Exec("INSERT INTO Settings(name, value) VALUES($1, $2)", "version", fmt.Sprintf("%d", version))
	return
}

// get the current database version
func (self *PostgresDatabase) getDBVersion() (version int) {
	var val string
	var vers int64
	err := self.conn.QueryRow("SELECT value FROM Settings WHERE name = $1", "version").Scan(&val)
	if err == nil {
		vers, err = strconv.ParseInt(val, 10, 32)
		if err == nil {
			version = int(vers)
		} else {
			log.Fatal("cannot figure out db version", err)
		}
	} else {
		version = -1
	}
	return
}

func (self *PostgresDatabase) BanNewsgroup(group string) (err error) {
	_, err = self.conn.Exec("INSERT INTO BannedGroups(newsgroup, time_banned) VALUES($1, $2)", group, timeNow())
	return
}

func (self *PostgresDatabase) UnbanNewsgroup(group string) (err error) {
	_, err = self.conn.Exec("DELETE FROM BannedGroups WHERE newsgroup = $1", group)
	return
}

func (self *PostgresDatabase) NewsgroupBanned(group string) (banned bool, err error) {
	var count int64
	err = self.conn.QueryRow(self.stmt[NewsgroupBanned], group).Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	}
	banned = count > 0
	return
}

func (self *PostgresDatabase) NukeNewsgroup(group string, store ArticleStore) {
	// first delete all thread presences
	_, _ = self.conn.Exec("DELETE FROM ArticleThreads WHERE newsgroup = $1", group)
	// get all articles in that newsgroup
	chnl := make(chan ArticleEntry, 24)
	go func() {
		self.GetAllArticlesInGroup(group, chnl)
		close(chnl)
	}()
	// for each article delete it fully
	for {
		article, ok := <-chnl
		if ok {
			msgid := article.MessageID()
			log.Println("delete", msgid)
			// remove article from store
			fname := store.GetFilename(msgid)
			os.Remove(fname)
			// get all attachments
			for _, att := range self.GetPostAttachments(msgid) {
				// remove attachment
				log.Println("delete attachment", att)
				os.Remove(store.ThumbnailFilepath(att))
				os.Remove(store.AttachmentFilepath(att))
			}
			// delete from database
			self.DeleteArticle(msgid)
		} else {
			log.Println("nuke of", group, "done")
			return
		}
	}
}

func (self *PostgresDatabase) AddModPubkey(pubkey string) error {
	if self.CheckModPubkey(pubkey) {
		log.Println("did not add pubkey", pubkey, "already exists")
		return nil
	}
	_, err := self.conn.Exec("INSERT INTO ModPrivs(pubkey, newsgroup, permission) VALUES ( $1, $2, $3 )", pubkey, "ctl", "login")
	return err
}

func (self *PostgresDatabase) GetGroupForMessage(message_id string) (group string, err error) {
	err = self.conn.QueryRow("SELECT newsgroup FROM ArticlePosts WHERE message_id = $1", message_id).Scan(&group)
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (self *PostgresDatabase) GetPageForRootMessage(root_message_id string) (group string, page int64, err error) {
	err = self.conn.QueryRow("SELECT newsgroup FROM ArticleThreads WHERE root_message_id = $1", root_message_id).Scan(&group)
	if err == nil {
		perpage, _ := self.GetPagesPerBoard(group)
		err = self.conn.QueryRow("WITH thread(bump) AS (SELECT last_bump FROM ArticleThreads WHERE root_message_id = $1 ) SELECT COUNT(*) FROM ( SELECT last_bump FROM ArticleThreads INNER JOIN thread ON (thread.bump <= ArticleThreads.last_bump AND newsgroup = $2 ) ) AS amount", root_message_id, group).Scan(&page)
		return group, page / int64(perpage), err
	}
	return
}

func (self *PostgresDatabase) GetInfoForMessage(msgid string) (root string, newsgroup string, page int64, err error) {
	err = self.conn.QueryRow("SELECT newsgroup, ref_id FROM ArticlePosts WHERE message_id = $1", msgid).Scan(&newsgroup, &root)
	if err == nil {
		if root == "" {
			root = msgid
		}
		perpage, _ := self.GetPagesPerBoard(newsgroup)
		err = self.conn.QueryRow("WITH thread(bump) AS (SELECT last_bump FROM ArticleThreads WHERE root_message_id = $1 ) SELECT COUNT(*) FROM ( SELECT last_bump FROM ArticleThreads INNER JOIN thread ON (thread.bump <= ArticleThreads.last_bump AND newsgroup = $2 ) ) AS amount", root, newsgroup).Scan(&page)
		page = page / int64(perpage)
	} else if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (self *PostgresDatabase) CheckModPubkeyGlobal(pubkey string) bool {
	var result int64
	_ = self.conn.QueryRow("SELECT COUNT(*) FROM ModPrivs WHERE pubkey = $1 AND newsgroup = $2 AND permission = $3", pubkey, "overchan", "all").Scan(&result)
	return result > 0
}

func (self *PostgresDatabase) CheckModPubkeyCanModGroup(pubkey, newsgroup string) bool {
	var result int64
	_ = self.conn.QueryRow("SELECT COUNT(*) FROM ModPrivs WHERE pubkey = $1 AND newsgroup = $2", pubkey, newsgroup).Scan(&result)
	return result > 0
}

func (self *PostgresDatabase) CountPostsInGroup(newsgroup string, time_frame int64) (result int64) {
	if time_frame > 0 {
		time_frame = timeNow() - time_frame
	} else if time_frame < 0 {
		time_frame = 0
	}
	self.conn.QueryRow("SELECT COUNT(*) FROM ArticlePosts WHERE time_posted > $2 AND newsgroup = $1", newsgroup, time_frame).Scan(&result)
	return
}

func (self *PostgresDatabase) CheckModPubkey(pubkey string) bool {
	var result int64
	self.conn.QueryRow("SELECT COUNT(*) FROM ModPrivs WHERE pubkey = $1", pubkey).Scan(&result)
	return result > 0
}

func (self *PostgresDatabase) BanArticle(messageID, reason string) error {
	if self.ArticleBanned(messageID) {
		log.Println(messageID, "already banned")
		return nil
	}
	_, err := self.conn.Exec("INSERT INTO BannedArticles(message_id, time_banned, ban_reason) VALUES($1, $2, $3)", messageID, timeNow(), reason)
	return err
}

func (self *PostgresDatabase) ArticleBanned(messageID string) (result bool) {

	var count int64
	err := self.conn.QueryRow(self.stmt[ArticleBanned], messageID).Scan(&count)
	if err == nil {
		result = count > 0
	} else if err != sql.ErrNoRows {
		log.Println("error checking if article is banned", err)
	}
	return
}

func (self *PostgresDatabase) GetEncAddress(addr string) (encaddr string, err error) {
	var count int64
	err = self.conn.QueryRow("SELECT COUNT(addr) FROM EncryptedAddrs WHERE addr = $1", addr).Scan(&count)
	if err == nil {
		if count == 0 {
			// needs to be inserted
			var key string
			key, encaddr = newAddrEnc(addr)
			if len(encaddr) == 0 {
				err = errors.New("failed to generate new encryption key")
			} else {
				_, err = self.conn.Exec("INSERT INTO EncryptedAddrs(enckey, encaddr, addr, addr_cidr) VALUES($1, $2, $3, cidr($4))", key, encaddr, addr, addr+"/32")
			}
		} else {
			err = self.conn.QueryRow("SELECT encAddr FROM EncryptedAddrs WHERE addr = $1 LIMIT 1", addr).Scan(&encaddr)
		}
	}
	return
}

func (self *PostgresDatabase) GetEncKey(encAddr string) (enckey string, err error) {
	err = self.conn.QueryRow("SELECT enckey FROM EncryptedAddrs WHERE encaddr = $1 LIMIT 1", encAddr).Scan(&enckey)
	return
}

func (self *PostgresDatabase) CheckIPBanned(addr string) (banned bool, err error) {
	var amount int64
	err = self.conn.QueryRow("SELECT COUNT(*) FROM IPBans WHERE addr >>= $1 ", addr).Scan(&amount)
	banned = amount > 0
	return
}

func (self *PostgresDatabase) GetIPAddress(encaddr string) (addr string, err error) {
	var count int64
	err = self.conn.QueryRow("SELECT COUNT(encAddr) FROM EncryptedAddrs WHERE encAddr = $1", encaddr).Scan(&count)
	if err == nil && count > 0 {
		err = self.conn.QueryRow("SELECT addr FROM EncryptedAddrs WHERE encAddr = $1 LIMIT 1", encaddr).Scan(&addr)
	}
	return
}

func (self *PostgresDatabase) MarkModPubkeyGlobal(pubkey string) (err error) {
	if len(pubkey) != 64 {
		err = errors.New("invalid pubkey length")
		return
	}
	if self.CheckModPubkeyGlobal(pubkey) {
		// already marked
		log.Println("pubkey already marked as global", pubkey)
	} else {
		_, err = self.conn.Exec("INSERT INTO ModPrivs(pubkey, newsgroup, permission) VALUES ( $1, $2, $3 )", pubkey, "overchan", "all")
	}
	return
}

func (self *PostgresDatabase) MarkPubkeyAdmin(pubkey string) (err error) {
	var admin bool
	admin, err = self.CheckAdminPubkey(pubkey)
	if err == nil && !admin {
		// add as admin since it's not already there
		_, err = self.conn.Exec("INSERT INTO ModPrivs(pubkey, newsgroup, permission) VALUES ( $1, $2, $3 )", pubkey, "overchan", "admin")
	}
	return
}

func (self *PostgresDatabase) UnmarkPubkeyAdmin(pubkey string) (err error) {
	_, err = self.conn.Exec("DELETE FROM ModPrivs WHERE pubkey = $1 AND permission = $2", pubkey, "admin")
	return
}

func (self *PostgresDatabase) CheckAdminPubkey(pubkey string) (admin bool, err error) {
	var count int64
	err = self.conn.QueryRow("SELECT COUNT(pubkey) FROM ModPrivs WHERE pubkey = $1 AND permission = $2", pubkey, "admin").Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err == nil {
		admin = count > 0
	}
	return
}

func (self *PostgresDatabase) UnMarkModPubkeyGlobal(pubkey string) (err error) {
	if self.CheckModPubkeyGlobal(pubkey) {
		// already marked
		_, err = self.conn.Exec("DELETE FROM ModPrivs WHERE pubkey = $1 AND newsgroup = $2 AND permission = $3", pubkey, "overchan", "all")
	} else {
		err = errors.New("public key not marked as global")
	}
	return
}

func (self *PostgresDatabase) CountThreadReplies(root_message_id string) (repls int64) {
	_ = self.conn.QueryRow("SELECT COUNT(message_id) FROM ArticlePosts WHERE ref_id = $1", root_message_id).Scan(&repls)
	return
}

func (self *PostgresDatabase) GetRootPostsForExpiration(newsgroup string, threadcount int) (roots []string) {

	rows, err := self.conn.Query("SELECT root_message_id FROM ArticleThreads WHERE newsgroup = $1 AND root_message_id NOT IN ( SELECT root_message_id FROM ArticleThreads WHERE newsgroup = $1 ORDER BY last_bump DESC LIMIT $2)", newsgroup, threadcount)
	if err == nil {
		// get results
		for rows.Next() {
			var root string
			rows.Scan(&root)
			roots = append(roots, root)
			log.Println(root)
		}
		rows.Close()
	} else {
		log.Println("failed to get root posts for expiration", err)
	}
	// return the list of expired roots
	return
}

// register an article in a newsgroup with the ArticleNumbers table
func (self *PostgresDatabase) registerNNTPNumber(group, msgid string) (err error) {
	_, err = self.conn.Exec("WITH x(msg_no) AS ( SELECT MAX(message_no) AS msg_no FROM ArticleNumbers WHERE newsgroup = $1 ) INSERT INTO ArticleNumbers(newsgroup, message_id, message_no) VALUES($1, $2, (SELECT CASE WHEN msg_no IS NULL THEN 0 ELSE msg_no END FROM x) + 1 )", group, msgid)
	return
}

func (self *PostgresDatabase) GetAllNewsgroups() (groups []string) {

	rows, err := self.conn.Query(self.stmt[GetAllNewsgroups])
	if err == nil {
		for rows.Next() {
			var group string
			rows.Scan(&group)
			groups = append(groups, group)
		}
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) GetGroupPageCount(newsgroup string) int64 {
	var count int64
	err := self.conn.QueryRow("SELECT COUNT(*) FROM ArticleThreads WHERE newsgroup = $1", newsgroup).Scan(&count)
	if err != nil {
		log.Println("failed to count pages in group", newsgroup, err)
	}
	// divide by threads per page
	return int64(math.Ceil(float64(count/10)) + 1)
}

// only fetches root posts
// does not update the thread contents
func (self *PostgresDatabase) GetGroupForPage(prefix, frontend, newsgroup string, pageno, perpage int) BoardModel {
	var threads []ThreadModel
	pages := self.GetGroupPageCount(newsgroup)
	rows, err := self.conn.Query("WITH roots(root_message_id, last_bump) AS ( SELECT root_message_id, last_bump FROM ArticleThreads WHERE newsgroup = $1 ORDER BY last_bump DESC OFFSET $2 LIMIT $3 ) SELECT p.newsgroup, p.message_id, p.name, p.subject, p.path, p.time_posted, p.message, p.addr FROM ArticlePosts p INNER JOIN roots ON ( roots.root_message_id = p.message_id ) ORDER BY roots.last_bump DESC", newsgroup, pageno*perpage, perpage)
	if err == nil {
		for rows.Next() {

			p := &post{
				prefix: prefix,
			}
			rows.Scan(&p.board, &p.Message_id, &p.PostName, &p.PostSubject, &p.MessagePath, &p.Posted, &p.PostMessage, &p.addr)
			p.Parent = p.Message_id
			p.op = true
			_ = self.conn.QueryRow("SELECT pubkey FROM ArticleKeys WHERE message_id = $1", p.Message_id).Scan(&p.Key)
			p.sage = isSage(p.PostSubject)
			atts := self.GetPostAttachmentModels(prefix, p.Message_id)
			if atts != nil {
				p.Files = append(p.Files, atts...)
			}
			threads = append(threads, createThreadModel(p))
		}
		rows.Close()
	} else {
		log.Println("failed to fetch board model for", newsgroup, "page", pageno, err)
	}
	return &boardModel{
		prefix:   prefix,
		frontend: frontend,
		board:    newsgroup,
		page:     pageno,
		pages:    int(pages),
		threads:  threads,
	}
}

func (self *PostgresDatabase) GetNNTPPostsInGroup(newsgroup string) (models []PostModel, err error) {
	rows, err := self.conn.Query(self.stmt[GetNNTPPostsInGroup], newsgroup)
	if err == nil {
		for rows.Next() {
			model := new(post)
			model.Newsgroup = newsgroup
			rows.Scan(&model.nntp_id, &model.Message_id, &model.PostSubject, &model.Posted, &model.Parent, &model.PostName, &model.MessagePath)
			models = append(models, model)
		}
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) GetPostsInGroup(newsgroup string) (models []PostModel, err error) {
	rows, err := self.conn.Query(self.stmt[GetPostsInGroup], newsgroup)
	if err == nil {
		for rows.Next() {
			model := new(post)
			rows.Scan(&model.board, &model.Message_id, &model.Parent, &model.PostName, &model.PostSubject, &model.MessagePath, &model.Posted, &model.PostMessage, &model.addr, &model.FrontendPublicKey)
			models = append(models, model)
		}
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) GetPostModel(prefix, messageID string) PostModel {
	model := new(post)
	err := self.conn.QueryRow(self.stmt[GetPostModel], messageID).Scan(&model.board, &model.Message_id, &model.Parent, &model.PostName, &model.PostSubject, &model.MessagePath, &model.Posted, &model.PostMessage, &model.addr, &model.FrontendPublicKey)
	if err == nil {
		model.op = len(model.Parent) == 0
		if len(model.Parent) == 0 {
			model.Parent = model.Message_id
		}
		model.sage = isSage(model.PostSubject)
		atts := self.GetPostAttachmentModels(prefix, messageID)
		if atts != nil {
			model.Files = append(model.Files, atts...)
		}
		// quiet fail
		self.conn.QueryRow(self.stmt[GetArticlePubkey], messageID).Scan(&model.Key)
		return model
	} else if err != sql.ErrNoRows {
		log.Println("failed to prepare query for geting post model for", messageID, err)
		return nil
	}
	return nil
}

func (self *PostgresDatabase) GetCitesByPostHashLike(like string) (cites []MessageIDTuple, err error) {
	var rows *sql.Rows
	rows, err = self.conn.Query(self.stmt[GetCitesByPostHashLike], like+"%")
	if err == nil {
		for rows.Next() {
			var tup MessageIDTuple
			rows.Scan(&tup[0], &tup[1])
			cites = append(cites, tup)
		}
		rows.Close()
	} else if err != sql.ErrNoRows {
		log.Println("error getting post models like", like, err)
	}
	return
}

func (self *PostgresDatabase) GetThreadModel(prefix, msgid string) (th ThreadModel, err error) {

	var posts []PostModel
	var rows *sql.Rows
	pmap := make(map[string]*post)
	rows, err = self.conn.Query(self.stmt[GetThreadModel], msgid)
	for err == nil && rows.Next() {
		p := new(post)
		p.Parent = msgid
		err = rows.Scan(&p.board, &p.Message_id, &p.PostName, &p.PostSubject, &p.Posted, &p.PostMessage, &p.addr, &p.FrontendPublicKey)
		pmap[p.Message_id] = p
		posts = append(posts, p)
	}
	rows.Close()
	rows, err = self.conn.Query(self.stmt[GetThreadModelAttachments], msgid)
	for err == nil && rows.Next() {
		att := &attachment{
			prefix: prefix,
		}
		var att_msgid string
		rows.Scan(&att.Name, &att.Path, &att_msgid)
		p, ok := pmap[att_msgid]
		if ok {
			p.Files = append(p.Files, att)
		}
	}
	rows.Close()
	rows, err = self.conn.Query(self.stmt[GetThreadModelPubkeys], msgid)
	if err != nil {
		log.Println(err)
	}
	for err == nil && rows.Next() {
		var key_msgid, key string
		rows.Scan(&key, &key_msgid)
		p, ok := pmap[key_msgid]
		if ok {
			p.Key = key
		}
	}
	rows.Close()
	th = createThreadModel(posts...)
	return
}

func (self *PostgresDatabase) DeleteThread(msgid string) (err error) {
	_, err = self.conn.Exec(self.stmt[DeleteThreadV8], msgid)
	return
}

func (self *PostgresDatabase) DeleteArticle(msgid string) (err error) {
	/*
		for _, q := range []string{DeleteArticle_1, DeleteArticle_2, DeleteArticle_3, DeleteArticle_4, DeleteArticle_5} {
			_, err = self.conn.Exec(self.stmt[q], msgid)
			if err != nil {
				break
			}
		}
	*/
	_, err = self.conn.Exec(self.stmt[DeleteArticleV8], msgid)
	return

}
func (self *PostgresDatabase) RemoveArticle(msgid string) (err error) {
	_, err = self.conn.Exec(self.stmt[DeleteArticleV8], msgid)
	if err == nil {
		_, err = self.conn.Exec(self.stmt[RemoveArticle], msgid)
	}
	return
}

func (self *PostgresDatabase) GetThreadReplyPostModels(prefix, rootpost string, start, limit int) (repls []PostModel) {
	var rows *sql.Rows
	var err error
	if limit > 0 {
		rows, err = self.conn.Query(self.stmt[GetThreadReplyPostModels_1], rootpost, limit)
	} else {
		rows, err = self.conn.Query(self.stmt[GetThreadReplyPostModels_2], rootpost)
	}
	offset := start
	if err == nil {
		for rows.Next() {
			// TODO: this is a hack, optimize queries plz
			if offset > 0 {
				offset--
				continue
			}
			model := new(post)
			model.prefix = prefix
			rows.Scan(&model.board, &model.Message_id, &model.Parent, &model.PostName, &model.PostSubject, &model.MessagePath, &model.Posted, &model.PostMessage, &model.addr, &model.FrontendPublicKey)
			model.op = len(model.Parent) == 0
			if len(model.Parent) == 0 {
				model.Parent = model.Message_id
			}
			model.sage = isSage(model.PostSubject)
			atts := self.GetPostAttachmentModels(prefix, model.Message_id)
			if atts != nil {
				model.Files = append(model.Files, atts...)
			}
			// get pubkey if it exists
			// quiet fail
			self.conn.QueryRow(self.stmt[GetArticlePubkey], model.Message_id).Scan(model.Key)
			repls = append(repls, model)
		}
		rows.Close()
	} else {
		log.Println("failed to get thread replies", rootpost, err)
	}

	return

}

func (self *PostgresDatabase) GetThreadReplies(rootpost string, start, limit int) (repls []string) {
	var rows *sql.Rows
	var err error
	if limit > 0 {
		rows, err = self.conn.Query(self.stmt[GetThreadReplies_1], rootpost, limit)
	} else {
		rows, err = self.conn.Query(self.stmt[GetThreadReplies_2], rootpost)
	}
	offset := start
	if err == nil {
		for rows.Next() {
			// TODO: this is a hack, optimize queries plz
			if offset > 0 {
				offset--
				continue
			}
			var msgid string
			rows.Scan(&msgid)
			repls = append(repls, msgid)
		}
		rows.Close()
	} else {
		log.Println("failed to get thread replies", rootpost, err)
	}
	return
}

func (self *PostgresDatabase) ThreadHasReplies(rootpost string) bool {
	var count int64
	err := self.conn.QueryRow("SELECT COUNT(message_id) FROM ArticlePosts WHERE ref_id = $1", rootpost).Scan(&count)
	if err != nil {
		log.Println("failed to count thread replies", err)
	}
	return count > 0
}

func (self *PostgresDatabase) GetGroupThreads(group string, recv chan ArticleEntry) {
	rows, err := self.conn.Query(self.stmt[GetGroupThreads], group)
	if err == nil {
		for rows.Next() {
			var msgid string
			rows.Scan(&msgid)
			recv <- ArticleEntry{msgid, group}
		}
		rows.Close()
	} else if err != sql.ErrNoRows {
		log.Println("failed to get group threads", err)
	}
}

func (self *PostgresDatabase) GetLastBumpedThreads(newsgroups string, threads int) []ArticleEntry {
	return self.GetLastBumpedThreadsPaginated(newsgroups, threads, 0)
}

func (self *PostgresDatabase) GetLastBumpedThreadsPaginated(newsgroup string, threads, offset int) (roots []ArticleEntry) {
	var err error
	var rows *sql.Rows
	if len(newsgroup) > 0 {
		rows, err = self.conn.Query(self.stmt[GetLastBumpedThreadsPaginated_1], newsgroup, threads+offset)
	} else {
		rows, err = self.conn.Query(self.stmt[GetLastBumpedThreadsPaginated_2], threads+offset)
	}

	if err == nil {
		for rows.Next() {
			var ent ArticleEntry
			rows.Scan(&ent[0], &ent[1])
			if offset > 0 {
				offset--
			} else {
				roots = append(roots, ent)
			}
		}
		rows.Close()
	} else {
		log.Println("failed to get last bumped", err)
	}
	return
}

func (self *PostgresDatabase) GroupHasPosts(group string) bool {

	var count int64
	err := self.conn.QueryRow("SELECT 1 FROM ArticlePosts WHERE newsgroup = $1", group).Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	}
	return count > 0
}

// check if a newsgroup exists
func (self *PostgresDatabase) HasNewsgroup(group string) bool {
	var count int64
	err := self.conn.QueryRow(self.stmt[HasNewsgroup], group).Scan(&count)
	return err != sql.ErrNoRows && count > 0
}

// check if an article exists
func (self *PostgresDatabase) HasArticle(message_id string) bool {
	var count int64
	err := self.conn.QueryRow(self.stmt[HasArticle], message_id).Scan(&count)
	return err != sql.ErrNoRows && count > 0
}

// check if an article exists locally
func (self *PostgresDatabase) HasArticleLocal(message_id string) bool {
	var count int64
	err := self.conn.QueryRow(self.stmt[HasArticleLocal], message_id).Scan(&count)
	return err != sql.ErrNoRows && count > 0
}

// count articles we have
func (self *PostgresDatabase) ArticleCount() (count int64) {

	err := self.conn.QueryRow("SELECT COUNT(message_id) FROM ArticlePosts").Scan(&count)
	if err != nil {
		log.Println("failed to count articles", err)
	}
	return
}

// register a new newsgroup
func (self *PostgresDatabase) RegisterNewsgroup(group string) {
	_, err := self.conn.Exec("INSERT INTO Newsgroups (name, last_post) VALUES($1, $2)", group, timeNow())
	if err != nil {
		log.Println("failed to register newsgroup", group, err)
	}
}

func (self *PostgresDatabase) GetThreadAttachments(rootmsgid string) (atts []string, err error) {
	var rows *sql.Rows
	rows, err = self.conn.Query(self.stmt[GetThreadAttachments], rootmsgid)
	if err == nil {
		for rows.Next() {
			var msgid string
			rows.Scan(&msgid)
			atts = append(atts, msgid)
		}
		rows.Close()
	} else if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (self *PostgresDatabase) GetPostAttachments(messageID string) (atts []string) {
	rows, err := self.conn.Query(self.stmt[GetPostAttachments], messageID)
	if err == nil {
		for rows.Next() {
			var val string
			rows.Scan(&val)
			atts = append(atts, val)
		}
		rows.Close()
	} else {
		log.Println("cannot find attachments for", messageID, err)
	}
	return
}

func (self *PostgresDatabase) GetPostAttachmentModels(prefix, messageID string) (atts []AttachmentModel) {
	rows, err := self.conn.Query(self.stmt[GetPostAttachmentModels], messageID)
	if err == nil {
		for rows.Next() {
			var fpath, fname string
			rows.Scan(&fpath, &fname)
			atts = append(atts, &attachment{
				prefix: prefix,
				Path:   fpath,
				Name:   fname,
			})
		}
		rows.Close()
	} else {
		log.Println("failed to get attachment models for", messageID, err)
	}
	return
}

// register a message with the database
func (self *PostgresDatabase) RegisterArticle(message NNTPMessage) (err error) {

	msgid := message.MessageID()
	group := message.Newsgroup()

	if !self.HasNewsgroup(group) {
		self.RegisterNewsgroup(group)
	}
	if self.HasArticle(msgid) {
		return
	}
	now := timeNow()
	// insert article metadata
	_, err = self.conn.Exec(self.stmt[RegisterArticle_1], msgid, HashMessageID(msgid), group, now, message.Reference())
	if err != nil {
		log.Println("failed to insert article metadata", err)
		return
	}
	// update newsgroup
	_, err = self.conn.Exec(self.stmt[RegisterArticle_2], now, group)
	if err != nil {
		log.Println("failed to update newsgroup last post", err)
		return
	}
	// insert article post
	_, err = self.conn.Exec(self.stmt[RegisterArticle_3], group, msgid, message.Reference(), message.Name(), message.Subject(), message.Path(), message.Posted(), message.Message(), message.Addr(), message.FrontendPubkey())
	if err != nil {
		log.Println("cannot insert article post", err)
		return
	}

	// set / update thread state
	if message.OP() {
		// insert new thread for op
		_, err = self.conn.Exec(self.stmt[RegisterArticle_4], message.MessageID(), message.Posted(), group)

		if err != nil {
			log.Println("cannot register thread", msgid, err)
			return
		}
	} else {
		ref := message.Reference()
		postedAt := message.Posted()
		var other int64
		err = self.conn.QueryRow(self.stmt[RegisterArticle_GetLastBump], ref).Scan(&other)
		if err == nil && other > postedAt {
			postedAt = other
		}
		now := timeNow()
		if postedAt > now {
			postedAt = now
		}
		if !message.Sage() {
			// TODO: this could be 1 query possibly?
			var posts int64
			err = self.conn.QueryRow(self.stmt[RegisterArticle_5], ref).Scan(&posts)
			if err == nil && posts <= BumpLimit {
				// bump it nigguh
				_, err = self.conn.Exec(self.stmt[RegisterArticle_6], ref, postedAt)
			}
			if err != nil {
				log.Println("failed to bump thread", ref, err)
				return
			}
		}
		// update last posted
		_, err = self.conn.Exec(self.stmt[RegisterArticle_7], ref, postedAt)
		if err != nil {
			log.Println("failed to update post time for", ref, err)
			return
		}
	}

	var tx *sql.Tx
	tx, err = self.conn.Begin()
	if err == nil {
		var st *sql.Stmt
		st, err = tx.Prepare(pq.CopyIn("nntpheaders", "header_name", "header_value", "header_article_message_id"))
		if err != nil {
			log.Printf("error with copyin: %s", err)
		}
		// register article header key value pairs
		for k, val := range message.Headers() {
			k = strings.ToLower(k)
			for _, v := range val {
				_, err = st.Exec(k, v, msgid)
				if err != nil {
					log.Println("failed to register nntp article header in transaction", err)
					break
				}
			}
		}
		_, err = st.Exec()
		if err == nil {
			st.Close()
			err = tx.Commit()
			if err != nil {
				log.Println("failed to commit nntp article header values:", err)
				return
			}
		} else {
			log.Println("failed to execute prepared statement for nntp article header values:", err)
		}
	}
	err = self.registerNNTPNumber(group, msgid)
	if err != nil {
		log.Println("failed to register nntp number for", msgid, err)
		return
	}
	// register all attachments
	atts := message.Attachments()
	if atts == nil {
		// no attachments
		return
	}
	for _, att := range atts {
		h := hex.EncodeToString(att.Hash())
		_, err = self.conn.Exec(self.stmt[RegisterArticle_8], msgid, h, att.Filename(), att.Filepath())
		if err != nil {
			log.Println("failed to register attachment", err)
			continue
		}
	}
	return
}

//
// get message ids of articles with this header name and value
//
func (self *PostgresDatabase) GetMessageIDByHeader(name, val string) (msgids []string, err error) {
	var rows *sql.Rows
	name = strings.ToLower(name)
	rows, err = self.conn.Query(self.stmt[GetMessageIDByHeader], name, val)
	if err == nil {
		for rows.Next() {
			var msgid string
			rows.Scan(&msgid)
			msgids = append(msgids, msgid)
		}
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) RegisterSigned(message_id, pubkey string) (err error) {
	_, err = self.conn.Exec(self.stmt[RegisterSigned], message_id, pubkey)
	return
}

// get all articles in a newsgroup
// send result down a channel
func (self *PostgresDatabase) GetAllArticlesInGroup(group string, recv chan ArticleEntry) {
	rows, err := self.conn.Query(self.stmt[GetAllArticlesInGroup], group)
	if err != nil {
		log.Printf("failed to get all articles in %s: %s", group, err)
		return
	}
	for rows.Next() {
		var msgid string
		rows.Scan(&msgid)
		recv <- ArticleEntry{msgid, group}
	}
	rows.Close()
}

// get all articles
// send result down a channel
func (self *PostgresDatabase) GetAllArticles() (articles []ArticleEntry) {
	rows, err := self.conn.Query(self.stmt[GetAllArticles])
	if err == nil {
		for rows.Next() {
			var entry ArticleEntry
			rows.Scan(&entry[0], &entry[1])
			articles = append(articles, entry)
		}
		rows.Close()
	} else {
		log.Println("failed to get all articles", err)
	}
	return articles
}

func (self *PostgresDatabase) GetPagesPerBoard(group string) (int, error) {
	//XXX: hardcoded
	return 10, nil
}

func (self *PostgresDatabase) GetThreadsPerPage(group string) (int, error) {
	//XXX: hardcoded
	return 10, nil
}

func (self *PostgresDatabase) GetMessageIDByHash(hash string) (article ArticleEntry, err error) {
	err = self.conn.QueryRow(self.stmt[GetMessageIDByHash], hash).Scan(&article[0], &article[1])
	return
}

func (self *PostgresDatabase) BanAddr(addr string) (err error) {
	_, err = self.conn.Exec("INSERT INTO IPBans(addr, made, expires) VALUES($1, $2, $3)", addr, timeNow(), -1)
	return
}

// assumes it is there
func (self *PostgresDatabase) UnbanAddr(addr string) (err error) {
	_, err = self.conn.Exec("DELETE FROM IPBans WHERE addr >>= $1", addr)
	return
}

func (self *PostgresDatabase) CheckEncIPBanned(encaddr string) (banned bool, err error) {
	var result int64
	err = self.conn.QueryRow(self.stmt[CheckEncIPBanned], encaddr).Scan(&result)
	if err == sql.ErrNoRows {
		err = nil
	}
	banned = result > 0
	return
}

func (self *PostgresDatabase) BanEncAddr(encaddr string) (err error) {
	_, err = self.conn.Exec("INSERT INTO EncIPBans(encaddr, made, expires) VALUES($1, $2, $3)", encaddr, timeNow(), -1)
	return
}

func (self *PostgresDatabase) GetLastAndFirstForGroup(group string) (last, first int64, err error) {
	var rows *sql.Rows
	rows, err = self.conn.Query(self.stmt[GetFirstAndLastForGroup], group)
	if err == nil {
		if rows.Next() {
			err = rows.Scan(&first)
			if err == nil {
				if rows.Next() {
					err = rows.Scan(&last)
				}
			}
		}
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) GetMessageIDForNNTPID(group string, id int64) (msgid string, err error) {
	err = self.conn.QueryRow(self.stmt[GetMessageIDForNNTPID], group, id).Scan(&msgid)
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (self *PostgresDatabase) GetNNTPIDForMessageID(group, msgid string) (id int64, err error) {
	err = self.conn.QueryRow(self.stmt[GetNNTPIDForMessageID], group, msgid).Scan(&id)
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (self *PostgresDatabase) MarkModPubkeyCanModGroup(pubkey, group string) (err error) {
	_, err = self.conn.Exec("INSERT INTO ModPrivs(pubkey, newsgroup, permission) VALUES($1, $2, $3)", pubkey, group, "all")
	return
}

func (self *PostgresDatabase) UnMarkModPubkeyCanModGroup(pubkey, group string) (err error) {
	_, err = self.conn.Exec("DELETE FROM ModPrivs WHERE pubkey = $1 AND newsgroup = $2", pubkey, group)
	return
}

func (self *PostgresDatabase) IsExpired(root_message_id string) bool {
	var count int
	err := self.conn.QueryRow(self.stmt[IsExpired], root_message_id).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		log.Println("error checking for expired article:", err)
	}
	return count == 0
}

func (self *PostgresDatabase) GetLastDaysPostsForGroup(newsgroup string, n int64) (posts []PostEntry) {

	day := time.Hour * 24
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for n > 0 {
		var num int64
		err := self.conn.QueryRow(self.stmt[GetLastDaysPostsForGroup], now.Add(day).Unix(), now.Unix(), newsgroup).Scan(&num)
		if err == nil {
			posts = append(posts, PostEntry{now.Unix(), num})
			now = now.Add(-day)
		} else {
			log.Println("error counting last n days posts", err)
			return nil
		}
		n--
	}
	return
}

func (self *PostgresDatabase) GetLastDaysPosts(n int64) (posts []PostEntry) {

	day := time.Hour * 24
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for n > 0 {
		var num int64
		err := self.conn.QueryRow(self.stmt[GetLastDaysPosts], now.Add(day).Unix(), now.Unix()).Scan(&num)
		if err == nil {
			posts = append(posts, PostEntry{now.Unix(), num})
			now = now.Add(-day)
		} else {
			log.Println("error counting last n days posts", err)
			return nil
		}
		n--
	}
	return
}

func (self *PostgresDatabase) GetLastPostedPostModels(prefix string, n int64) (posts []PostModel) {

	rows, err := self.conn.Query(self.stmt[GetLastPostedPostModels], n)
	if err == nil {
		for rows.Next() {
			model := new(post)
			rows.Scan(&model.board, &model.Message_id, &model.Parent, &model.PostName, &model.PostSubject, &model.MessagePath, &model.Posted, &model.PostMessage, &model.addr)
			model.op = len(model.Parent) == 0
			if len(model.Parent) == 0 {
				model.Parent = model.Message_id
			}
			model.sage = isSage(model.PostSubject)
			atts := self.GetPostAttachmentModels(prefix, model.Message_id)
			if atts != nil {
				model.Files = append(model.Files, atts...)
			}
			// quiet fail
			self.conn.QueryRow(self.stmt[GetArticlePubkey], model.Message_id).Scan(&model.Key)
			posts = append(posts, model)
		}
		rows.Close()
		return
	} else {
		log.Println("failed to prepare query for geting last post models", err)
		return nil
	}
}

func (self *PostgresDatabase) GetMonthlyPostHistory() (posts []PostEntry) {
	rows, err := self.conn.Query(self.stmt[GetYearlyPostHistory])
	if rows != nil {
		for rows.Next() {
			var begin, end, mag int64
			rows.Scan(&begin, &end, &mag)
			posts = append(posts, PostEntry{begin, mag})
		}
		rows.Close()
	}
	if err != nil {
		log.Println("failed getting monthly post history", err)
	}
	return
}

func (self *PostgresDatabase) CheckNNTPLogin(username, passwd string) (valid bool, err error) {
	var login_hash, login_salt string
	err = self.conn.QueryRow(self.stmt[CheckNNTPLogin], username).Scan(&login_hash, &login_salt)
	if err == nil {
		// no errors
		if len(login_hash) > 0 && len(login_salt) > 0 {
			valid = nntpLoginCredHash(passwd, login_salt) == login_hash
		}
	} else if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (self *PostgresDatabase) AddNNTPLogin(username, passwd string) (err error) {
	login_salt := genLoginCredSalt()
	login_hash := nntpLoginCredHash(passwd, login_salt)
	_, err = self.conn.Exec("INSERT INTO NNTPUsers(username, login_hash, login_salt) VALUES($1, $2, $3)", username, login_hash, login_salt)
	return
}

func (self *PostgresDatabase) RemoveNNTPLogin(username string) (err error) {
	_, err = self.conn.Exec("DELETE FROM NNTPUsers WHERE username = $1", username)
	return
}

func (self *PostgresDatabase) CheckNNTPUserExists(username string) (exists bool, err error) {
	var count int64
	err = self.conn.QueryRow(self.stmt[CheckNNTPUserExists], username).Scan(&count)
	exists = count > 0
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (self *PostgresDatabase) GetHeadersForMessage(msgid string) (hdr ArticleHeaders, err error) {
	var rows *sql.Rows
	rows, err = self.conn.Query(self.stmt[GetHeadersForMessage], msgid)
	if err == nil {
		hdr = make(ArticleHeaders)
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			hdr.Add(k, v)
		}
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) CountAllArticlesInGroup(group string) (count int64, err error) {
	err = self.conn.QueryRow(self.stmt[CountAllArticlesInGroup], group).Scan(&count)
	return
}

func (self *PostgresDatabase) GetMessageIDByCIDR(cidr *net.IPNet) (msgids []string, err error) {
	var rows *sql.Rows
	rows, err = self.conn.Query(self.stmt[GetMessageIDByCIDR], cidr.String())
	for err == nil && rows.Next() {
		var msgid string
		err = rows.Scan(&msgid)
		if err == nil {
			msgids = append(msgids, msgid)
		}
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) GetMessageIDByEncryptedIP(encaddr string) (msgids []string, err error) {
	var rows *sql.Rows
	rows, err = self.conn.Query(self.stmt[GetMessageIDByEncryptedIP], encaddr)
	for err == nil && rows.Next() {
		var msgid string
		err = rows.Scan(&msgid)
		if err == nil {
			msgids = append(msgids, msgid)
		}
	}
	if rows != nil {
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) WhitelistPubkey(pubkey string) (err error) {
	_, err = self.conn.Exec("INSERT INTO nntpchan_pubkeys VALUES ('whitelist', $1)", pubkey)
	return
}

func (self *PostgresDatabase) DeletePubkey(pubkey string) (err error) {
	_, err = self.conn.Exec("DELETE FROM nntpchan_pubkeys WHERE pubkey = $1", pubkey)
	return
}

func (self *PostgresDatabase) BlacklistPubkey(pubkey string) (err error) {
	_, err = self.conn.Exec("INSERT INTO nntpchan_pubkeys VALUES ('blacklist', $1)", pubkey)
	return
}

// return true if we should drop this message with this frontend pubkey
func (self *PostgresDatabase) PubkeyRejected(pubkey string) (bool, error) {
	var num int64
	var drop bool
	var err error
	err = self.conn.QueryRow("SELECT COUNT(pubkey) FROM nntpchan_pubkeys WHERE pubkey = $1 AND status = 'whitelist'", pubkey).Scan(&num)
	if err == nil && num == 0 {
		err = self.conn.QueryRow("SELECT COUNT(pubkey) FROM nntpchan_pubkeys WHERE pubkey = $1 and status = 'blacklist'", pubkey).Scan(&num)
		drop = num > 0
	}
	return drop, err
}

func (self *PostgresDatabase) GetPostsBefore(t time.Time) (msgids []string, err error) {
	var rows *sql.Rows
	rows, err = self.conn.Query(self.stmt[GetPostsBefore], t.Unix())
	if err == nil {
		for rows.Next() {
			var msgid string
			rows.Scan(&msgid)
			msgids = append(msgids, msgid)
		}
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) GetPostingStats(gran, begin, end int64) (st PostingStats, err error) {
	return
}

func (self *PostgresDatabase) SearchQuery(prefix, group string, text string, chnl chan PostModel, limit int) (err error) {
	if text != "" && strings.Count(text, "%") == 0 {
		text = "%" + text + "%"
		var rows *sql.Rows
		if group == "" {
			rows, err = self.conn.Query(self.stmt[SearchQuery_1], text, limit)
		} else {
			rows, err = self.conn.Query(self.stmt[SearchQuery_2], group, text, limit)
		}
		if err == nil {
			for rows.Next() {
				p := new(post)
				rows.Scan(&p.board, &p.Message_id, &p.Parent)
				chnl <- p
			}
			rows.Close()
		}
	}
	close(chnl)
	return
}
func (self *PostgresDatabase) SearchByHash(prefix, group, text string, chnl chan PostModel, limit int) (err error) {
	if text != "" && strings.Count(text, "%") == 0 {
		text = "%" + text + "%"
		var rows *sql.Rows
		if group == "" {
			rows, err = self.conn.Query(self.stmt[SearchByHash_1], text, limit)
		} else {

			rows, err = self.conn.Query(self.stmt[SearchByHash_2], text, group, limit)
		}
		if err == nil {
			for rows.Next() {
				p := new(post)
				rows.Scan(&p.board, &p.Message_id, &p.Parent)
				chnl <- p
			}
			rows.Close()
		}
	}
	close(chnl)
	return
}

func (self *PostgresDatabase) GetNewsgroupList() (list NewsgroupList, err error) {
	var rows *sql.Rows
	rows, err = self.conn.Query(self.stmt[GetNewsgroupList])
	if err == nil {
		for rows.Next() {
			var l NewsgroupListEntry
			var lo, hi int64
			rows.Scan(&l[0], &lo, &hi)
			l[1] = fmt.Sprintf("%d", lo)
			l[2] = fmt.Sprintf("%d", hi)
			list = append(list, l)
		}
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) FindCitesInText(text string) (msgids []string, err error) {
	hashes := findBacklinks(text)
	if len(hashes) > 0 {
		q := "SELECT message_id FROM Articles WHERE "
		var params []string
		var qparams []interface{}
		for idx := range hashes {
			params = append(params, fmt.Sprintf(" message_id_hash ILIKE $%d", idx+1))
			qparams = append(qparams, strings.Trim(hashes[idx][2:], " ")+"%")
		}
		q += strings.Join(params, " OR ")
		var rows *sql.Rows
		rows, err = self.conn.Query(q, qparams...)
		if err == sql.ErrNoRows {
			err = nil
		} else if err == nil {
			for rows.Next() {
				var msgid string
				rows.Scan(&msgid)
				msgids = append(msgids, msgid)
			}
		}
	}
	return
}

func (self *PostgresDatabase) GetUkkoPageCount(perpage int) (count int64, err error) {
	err = self.conn.QueryRow(self.stmt[CountUkko]).Scan(&count)
	count /= int64(perpage)
	return
}

func (self *PostgresDatabase) GetNewsgroupStats() (stats []NewsgroupStats, err error) {
	var rows *sql.Rows
	rows, err = self.conn.Query(self.stmt[GetNewsgroupStats])
	if err == nil {
		for rows.Next() {
			var s NewsgroupStats
			rows.Scan(&s.PPD, &s.Name)
			stats = append(stats, s)
		}
		rows.Close()
	}
	return
}

func (self *PostgresDatabase) FindHeaders(group, headername string, lo, hi int64) (hdr ArticleHeaders, err error) {
	hdr = make(ArticleHeaders)
	q := "SELECT header_value FROM nntpheaders WHERE header_name = $1 AND header_article_message_id IN ( SELECT message_id FROM articleposts WHERE newsgroup = $2 )"
	var rows *sql.Rows
	rows, err = self.conn.Query(q, strings.ToLower(headername), group)
	if err == nil {
		for rows.Next() {
			var str string
			rows.Scan(&str)
			hdr.Add(headername, str)
		}
		rows.Close()
	} else if err == sql.ErrNoRows {
		err = nil
	}
	return
}
