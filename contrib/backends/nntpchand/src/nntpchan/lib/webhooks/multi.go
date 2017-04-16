package webhooks

import (
	"nntpchan/lib/nntp"
)

// webhook multiplexer
type multiWebhook struct {
	hooks []Webhook
}

// got an article
func (m *multiWebhook) GotArticle(msgid nntp.MessageID, group nntp.Newsgroup) {
	for _, h := range m.hooks {
		h.GotArticle(msgid, group)
	}
}

func (m *multiWebhook) SentArticleVia(msgid nntp.MessageID, feedname string) {
	for _, h := range m.hooks {
		h.SentArticleVia(msgid, feedname)
	}
}
