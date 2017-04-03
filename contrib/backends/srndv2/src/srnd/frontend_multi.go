//
// frontend_multi.go
// frontend multiplexer
//

package srnd

// muxed frontend for holding many frontends
type multiFrontend struct {
	muxedpostchan chan frontendPost
	frontends     []Frontend
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

	// poll for incoming
	chnl := self.PostsChan()
	for {
		select {
		case nntp := <-chnl:
			for _, frontend := range self.frontends {
				if frontend.AllowNewsgroup(nntp.Newsgroup()) {
					ch := frontend.PostsChan()
					ch <- nntp
				}
			}
			break
		}
	}
}

func (self multiFrontend) PostsChan() chan frontendPost {
	return self.muxedpostchan
}

func MuxFrontends(fronts ...Frontend) Frontend {
	var front multiFrontend
	front.muxedpostchan = make(chan frontendPost, 64)
	front.frontends = fronts
	return front
}
