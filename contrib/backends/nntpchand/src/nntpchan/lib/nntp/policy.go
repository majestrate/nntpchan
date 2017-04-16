package nntp

//
// a policy that governs whether we federate an article via a feed
//
type FeedPolicy struct {
	// list of whitelist regexps for newsgorups
	Whitelist []string `json:"whitelist"`
	// list of blacklist regexps for newsgroups
	Blacklist []string `json:"blacklist"`
	// are anon posts of any kind allowed?
	AllowAnonPosts bool `json:"anon"`
	// are anon posts with attachments allowed?
	AllowAnonAttachments bool `json:"anon_attachments"`
	// are any attachments allowed?
	AllowAttachments bool `json:"attachments"`
	// do we require Proof Of Work for untrusted connections?
	UntrustedRequiresPoW bool `json:"pow"`
}

// default feed policy to be used if not configured explicitly
var DefaultFeedPolicy = &FeedPolicy{
	Whitelist:            []string{"ctl", "overchan.test"},
	Blacklist:            []string{`!^overchan\.`},
	AllowAnonPosts:       true,
	AllowAnonAttachments: false,
	UntrustedRequiresPoW: true,
	AllowAttachments:     true,
}
