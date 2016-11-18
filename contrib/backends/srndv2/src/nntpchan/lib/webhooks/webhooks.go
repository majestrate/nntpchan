package webhooks

import (
	"nntpchan/lib/config"
	"nntpchan/lib/nntp"
	"nntpchan/lib/nntp/message"
	"nntpchan/lib/store"
)

type Webhook interface {
	// implements nntp.EventHooks
	nntp.EventHooks
}

// create webhook multiplexing multiple web hooks
func NewWebhooks(conf []*config.WebhookConfig, st store.Storage) Webhook {
	h := message.NewHeaderIO()
	var hooks []Webhook
	for _, c := range conf {
		hooks = append(hooks, &httpWebhook{
			conf:    c,
			storage: st,
			hdr:     h,
		})
	}

	return &multiWebhook{
		hooks: hooks,
	}
}
