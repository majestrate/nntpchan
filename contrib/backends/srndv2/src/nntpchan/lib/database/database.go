package database

import (
	"errors"
	"nntpchan/lib/config"
	"nntpchan/lib/model"
	"net"
	"strings"
)

// generic database driver
type DB interface {
	// finalize all transactions and close connection
	// after calling this db driver can no longer be used
	Close()
	// ensire database is well formed
	Ensure() error
	// do we have a newsgroup locally?
	HasNewsgroup(group string) (bool, error)
	// have we seen an article with message-id before?
	SeenArticle(message_id string) (bool, error)
	// do we have an article locally given message-id?
	HasArticle(message_id string) (bool, error)
	// register a newsgroup with database
	RegisterNewsgroup(group string) error
	// register an article
	RegisterArticle(a *model.Article) error
	// get all articles in a newsgroup
	// send entries down a channel
	// return error if one happens while fetching
	GetAllArticlesInGroup(group string, send chan model.ArticleEntry) error
	// count all the articles in a newsgroup
	CountAllArticlesInGroup(group string) (int64, error)
	// get all articles locally known
	GetAllArticles() ([]model.ArticleEntry, error)

	// check if a newsgroup is banned
	NewsgroupBanned(group string) (bool, error)

	// ban newsgroup
	BanNewsgroup(group string) error
	// unban newsgroup
	UnbanNewsgroup(group string) error

	// return true if this is root post has expired
	IsExpired(root_message_id string) (bool, error)

	// get an article's MessageID given the hash of the MessageID
	// return an article entry or nil when it doesn't exist + and error if it happened
	GetMessageIDByHash(hash string) (model.ArticleEntry, error)

	// get root message_id, newsgroup, pageno for a post regardless if it's rootpost or not
	GetInfoForMessage(msgid string) (string, string, int64, error)

	// what page is the thread with this root post on?
	// return newsgroup, pageno
	GetPageForRootMessage(root_message_id string) (string, int64, error)

	// record that a message given a message id was posted signed by this pubkey
	RegisterSigned(message_id, pubkey string) error

	// get the number of articles we have in all groups
	ArticleCount() (int64, error)

	// return true if a thread with given root post with message-id has any replies
	ThreadHasReplies(root_message_id string) (bool, error)

	// get the number of posts in a certain newsgroup since N seconds ago
	// if N <= 0 then count all we have now
	CountPostsInGroup(group string, time_frame int64) (int64, error)

	// get all replies' message-id to a thread
	// if last > 0 then get that many of the last replies
	// start at reply number start
	GetThreadReplies(root_message_id string, start, last int) ([]string, error)

	// get a thread model given root post's message id
	GetThread(root_message_id string) (model.Thread, error)
	// get a thread model given root post hash
	GetThreadByHash(hash string) (model.Thread, error)

	// count the number of replies to this thread
	CountThreadReplies(root_message_id string) (int64, error)

	// get all attachments for a message given its message-id
	GetPostAttachments(message_id string) ([]*model.Attachment, error)

	// return true if this newsgroup has posts
	GroupHasPosts(newsgroup string) (bool, error)

	// get all active threads on a board
	// send each thread's ArticleEntry down a channel
	// return error if one happens while fetching
	GetGroupThreads(newsgroup string, send chan model.ArticleEntry) error

	// get every message id for root posts that need to be expired in a newsgroup
	// threadcount is the upperbound limit to how many root posts we keep
	GetRootPostsForExpiration(newsgroup string, threadcount int) ([]string, error)

	// get the number of pages a board has
	GetGroupPageCount(newsgroup string) (int64, error)

	// get board page number N
	// fully loads all models
	GetGroupForPage(newsgroup string, pageno, perpage int) (*model.BoardPage, error)

	// get the threads for ukko page
	GetUkkoThreads(threadcount int) ([]*model.Thread, error)

	// get a post model for a single post
	GetPost(messageID string) (*model.Post, error)

	// add a public key to the database
	AddModPubkey(pubkey string) error

	// mark that a mod with this pubkey can act on all boards
	MarkModPubkeyGlobal(pubkey string) error

	// revoke mod with this pubkey the privilege of being able to act on all boards
	UnMarkModPubkeyGlobal(pubkey string) error

	// check if this mod pubkey can moderate at a global level
	CheckModPubkeyGlobal(pubkey string) bool

	// check if a mod with this pubkey has permission to moderate at all
	CheckModPubkey(pubkey string) (bool, error)

	// check if a mod with this pubkey can moderate on the given newsgroup
	CheckModPubkeyCanModGroup(pubkey, newsgroup string) (bool, error)

	// add a pubkey to be able to mod a newsgroup
	MarkModPubkeyCanModGroup(pubkey, newsgroup string) error

	// remote a pubkey to they can't mod a newsgroup
	UnMarkModPubkeyCanModGroup(pubkey, newsgroup string) error

	// ban an article
	BanArticle(messageID, reason string) error

	// check if an article is banned or not
	ArticleBanned(messageID string) (bool, error)

	// Get ip address given the encrypted version
	// return emtpy string if we don't have it
	GetIPAddress(encAddr string) (string, error)

	// check if an ip is banned from our local
	CheckIPBanned(addr string) (bool, error)

	// check if an encrypted ip is banned from our local
	CheckEncIPBanned(encAddr string) (bool, error)

	// ban an ip address from the local
	BanAddr(addr string) error

	// unban an ip address from the local
	UnbanAddr(addr string) error

	// ban an encrypted ip address from the remote
	BanEncAddr(encAddr string) error

	// return the encrypted version of an IPAddress
	// if it's not already there insert it into the database
	GetEncAddress(addr string) (string, error)

	// get the decryption key for an encrypted address
	// return empty string if we don't have it
	GetEncKey(encAddr string) (string, error)

	// delete an article from the database
	// if the article is a root post then all replies are also deleted
	DeleteArticle(msg_id string) error

	// forget that we tracked an article given the messgae-id
	ForgetArticle(msg_id string) error

	// get threads per page for a newsgroup
	GetThreadsPerPage(group string) (int, error)

	// get pages per board for a newsgroup
	GetPagesPerBoard(group string) (int, error)

	// get every newsgroup we current carry
	GetAllNewsgroups() ([]string, error)

	// get the numerical id of the last , first article for a given group
	GetLastAndFirstForGroup(group string) (int64, int64, error)

	// get a message id give a newsgroup and the nntp id
	GetMessageIDForNNTPID(group string, id int64) (string, error)

	// get nntp id for a given message-id
	GetNNTPIDForMessageID(group, msgid string) (int64, error)

	// get the last N days post count in decending order
	GetLastDaysPosts(n int64) ([]model.PostEntry, error)

	// get the last N days post count in decending order
	GetLastDaysPostsForGroup(newsgroup string, n int64) ([]model.PostEntry, error)

	// get post history per month since beginning of time
	GetMonthlyPostHistory() ([]model.PostEntry, error)

	// check if an nntp login cred is correct
	CheckNNTPLogin(username, passwd string) (bool, error)

	// add an nntp login credential
	AddNNTPLogin(username, passwd string) error

	// remove an nntp login credential
	RemoveNNTPLogin(username string) error

	// check if an nntp login credential given a user exists
	CheckNNTPUserExists(username string) (bool, error)

	// get the message ids of an article that has this header with the given value
	GetMessageIDByHeader(name, value string) ([]string, error)

	// get the header for a message given its message-id
	GetHeadersForMessage(msgid string) (model.ArticleHeader, error)

	// get all message-ids posted by posters in this cidr
	GetMessageIDByCIDR(cidr *net.IPNet) ([]string, error)

	// get all message-ids posted by poster with encrypted ip
	GetMessageIDByEncryptedIP(encaddr string) ([]string, error)
}

// type for webhooks db backend
type WebhookDB interface {

	// mark article sent
	MarkMessageSent(msgid, feedname string) error

	// check if an article was sent
	CheckMessageSent(msgid, feedname string) (bool, error)
}

// get new database connector from configuration
func NewDBFromConfig(c *config.DatabaseConfig) (db DB, err error) {
	dbtype := strings.ToLower(c.Type)
	if dbtype == "postgres" {
		err = errors.New("postgres not supported")
	} else {
		err = errors.New("no such database driver: " + c.Type)
	}
	return
}

func NewWebhooksDBFromConfig(c *config.DatabaseConfig) (db WebhookDB, err error) {
	dbtype := strings.ToLower(c.Type)
	if dbtype == "postgres" {
		err = errors.New("postgres not supported")
	} else {
		err = errors.New("no such database driver: " + c.Type)
	}
	return
}
