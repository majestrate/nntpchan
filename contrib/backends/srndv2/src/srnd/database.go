//
// database.go
//
package srnd

import (
	"log"
	"net"
	"strings"
	"time"
)

// a ( MessageID , newsgroup ) tuple
type ArticleEntry [2]string

func (self ArticleEntry) Newsgroup() string {
	return self[1]
}

func (self ArticleEntry) MessageID() string {
	return self[0]
}

// a (messageID , parent messageID) tuple
type MessageIDTuple [2]string

func (self MessageIDTuple) MessageID() string {
	return strings.Trim(self[0], " ")
}

func (self MessageIDTuple) Reference() string {
	r := strings.Trim(self[1], " ")
	if len(r) == 0 {
		return self.MessageID()
	}
	return r
}

// a ( time point, magnitude ) tuple
type PostEntry [2]int64

func (self PostEntry) Time() time.Time {
	return time.Unix(self[0], 0)
}

func (self PostEntry) Count() int64 {
	return self[1]
}

type PostEntryList []PostEntry

// stats about newsgroup postings
type NewsgroupStats struct {
	PPD  int64
	Name string
}

type PostingStatsEntry struct {
	Groups []NewsgroupStats
}

type PostingStats struct {
	History []PostingStatsEntry
}

// newsgroup, first, last
type NewsgroupListEntry [3]string

type NewsgroupList []NewsgroupListEntry

