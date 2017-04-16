package nntp

import (
	"fmt"
	"strings"
)

// an nntp stream event
// these are pipelined between nntp servers
type StreamEvent string

func (ev StreamEvent) MessageID() MessageID {
	parts := strings.Split(string(ev), " ")
	if len(parts) > 1 {
		return MessageID(parts[1])
	}
	return ""
}

func (ev StreamEvent) String() string {
	return string(ev)
}

func (ev StreamEvent) Command() string {
	return strings.Split(ev.String(), " ")[0]
}

func (ev StreamEvent) Valid() bool {
	return strings.Count(ev.String(), " ") == 1 && ev.MessageID().Valid()
}

var stream_TAKETHIS = "TAKETHIS"
var stream_CHECK = "CHECK"

func createStreamEvent(cmd string, msgid MessageID) StreamEvent {
	if msgid.Valid() {
		return StreamEvent(fmt.Sprintf("%s %s", cmd, msgid))
	} else {
		return ""
	}
}

func stream_rpl_Accept(msgid MessageID) StreamEvent {
	return createStreamEvent(RPL_StreamingAccept, msgid)
}

func stream_rpl_Reject(msgid MessageID) StreamEvent {
	return createStreamEvent(RPL_StreamingReject, msgid)
}

func stream_rpl_Defer(msgid MessageID) StreamEvent {
	return createStreamEvent(RPL_StreamingDefer, msgid)
}

func stream_rpl_Failed(msgid MessageID) StreamEvent {
	return createStreamEvent(RPL_StreamingFailed, msgid)
}

func stream_cmd_TAKETHIS(msgid MessageID) StreamEvent {
	return createStreamEvent(stream_TAKETHIS, msgid)
}

func stream_cmd_CHECK(msgid MessageID) StreamEvent {
	return createStreamEvent(stream_CHECK, msgid)
}
