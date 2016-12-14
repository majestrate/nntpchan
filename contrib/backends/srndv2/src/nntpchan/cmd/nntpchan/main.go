package main

import (
	log "github.com/Sirupsen/logrus"
	"nntpchan/lib/config"
	"nntpchan/lib/nntp"
	"nntpchan/lib/store"
	"nntpchan/lib/webhooks"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	go func() {
		err := http.ListenAndServe("127.0.0.1:7700", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
	log.Info("starting up nntpchan...")
	cfg_fname := "nntpchan.json"
	conf, err := config.Ensure(cfg_fname)
	if err != nil {
		log.Fatal(err)
	}

	if conf.Log == "debug" {
		log.SetLevel(log.DebugLevel)
	}

	sconfig := conf.Store

	if sconfig == nil {
		log.Fatal("no article storage configured")
	}

	nconfig := conf.NNTP

	if nconfig == nil {
		log.Fatal("no nntp server configured")
	}

	dconfig := conf.Database

	if dconfig == nil {
		log.Fatal("no database configured")
	}

	// create nntp server
	nserv := nntp.NewServer()
	nserv.Config = nconfig
	nserv.Feeds = conf.Feeds

	if nconfig.LoginsFile != "" {
		nserv.Auth = nntp.FlatfileAuth(nconfig.LoginsFile)
	}

	// create article storage
	nserv.Storage, err = store.NewFilesytemStorage(sconfig.Path, true)
	if err != nil {
		log.Fatal(err)
	}

	if conf.WebHooks != nil && len(conf.WebHooks) > 0 {
		// put webhooks into nntp server event hooks
		nserv.Hooks = webhooks.NewWebhooks(conf.WebHooks, nserv.Storage)
	}

	if conf.NNTPHooks != nil && len(conf.NNTPHooks) > 0 {
		var hooks nntp.MulitHook
		if nserv.Hooks != nil {
			hooks = append(hooks, nserv.Hooks)
		}
		for _, h := range conf.NNTPHooks {
			hooks = append(hooks, nntp.NewHook(h))
		}
		nserv.Hooks = hooks
	}

	// nntp server loop
	go func() {
		for {
			naddr := conf.NNTP.Bind
			log.Infof("Bind nntp server to %s", naddr)
			nl, err := net.Listen("tcp", naddr)
			if err == nil {
				err = nserv.Serve(nl)
				if err != nil {
					nl.Close()
					log.Errorf("nntpserver.serve() %s", err.Error())
				}
			} else {
				log.Errorf("nntp server net.Listen failed: %s", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()

	// start persisting feeds
	go nserv.PersistFeeds()

	// handle signals
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl, syscall.SIGHUP)
	for {
		s := <-sigchnl
		if s == syscall.SIGHUP {
			// handle SIGHUP
			conf, err := config.Ensure(cfg_fname)
			if err == nil {
				log.Infof("reloading config: %s", cfg_fname)
				nserv.ReloadServer(conf.NNTP)
				nserv.ReloadFeeds(conf.Feeds)
			}
		}
	}
}