type Database interface {
	Close()
	CreateTables()
	HasNewsgroup(group string) bool
	HasArticle(message_id string) bool
	HasArticleLocal(message_id string) bool
	RegisterNewsgroup(group string)
	RegisterArticle(article NNTPMessage) error
	GetAllArticlesInGroup(group string, send chan ArticleEntry)
	CountAllArticlesInGroup(group string) (int64, error)
	GetAllArticles() []ArticleEntry

	SetConnectionLifetime(seconds int)
	SetMaxOpenConns(n int)
	SetMaxIdleConns(n int)

	// check if a newsgroup is banned
	NewsgroupBanned(group string) (bool, error)

	// ban / unban newsgroup
	BanNewsgroup(group string) error
	UnbanNewsgroup(group string) error

	// delete an entire newsgroup
	// delete from the article store too
	NukeNewsgroup(group string, store ArticleStore)

	// return true if this is root post has expired
	IsExpired(root_message_id string) bool

	// get an article's MessageID given the hash of the MessageID
	// return an article entry or nil when it doesn't exist + and error if it happened
	GetMessageIDByHash(hash string) (ArticleEntry, error)

	// get root message_id, newsgroup, pageno for a post regardless if it's rootpost or not
	GetInfoForMessage(msgid string) (string, string, int64, error)

	// what page is the thread with this root post on?
	// return newsgroup, pageno
	GetPageForRootMessage(root_message_id string) (string, int64, error)

	// record that a message given a message id was posted signed by this pubkey
	RegisterSigned(message_id, pubkey string) error

	// get the number of articles we have
	ArticleCount() int64

	// return true if this thread has any replies
	ThreadHasReplies(root_message_id string) bool

	// get the number of posts in a certain newsgroup since N seconds ago
	// if N <= 0 then count all we have now
	CountPostsInGroup(group string, time_frame int64) int64

	// get the stats for the overview page
	GetNewsgroupStats() ([]NewsgroupStats, error)

	// get all replies to a thread
	// if last > 0 then get that many of the last replies
	// start at reply number start
	GetThreadReplies(root_message_id string, start, last int) []string

	// count the number of replies to this thread
	CountThreadReplies(root_message_id string) int64

	// get all attachments for this message
	GetPostAttachments(message_id string) []string

	// get all attachments in a thread
	GetThreadAttachments(rootmsgid string) ([]string, error)

	// get all attachments for this message
	GetPostAttachmentModels(prefix, message_id string) []AttachmentModel

	// return true if this newsgroup has posts
	GroupHasPosts(newsgroup string) bool

	// get all active threads on a board
	// send each thread's ArticleEntry down a channel
	GetGroupThreads(newsgroup string, send chan ArticleEntry)

	// get every message id for root posts that need to be expired in a newsgroup
	// threadcount is the upperbound limit to how many root posts we keep
	GetRootPostsForExpiration(newsgroup string, threadcount int) []string

	// get the number of pages a board has
	GetGroupPageCount(newsgroup string) int64

	// get board page number N
	// prefix and frontend are injected
	// does not load replies for thread, only gets root posts
	GetGroupForPage(prefix, frontend, newsgroup string, pageno, perpage int) BoardModel

	// get the root posts of the last N bumped threads in a given newsgroup or "" for ukko
	GetLastBumpedThreads(newsgroup string, threadcount int) []ArticleEntry

	// get root posts of last N bumped threads with pagination offset
	GetLastBumpedThreadsPaginated(newsgroup string, threadcount, offset int) []ArticleEntry

	// get the PostModels for replies to a thread
	// prefix is injected into the post models
	GetThreadReplyPostModels(prefix, rootMessageID string, start, limit int) []PostModel

	// get a post model for a post
	// prefix is injected into the post model
	GetPostModel(prefix, messageID string) PostModel

	// add a public key to the database
	AddModPubkey(pubkey string) error

	// mark that a mod with this pubkey can act on all boards
	MarkModPubkeyGlobal(pubkey string) error

	// revoke mod with this pubkey the privilege of being able to act on all boards
	UnMarkModPubkeyGlobal(pubkey string) error

	// check if this mod pubkey can moderate at a global level
	CheckModPubkeyGlobal(pubkey string) bool

	// check if a mod with this pubkey has permission to moderate at all
	CheckModPubkey(pubkey string) bool

	// check if a pubkey has admin privs
	CheckAdminPubkey(pubkey string) (bool, error)

	// mark a key as having admin privs
	MarkPubkeyAdmin(pubkey string) error

	// unmark a key as having admin privs
	UnmarkPubkeyAdmin(pubkey string) error

	// check if a mod with this pubkey can moderate on the given newsgroup
	CheckModPubkeyCanModGroup(pubkey, newsgroup string) bool

	// add a pubkey to be able to mod a newsgroup
	MarkModPubkeyCanModGroup(pubkey, newsgroup string) error

	// remote a pubkey to they can't mod a newsgroup
	UnMarkModPubkeyCanModGroup(pubkey, newsgroup string) error

	// ban an article
	BanArticle(messageID, reason string) error

	// check if an article is banned or not
	ArticleBanned(messageID string) bool

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
	DeleteArticle(msg_id string) error

	// remove an article from the database
	RemoveArticle(msg_id string) error

	// detele the existance of a thread from the threads table, does NOT remove replies
	DeleteThread(root_msg_id string) error

	// get threads per page for a newsgroup
	GetThreadsPerPage(group string) (int, error)

	// get pages per board for a newsgroup
	GetPagesPerBoard(group string) (int, error)

	// get every newsgroup we know of
	GetAllNewsgroups() []string

	// get all post models in a newsgroup
	// ordered from oldest to newest
	GetPostsInGroup(group string) ([]PostModel, error)

	// get the numerical id of the last , first article for a given group
	GetLastAndFirstForGroup(group string) (int64, int64, error)

	// get a message id give a newsgroup and the nntp id
	GetMessageIDForNNTPID(group string, id int64) (string, error)

	// get nntp id for a given message-id
	GetNNTPIDForMessageID(group, msgid string) (int64, error)

	// get the last N days post count in decending order
	GetLastDaysPosts(n int64) []PostEntry

	// get the last N days post count in decending order
	GetLastDaysPostsForGroup(newsgroup string, n int64) []PostEntry

	// get post history per month since beginning of time
	GetMonthlyPostHistory() []PostEntry

	// get the last N posts that were made globally
	GetLastPostedPostModels(prefix string, n int64) []PostModel

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

	// get the headers for a message given its message-id
	GetHeadersForMessage(msgid string) (ArticleHeaders, error)

	// get all message-ids posted by posters in this cidr
	GetMessageIDByCIDR(cidr *net.IPNet) ([]string, error)

	// get all message-ids posted by poster with encrypted ip
	GetMessageIDByEncryptedIP(encaddr string) ([]string, error)

	// check if this public key is banned from posting
	PubkeyRejected(pubkey string) (bool, error)

	// ban a public key from posting
	BlacklistPubkey(pubkey string) error
	WhitelistPubkey(pubkey string) error
	DeletePubkey(pubkey string) error

	// get all message-id posted before a time
	GetPostsBefore(t time.Time) ([]string, error)

	// get statistics about posting in a time slice
	GetPostingStats(granularity, begin, end int64) (PostingStats, error)

	// peform search query
	SearchQuery(prefix, group, text string, chnl chan PostModel, limit int) error

	// find posts with similar hash
	SearchByHash(prefix, group, posthash string, chnl chan PostModel, limit int) error

	// get full thread model
	GetThreadModel(prefix, root_msgid string) (ThreadModel, error)

	// get post models with nntp id in a newsgroup
	GetNNTPPostsInGroup(newsgroup string) ([]PostModel, error)

	// get post message-id where hash is similar to string
	GetCitesByPostHashLike(like string) ([]MessageIDTuple, error)

	// get newsgroup list with watermarks
	GetNewsgroupList() (NewsgroupList, error)

	// find cites in text
	FindCitesInText(msg string) ([]string, error)

	// find headers in group with lo/hi watermark and list of patterns
	FindHeaders(group, headername string, lo, hi int64) (ArticleHeaders, error)

	// count ukko pages
	GetUkkoPageCount(perpage int) (int64, error)
}

func NewDatabase(db_type, schema, host, port, user, password string) Database {
	if db_type == "postgres" {
		if schema == "srnd" {
			return NewPostgresDatabase(host, port, user, password)
		}
	}
	log.Fatalf("invalid database type: %s/%s", db_type, schema)
	return nil
}
