package nntp

// callback hooks fired on certain events
type EventHooks interface {
	// called when we have obtained an article given its message-id
	GotArticle(msgid MessageID, group Newsgroup)
	// called when we have sent an article to a single remote feed
	SentArticleVia(msgid MessageID, feedname string)
}
