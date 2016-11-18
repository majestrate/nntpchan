package nntp

// an nntp connection
type Conn interface {

	// negotiate an nntp session on this connection
	// returns nil if we negitated successfully
	// returns ErrAuthRejected if the remote server rejected any authentication
	// we sent or another error if one occured while negotiating
	Negotiate(stream bool) error

	// obtain connection state
	GetState() *ConnState

	// retutrn true if posting is allowed
	// return false if posting is not allowed
	PostingAllowed() bool

	// handle inbound non-streaming connection
	// call event hooks on event
	ProcessInbound(hooks EventHooks)

	// does this connection want to do nntp streaming?
	WantsStreaming() bool

	// what mode are we in?
	// returns mode in all caps
	Mode() Mode

	// initiate nntp streaming
	// after calling this the caller MUST call StreamAndQuit()
	// returns a channel for message ids, true if caller sends on the channel or
	// returns nil and ErrStreamingNotAllowed if streaming is not allowed on this
	// connection or another error if one occurs while trying to start streaming
	StartStreaming() (chan ArticleEntry, error)

	// stream articles and quit when the channel obtained by StartStreaming() is
	// closed, after which this nntp connection is no longer open
	StreamAndQuit()

	// is this nntp connection open?
	IsOpen() bool

	// send quit command and close connection
	Quit()

	// download all articles in a newsgroup
	// returns error if a network error occurs
	DownloadGroup(g Newsgroup) error

	// get list of active newsgroups
	ListNewsgroups() ([]Newsgroup, error)
}
