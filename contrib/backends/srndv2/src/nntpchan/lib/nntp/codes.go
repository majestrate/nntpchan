package nntp

// 1xx codes

// help info follows
const RPL_Help = "100"

// capabilities info follows
const RPL_Capabilities = "101"

// server date time follows
const RPL_Date = "111"

// 2xx codes

// posting is allowed
const RPL_PostingAllowed = "200"

// posting is not allowed
const RPL_PostingNotAllowed = "201"

// streaming mode enabled
const RPL_PostingStreaming = "203"

// reply to QUIT command, we will close the connection
const RPL_Quit = "205"

// reply for GROUP and LISTGROUP commands
const RPL_Group = "211"

// info list follows
const RPL_List = "215"

// index follows
const RPL_Index = "218"

// article follows
const RPL_Article = "220"

// article headers follows
const RPL_ArticleHeaders = "221"

// article body follows
const RPL_ArticleBody = "222"

// selected article exists
const RPL_ArticleSelectedExists = "223"

// overview info follows
const RPL_Overview = "224"

// list of article heards follows
const RPL_HeadersList = "225"

// list of new articles follows
const RPL_NewArticles = "230"

// list of newsgroups followes
const RPL_NewsgroupList = "231"

// article was transfered okay by IHAVE command
const RPL_TransferOkay = "235"

// article is not found by CHECK and we want it
const RPL_StreamingAccept = "238"

// article was transfered via TAKETHIS successfully
const RPL_StreamingTransfered = "239"

// article was transfered by POST command successfully
const RPL_PostReceived = "240"

// AUTHINFO SIMPLE accepted
const RPL_AuthInfoAccepted = "250"

// authentication creds have been accepted
const RPL_AuthAccepted = "281"

// binary content follows
const RPL_Binary = "288"

// line sent for posting allowed
const Line_PostingAllowed = RPL_PostingAllowed + " Posting Allowed"

// line sent for posting not allowed
const Line_PostingNotAllowed = RPL_PostingNotAllowed + " Posting Not Allowed"

// 3xx codes

// article is accepted via IHAVE
const RPL_TransferAccepted = "335"

// article was accepted via POST
const RPL_PostAccepted = "340"

// continue with authorization
const RPL_ContinueAuthorization = "350"

// more authentication info required
const RPL_MoreAuth = "381"

// continue with tls handshake
const RPL_TLSContinue = "382"

// 4xx codes

// server says servive is not avaiable on initial connection
const RPL_NotAvaiable = "400"

// server is in the wrong mode
const RPL_WrongMode = "401"

// generic fault prevent action from being taken
const RPL_GenericError = "403"

// newsgroup does not exist
const RPL_NoSuchGroup = "411"

// no newsgroup has been selected
const RPL_NoGroupSelected = "412"

// no tin style index available
const RPL_NoIndex = "418"

// current article number is invalid
const RPL_NoArticleNum = "420"

// no next article in this group (NEXT)
const RPL_NoNextArticle = "421"

// no previous article in this group (LAST)
const RPL_NoPrevArticle = "422"

// no article in specified range
const RPL_NoArticleRange = "423"

// no article with that message-id
const RPL_NoArticleMsgID = "430"

// defer article asked by CHECK comamnd
const RPL_StreamingDefer = "431"

// article is not wanted (1st stage of IHAVE)
const RPL_TransferNotWanted = "435"

// article was not sent defer sending (either stage of IHAVE)
const RPL_TransferDefer = "436"

// reject transfer do not retry (2nd stage IHAVE)
const RPL_TransferReject = "437"

// reject article and don't ask again (CHECK command)
const RPL_StreamingReject = "438"

// article transfer via streaming failed (TAKETHIS)
const RPL_StreamingFailed = "439"

// posting not permitted (1st stage of POST command)
const RPL_PostingNotPermitted = "440"

// posting failed (2nd stage of POST command)
const RPL_PostingFailed = "441"

// authorization required
const RPL_AuthorizeRequired = "450"

// authorization rejected
const RPL_AuthorizeRejected = "452"

// command unavaibale until client has authenticated
const RPL_AuthenticateRequired = "480"

// authentication creds rejected
const RPL_AuthenticateRejected = "482"

// command unavailable until connection is encrypted
const RPL_EncryptionRequired = "483"

// 5xx codes

// got an unknown command
const RPL_UnknownCommand = "500"

// got a command with invalid syntax
const RPL_SyntaxError = "501"

// fatal error happened and connection will close
const RPL_GenericFatal = "502"

// feature is not supported
const RPL_FeatureNotSupported = "503"

// message encoding is bad
const RPL_EncodingError = "504"

// starttls can not be done
const RPL_TLSRejected = "580"

// line sent on invalid mode
const Line_InvalidMode = RPL_SyntaxError + " Invalid Mode Selected"

// line sent on successful streaming
const Line_StreamingAllowed = RPL_PostingStreaming + " aw yeh streamit brah"

// send this when we handle a QUIT command
const Line_RPLQuit = RPL_Quit + " bai"
