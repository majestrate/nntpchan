package nntp

// multiplexed event hook
type MulitHook []EventHooks

func (m MulitHook) GotArticle(msgid MessageID, group Newsgroup) {
	for _, h := range m {
		h.GotArticle(msgid, group)
	}
}

func (m MulitHook) SentArticleVia(msgid MessageID, feedname string) {
	for _, h := range m {
		h.SentArticleVia(msgid, feedname)
	}
}
