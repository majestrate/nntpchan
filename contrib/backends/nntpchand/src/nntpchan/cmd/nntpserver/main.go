package main

// simple nntp server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/majestrate/srndv2/lib/config"
	"github.com/majestrate/srndv2/lib/nntp"
	"github.com/majestrate/srndv2/lib/store"
	"net"
)

func main() {

	log.Info("starting NNTP server...")
	conf, err := config.Ensure("settings.json")
	if err != nil {
		log.Fatal(err)
	}

	if conf.Log == "debug" {
		log.SetLevel(log.DebugLevel)
	}

	serv := &nntp.Server{
		Config: conf.NNTP,
		Feeds:  conf.Feeds,
	}
	serv.Storage, err = store.NewFilesytemStorage(conf.Store.Path, false)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.Listen("tcp", conf.NNTP.Bind)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("listening on ", l.Addr())
	err = serv.Serve(l)
	if err != nil {
		log.Fatal(err)
	}
}
