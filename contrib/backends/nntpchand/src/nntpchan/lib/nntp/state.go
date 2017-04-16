package nntp

// state of an nntp connection
type ConnState struct {
	// name of parent feed
	FeedName string `json:"feedname"`
	// name of the connection
	ConnName string `json:"connname"`
	// hostname of remote connection
	HostName string `json:"hostname"`
	// current nntp mode
	Mode Mode `json:"mode"`
	// current selected nntp newsgroup
	Group Newsgroup `json:"newsgroup"`
	// current selected nntp article
	Article string `json:"article"`
	// parent feed's policy
	Policy *FeedPolicy `json:"feedpolicy"`
	// is this connection open?
	Open bool `json:"open"`
}
