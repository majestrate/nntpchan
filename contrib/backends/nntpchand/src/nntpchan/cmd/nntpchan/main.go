package nntpchan

import (
	log "github.com/Sirupsen/logrus"
	"net"
	_ "net/http/pprof"
	"nntpchan/lib/config"
	"nntpchan/lib/database"
	"nntpchan/lib/frontend"
	"nntpchan/lib/nntp"
	"nntpchan/lib/store"
	"nntpchan/lib/webhooks"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type runStatus struct {
	nntpListener net.Listener
	run          bool
	done         chan error
}

func (st *runStatus) Stop() {
	st.run = false
	if st.nntpListener != nil {
		st.nntpListener.Close()
	}
	st.nntpListener = nil
	log.Info("stopping daemon process")
}

func Main() {
	st := &runStatus{
		run:  true,
		done: make(chan error),
	}
	log.Info("starting up nntpchan...")
	cfgFname := "nntpchan.json"
	conf, err := config.Ensure(cfgFname)
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

	var db database.Database
	for _, fconf := range conf.Frontends {
		var f frontend.Frontend
		f, err = frontend.NewHTTPFrontend(fconf, db)
		if err == nil {
			go f.Serve()
		}
	}

	// start persisting feeds
	go nserv.PersistFeeds()

	// handle signals
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl, syscall.SIGHUP, os.Interrupt)
	go func() {
		for {
			s := <-sigchnl
			if s == syscall.SIGHUP {
				// handle SIGHUP
				conf, err := config.Ensure(cfgFname)
				if err == nil {
					log.Infof("reloading config: %s", cfgFname)
					nserv.ReloadServer(conf.NNTP)
					nserv.ReloadFeeds(conf.Feeds)
				} else {
					log.Errorf("failed to reload config: %s", err)
				}
			} else if s == os.Interrupt {
				// handle interrupted, clean close
				st.Stop()
				return
			}
		}
	}()
	go func() {
		var err error
		for st.run {
			var nl net.Listener
			naddr := conf.NNTP.Bind
			log.Infof("Bind nntp server to %s", naddr)
			nl, err = net.Listen("tcp", naddr)
			if err == nil {
				st.nntpListener = nl
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
		st.done <- err
	}()
	e := <-st.done
	if e != nil {
		log.Fatal(e)
	}
	log.Info("ended")
}
