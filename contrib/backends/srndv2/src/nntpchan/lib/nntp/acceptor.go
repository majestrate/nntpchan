package nntp

import (
	"nntpchan/lib/nntp/message"
)

const (
	// accepted article
	ARTICLE_ACCEPT = iota
	// reject article, don't send again
	ARTICLE_REJECT
	// defer article, send later
	ARTICLE_DEFER
	// reject + ban
	ARTICLE_BAN
)

type PolicyStatus int

const PolicyAccept = PolicyStatus(ARTICLE_ACCEPT)
const PolicyReject = PolicyStatus(ARTICLE_REJECT)
const PolicyDefer = PolicyStatus(ARTICLE_DEFER)
const PolicyBan = PolicyStatus(ARTICLE_BAN)

func (s PolicyStatus) String() string {
	switch int(s) {
	case ARTICLE_ACCEPT:
		return "ACCEPTED"
	case ARTICLE_REJECT:
		return "REJECTED"
	case ARTICLE_DEFER:
		return "DEFERRED"
	case ARTICLE_BAN:
		return "BANNED"
	default:
		return "[invalid policy status]"
	}
}

// is this an accept code?
func (s PolicyStatus) Accept() bool {
	return s == ARTICLE_ACCEPT
}

// is this a defer code?
func (s PolicyStatus) Defer() bool {
	return s == ARTICLE_DEFER
}

// is this a ban code
func (s PolicyStatus) Ban() bool {
	return s == ARTICLE_BAN
}

// is this a reject code?
func (s PolicyStatus) Reject() bool {
	return s == ARTICLE_BAN || s == ARTICLE_REJECT
}

// type defining a policy that determines if we want to accept/reject/defer an
// incoming article
type ArticleAcceptor interface {
	// check article given an article header
	CheckHeader(hdr message.Header) PolicyStatus
	// check article given a message id
	CheckMessageID(msgid MessageID) PolicyStatus
	// get max article size in bytes
	MaxArticleSize() int64
}
