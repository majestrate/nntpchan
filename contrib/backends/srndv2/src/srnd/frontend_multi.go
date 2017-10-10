//
// frontend_multi.go
// frontend multiplexer
//

package srnd

// muxed frontend for holding many frontends
type multiFrontend struct {
	frontends []Frontend
}

func (self multiFrontend) GetCacheHandler() CacheHandler {
	// TODO: fixme :^)
	return nil
}

func (self multiFrontend) AllowNewsgroup(newsgroup string) bool {
	return true
}

func (self multiFrontend) Regen(msg ArticleEntry) {
	for _, front := range self.frontends {
		front.Regen(msg)
	}
}

func (self multiFrontend) Mainloop() {
	for idx := range self.frontends {
		go self.frontends[idx].Mainloop()
	}
}

func (self multiFrontend) HandleNewPost(nntp frontendPost) {
	for idx := range self.frontends {
		self.frontends[idx].HandleNewPost(nntp)
	}
}

func (self multiFrontend) RegenOnModEvent(newsgroup, msgid, root string, page int) {
	for idx := range self.frontends {
		self.frontends[idx].RegenOnModEvent(newsgroup, msgid, root, page)
	}
}

func MuxFrontends(fronts ...Frontend) Frontend {
	var front multiFrontend
	front.frontends = fronts
	return front
}
