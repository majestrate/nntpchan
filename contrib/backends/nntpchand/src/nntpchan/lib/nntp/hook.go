package nntp

import (
	log "github.com/Sirupsen/logrus"
	"nntpchan/lib/config"
	"os/exec"
)

type Hook struct {
	cfg *config.NNTPHookConfig
}

func NewHook(cfg *config.NNTPHookConfig) *Hook {
	return &Hook{
		cfg: cfg,
	}
}

func (h *Hook) GotArticle(msgid MessageID, group Newsgroup) {
	c := exec.Command(h.cfg.Exec, group.String(), msgid.String())
	log.Infof("calling hook %s", h.cfg.Name)
	err := c.Run()
	if err != nil {
		log.Errorf("error in nntp hook %s: %s", h.cfg.Name, err.Error())
	}
}

func (*Hook) SentArticleVia(msgid MessageID, feedname string) {

}
